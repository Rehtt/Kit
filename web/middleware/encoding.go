package middleware

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/Rehtt/Kit/web"
)

type EncodingOption struct {
	// MinSize 响应小于该字节数时跳过压缩；0 走默认 1024。
	MinSize int
	// Level gzip / zlib 压缩级别；0 走默认。
	Level int
	// AllowedContentTypes Content-Type 前缀白名单；nil 走默认列表，
	// 空切片表示压缩全部类型。
	AllowedContentTypes []string
}

const defaultMinSize = 1024

var defaultAllowedTypes = []string{
	"text/",
	"application/json",
	"application/javascript",
	"application/xml",
	"application/wasm",
	"image/svg+xml",
}

var (
	gzipPool = sync.Pool{New: func() any { return gzip.NewWriter(io.Discard) }}
	zlibPool = sync.Pool{New: func() any { w := zlib.NewWriter(io.Discard); return w }}
)

// Encoding 根据 Accept-Encoding 启用 gzip / deflate。
// 缓冲到 MinSize 后再决策是否压缩，HEAD 与已编码响应直接放行。
func Encoding(opts ...EncodingOption) web.HandlerFunc {
	opt := EncodingOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.MinSize <= 0 {
		opt.MinSize = defaultMinSize
	}
	allowed := opt.AllowedContentTypes
	if allowed == nil {
		allowed = defaultAllowedTypes
	}

	return func(c *web.Context) {
		req := c.Request

		if req.Method == http.MethodHead {
			c.Next()
			return
		}
		if c.Writer.Header().Get("Content-Encoding") != "" {
			c.Next()
			return
		}

		algo := negotiate(req.Header.Get("Accept-Encoding"))
		if algo == "" {
			c.Next()
			return
		}

		// 即使最终未压缩也要写 Vary，避免缓存层错配。
		c.Writer.Header().Add("Vary", "Accept-Encoding")

		original := c.Writer
		ew := acquireWriter(original, algo, opt.Level, opt.MinSize, allowed)
		c.Writer = ew

		defer func() {
			ew.finish()
			c.Writer = original
			releaseWriter(ew)
		}()

		c.Next()
	}
}

// negotiate 按 gzip > deflate 顺序选第一个支持的算法（忽略 q-value）。
func negotiate(header string) string {
	if header == "" {
		return ""
	}
	for _, raw := range strings.Split(header, ",") {
		token := strings.TrimSpace(raw)
		if i := strings.IndexByte(token, ';'); i >= 0 {
			token = strings.TrimSpace(token[:i])
		}
		switch strings.ToLower(token) {
		case "gzip":
			return "gzip"
		case "deflate":
			return "deflate"
		}
	}
	return ""
}

// encodingWriter 在 commit 前缓冲写入，等看清 Content-Type / 是否到 minSize 再决策。
// WriteHeader 会被推迟到 commit，避免在决策前定型 header。
type encodingWriter struct {
	http.ResponseWriter
	algo    string
	level   int
	minSize int
	allowed []string

	buf       []byte
	committed bool
	compress  bool
	encoder   io.WriteCloser

	deferredStatus int
	hasDeferred    bool
}

var writerPool = sync.Pool{New: func() any { return &encodingWriter{buf: make([]byte, 0, defaultMinSize+128)} }}

func acquireWriter(rw http.ResponseWriter, algo string, level, minSize int, allowed []string) *encodingWriter {
	w := writerPool.Get().(*encodingWriter)
	w.ResponseWriter = rw
	w.algo = algo
	w.level = level
	w.minSize = minSize
	w.allowed = allowed
	w.buf = w.buf[:0]
	w.committed = false
	w.compress = false
	w.encoder = nil
	return w
}

func releaseWriter(w *encodingWriter) {
	w.ResponseWriter = nil
	w.encoder = nil
	w.allowed = nil
	w.deferredStatus = 0
	w.hasDeferred = false
	writerPool.Put(w)
}

func (w *encodingWriter) WriteHeader(code int) {
	if w.ResponseWriter.Header().Get("Content-Encoding") != "" {
		// 上游已选择编码，立即放行。
		w.commit(false)
		w.ResponseWriter.WriteHeader(code)
		return
	}
	w.deferredStatus = code
	w.hasDeferred = true
}

func (w *encodingWriter) Write(b []byte) (int, error) {
	if w.committed {
		if w.compress {
			return w.encoder.Write(b)
		}
		return w.ResponseWriter.Write(b)
	}
	w.buf = append(w.buf, b...)
	if len(w.buf) < w.minSize {
		return len(b), nil
	}
	if err := w.commitIfPending(true); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (w *encodingWriter) Flush() {
	// SSE 等流式场景需要尽快 commit。
	if !w.committed {
		_ = w.commitIfPending(true)
	}
	if w.compress {
		if f, ok := w.encoder.(interface{ Flush() error }); ok {
			_ = f.Flush()
		}
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// finish 在中间件返回前调用：未决策时按当前缓冲决策，已压缩时 Close。
func (w *encodingWriter) finish() {
	if !w.committed {
		_ = w.commitIfPending(false)
	}
	if w.compress && w.encoder != nil {
		_ = w.encoder.Close()
	}
}

// forceWrite=true 表示必须立即写出（缓冲到阈值 / Flush）；
// false 表示 finish 阶段，缓冲未到阈值则不压缩。
func (w *encodingWriter) commitIfPending(forceWrite bool) error {
	return w.commit(w.shouldCompress(forceWrite))
}

func (w *encodingWriter) shouldCompress(forceWrite bool) bool {
	if !forceWrite && len(w.buf) < w.minSize {
		return false
	}
	ct := w.ResponseWriter.Header().Get("Content-Type")
	if ct == "" {
		// 复用 stdlib 嗅探规则，仅用于决策不回写 header。
		ct = http.DetectContentType(w.buf)
	}
	if w.ResponseWriter.Header().Get("Content-Encoding") != "" {
		return false
	}
	return typeAllowed(ct, w.allowed)
}

func typeAllowed(ct string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	ct = strings.ToLower(ct)
	for _, p := range allowed {
		if strings.HasPrefix(ct, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

func (w *encodingWriter) commit(compress bool) error {
	if w.committed {
		return nil
	}
	w.committed = true
	w.compress = compress

	if compress {
		w.ResponseWriter.Header().Set("Content-Encoding", w.algo)
		w.ResponseWriter.Header().Del("Content-Length")
		switch w.algo {
		case "gzip":
			gw := gzipPool.Get().(*gzip.Writer)
			gw.Reset(w.ResponseWriter)
			w.encoder = &pooledGzip{Writer: gw}
		case "deflate":
			zw := zlibPool.Get().(*zlib.Writer)
			zw.Reset(w.ResponseWriter)
			w.encoder = &pooledZlib{Writer: zw}
		}
	}

	if w.hasDeferred {
		w.ResponseWriter.WriteHeader(w.deferredStatus)
	}

	if len(w.buf) == 0 {
		return nil
	}
	if compress {
		_, err := w.encoder.Write(w.buf)
		w.buf = w.buf[:0]
		return err
	}
	_, err := w.ResponseWriter.Write(w.buf)
	w.buf = w.buf[:0]
	return err
}

func (w *encodingWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

type pooledGzip struct {
	*gzip.Writer
}

func (p *pooledGzip) Close() error {
	err := p.Writer.Close()
	gzipPool.Put(p.Writer)
	p.Writer = nil
	return err
}

type pooledZlib struct {
	*zlib.Writer
}

func (p *pooledZlib) Close() error {
	err := p.Writer.Close()
	zlibPool.Put(p.Writer)
	p.Writer = nil
	return err
}
