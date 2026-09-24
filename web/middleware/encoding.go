package middleware

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Rehtt/Kit/web"
)

type EncodingOption struct {
	// MinSize 响应小于该字节数时跳过压缩；0 走默认 1024。
	MinSize int
	// Level 压缩级别；0 走默认级别，其余取 compress/flate 支持的级别。
	Level int
	// AllowedContentTypes Content-Type 前缀白名单；nil 走默认列表，
	// 空切片表示压缩全部类型。
	AllowedContentTypes []string
}

const (
	defaultMinSize    = 1024
	maxPooledBodySize = 64 << 10
)

var defaultAllowedTypes = []string{
	"text/",
	"application/json",
	"application/javascript",
	"application/xml",
	"application/wasm",
	"image/svg+xml",
}

// Levels are in [-2, 9]. Indexing by level+2 keeps separate pools for each
// configured compression level; level 0 is normalized to DefaultCompression.
var (
	gzipPools  [12]sync.Pool
	zlibPools  [12]sync.Pool
	writerPool = sync.Pool{New: func() any {
		return &encodingWriter{buf: make([]byte, 0, defaultMinSize+128)}
	}}
)

// Encoding 根据 Accept-Encoding 启用 gzip / deflate。
// 缓冲到 MinSize 后再决策；未指定 Content-Type 时至少缓冲 512 字节用于嗅探。
// 响应结束或显式 Flush 时使用已有数据；HEAD、协议升级与已编码响应直接放行。
func Encoding(opts ...EncodingOption) web.HandlerFunc {
	opt := EncodingOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.MinSize <= 0 {
		opt.MinSize = defaultMinSize
	}
	opt.Level = normalizeLevel(opt.Level)
	var allowed []string
	if opt.AllowedContentTypes != nil {
		allowed = append([]string(nil), opt.AllowedContentTypes...)
	} else {
		allowed = defaultAllowedTypes
	}

	return func(c *web.Context) {
		// Headers already sent cannot advertise a new content coding.
		if responseWritten(c.Writer) {
			c.Next()
			return
		}
		addVary(c.Writer.Header(), "Accept-Encoding")
		req := c.Request
		if req.Method == http.MethodHead || req.Header.Get("Upgrade") != "" || c.Writer.Header().Get("Content-Encoding") != "" {
			c.Next()
			return
		}

		algo := negotiate(req.Header.Get("Accept-Encoding"))
		if algo == "" {
			c.Next()
			return
		}

		original := c.Writer
		ew := acquireWriter(original, algo, opt.Level, opt.MinSize, allowed)
		c.Writer = ew
		completed := false
		defer func() {
			if completed {
				ew.finish()
			} else {
				// A panic must not flush a response that was only buffered.
				ew.abort()
			}
			c.Writer = original
			releaseWriter(ew)
		}()

		c.Next()
		completed = true
	}
}

// Follow wrappers that expose Unwrap, including nested encoding writers whose
// buffered headers have not necessarily reached the underlying response yet.
func responseWritten(w http.ResponseWriter) bool {
	for w != nil {
		if state, ok := w.(interface{ Written() bool }); ok && state.Written() {
			return true
		}
		unwrapper, ok := w.(web.ResponseWriterUnwrapper)
		if !ok {
			return false
		}
		w = unwrapper.Unwrap()
	}
	return false
}

func normalizeLevel(level int) int {
	if level == 0 {
		return flate.DefaultCompression
	}
	if level < flate.HuffmanOnly || level > flate.BestCompression {
		panic("[web] middleware.Encoding: invalid compression level")
	}
	return level
}

// negotiate returns the supported coding with the highest q-value. Explicit
// q=0 excludes a coding even when a wildcard is present; gzip wins ties.
func negotiate(header string) string {
	if header == "" {
		return ""
	}
	var gzipQ, deflateQ, wildcardQ float64
	var gzipSet, deflateSet, wildcardSet bool
	for rest := header; rest != ""; {
		raw, next, _ := strings.Cut(rest, ",")
		rest = next
		namePart, parameters, _ := strings.Cut(raw, ";")
		name := strings.TrimSpace(namePart)
		if name == "" {
			continue
		}
		q := 1.0
		for parameters != "" {
			parameter, nextParameter, _ := strings.Cut(parameters, ";")
			parameters = nextParameter
			key, value, ok := strings.Cut(parameter, "=")
			if !ok || !strings.EqualFold(strings.TrimSpace(key), "q") {
				continue
			}
			parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			if err != nil || parsed < 0 || parsed > 1 {
				q = 0
			} else {
				q = parsed
			}
		}
		switch {
		case name == "*":
			wildcardQ = q
			wildcardSet = true
		case strings.EqualFold(name, "gzip"):
			gzipQ = q
			gzipSet = true
		case strings.EqualFold(name, "deflate"):
			deflateQ = q
			deflateSet = true
		}
	}
	quality := func(name string) float64 {
		switch name {
		case "gzip":
			if gzipSet {
				return gzipQ
			}
		case "deflate":
			if deflateSet {
				return deflateQ
			}
		}
		// Preserve the original wildcard >= 0 check: NaN wildcards are ignored,
		// while explicit coding weights keep their existing NaN behavior.
		if wildcardSet && wildcardQ >= 0 {
			return wildcardQ
		}
		return 0
	}
	gzipQuality, deflateQuality := quality("gzip"), quality("deflate")
	if gzipQuality <= 0 && deflateQuality <= 0 {
		return ""
	}
	if gzipQuality >= deflateQuality {
		return "gzip"
	}
	return "deflate"
}

func addVary(header http.Header, value string) {
	for _, line := range header.Values("Vary") {
		for _, item := range strings.Split(line, ",") {
			item = strings.TrimSpace(item)
			if item == "*" || strings.EqualFold(item, value) {
				return
			}
		}
	}
	header.Add("Vary", value)
}

// encodingWriter buffers until the response is known to be compressible.
type encodingWriter struct {
	http.ResponseWriter
	algo    string
	level   int
	minSize int
	allowed []string

	buf       []byte
	started   bool // final status is fixed, even while body data is buffered
	committed bool
	compress  bool
	encoder   io.WriteCloser

	deferredStatus int
	hasDeferred    bool
}

func acquireWriter(rw http.ResponseWriter, algo string, level, minSize int, allowed []string) *encodingWriter {
	w := writerPool.Get().(*encodingWriter)
	w.ResponseWriter = rw
	w.algo = algo
	w.level = level
	w.minSize = minSize
	w.allowed = allowed
	w.buf = w.buf[:0]
	w.started = false
	w.committed = false
	w.compress = false
	w.encoder = nil
	w.deferredStatus = 0
	w.hasDeferred = false
	return w
}

func releaseWriter(w *encodingWriter) {
	w.closeEncoder()
	w.ResponseWriter = nil
	w.algo = ""
	w.level = 0
	w.minSize = 0
	w.allowed = nil
	w.deferredStatus = 0
	w.hasDeferred = false
	w.committed = false
	w.started = false
	w.compress = false
	if cap(w.buf) > maxPooledBodySize {
		w.buf = nil
	} else {
		w.buf = w.buf[:0]
	}
	writerPool.Put(w)
}

func (w *encodingWriter) WriteHeader(code int) {
	if w.committed || w.started {
		return
	}
	if code < 100 || code > 999 {
		panic(fmt.Sprintf("invalid WriteHeader code %v", code))
	}
	if code >= 100 && code < 200 && code != http.StatusSwitchingProtocols {
		// Informational headers are sent immediately and do not commit the
		// buffered final response.
		w.ResponseWriter.WriteHeader(code)
		return
	}
	if code == http.StatusSwitchingProtocols {
		w.ResponseWriter.WriteHeader(code)
		w.started = true
		w.committed = true
		return
	}
	w.started = true
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
	w.started = true

	threshold := w.minSize
	if w.Header().Get("Content-Type") == "" && threshold < 512 {
		threshold = 512
	}
	if len(w.buf) < threshold && len(b) > 0 {
		need := threshold - len(w.buf)
		if len(b) > need {
			w.buf = append(w.buf, b[:need]...)
			if err := w.commitIfPending(true); err != nil {
				return need, err
			}
			n, err := w.writeCommitted(b[need:])
			return need + n, err
		}
	}
	w.buf = append(w.buf, b...)
	if len(w.buf) < threshold {
		return len(b), nil
	}
	if err := w.commitIfPending(true); err != nil {
		return len(b), err
	}
	return len(b), nil
}

func (w *encodingWriter) writeCommitted(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	if w.compress {
		return w.encoder.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *encodingWriter) Flush() {
	_ = w.FlushError()
}

func (w *encodingWriter) FlushError() error {
	if !w.committed {
		if err := w.commitIfPending(true); err != nil {
			return err
		}
	}
	if w.compress {
		if f, ok := w.encoder.(interface{ Flush() error }); ok {
			if err := f.Flush(); err != nil {
				return err
			}
		}
	}
	return http.NewResponseController(w.ResponseWriter).Flush()
}

func (w *encodingWriter) finish() {
	if !w.committed {
		_ = w.commitIfPending(false)
	}
	w.closeEncoder()
}

func (w *encodingWriter) abort() {
	w.buf = w.buf[:0]
	if !responseWritten(w.ResponseWriter) {
		// Encoding starts only when Content-Encoding is absent. Any coding
		// introduced since then belongs to the discarded body, including one
		// produced by an inner encoder that has already finished.
		w.Header().Del("Content-Encoding")
	}
	// Discard buffered compressed data and the success trailer before returning
	// the encoder to its pool. A failed response must remain an incomplete stream.
	switch encoder := w.encoder.(type) {
	case *pooledGzip:
		encoder.Reset(io.Discard)
	case *pooledZlib:
		encoder.Reset(io.Discard)
	}
	w.closeEncoder()
}

func (w *encodingWriter) closeEncoder() {
	if w.encoder == nil {
		return
	}
	_ = w.encoder.Close()
	w.encoder = nil
}

func (w *encodingWriter) commitIfPending(forceWrite bool) error {
	return w.commit(w.shouldCompress(forceWrite))
}

func (w *encodingWriter) shouldCompress(forceWrite bool) bool {
	if !forceWrite && len(w.buf) < w.minSize {
		return false
	}
	if w.ResponseWriter.Header().Get("Content-Encoding") != "" {
		return false
	}
	// Ranges describe the original representation; compressing a selected
	// range would make its Content-Range offsets refer to different bytes.
	if w.deferredStatus == http.StatusPartialContent || w.Header().Get("Content-Range") != "" {
		return false
	}
	if w.hasDeferred && !bodyAllowed(w.deferredStatus) {
		return false
	}
	ct := w.ResponseWriter.Header().Get("Content-Type")
	if ct == "" && len(w.buf) > 0 {
		ct = http.DetectContentType(w.buf[:min(len(w.buf), 512)])
		w.ResponseWriter.Header().Set("Content-Type", ct)
	}
	return typeAllowed(ct, w.allowed)
}

func bodyAllowed(status int) bool {
	return status < 100 || status >= 200 && status != http.StatusNoContent && status != http.StatusNotModified
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
	if responseWritten(w.ResponseWriter) || w.ResponseWriter.Header().Get("Content-Encoding") != "" {
		compress = false
	}
	w.committed = true
	w.compress = compress

	if compress {
		w.ResponseWriter.Header().Set("Content-Encoding", w.algo)
		w.ResponseWriter.Header().Del("Content-Length")
		switch w.algo {
		case "gzip":
			gw := acquireGzip(w.level)
			gw.Reset(w.ResponseWriter)
			w.encoder = &pooledGzip{Writer: gw, pool: &gzipPools[w.level+2]}
		case "deflate":
			zw := acquireZlib(w.level)
			zw.Reset(w.ResponseWriter)
			w.encoder = &pooledZlib{Writer: zw, pool: &zlibPools[w.level+2]}
		}
	}
	if w.hasDeferred {
		w.ResponseWriter.WriteHeader(w.deferredStatus)
	}
	if len(w.buf) == 0 {
		return nil
	}
	_, err := w.writeCommitted(w.buf)
	w.buf = w.buf[:0]
	return err
}

func acquireGzip(level int) *gzip.Writer {
	pool := &gzipPools[level+2]
	if value := pool.Get(); value != nil {
		return value.(*gzip.Writer)
	}
	w, _ := gzip.NewWriterLevel(io.Discard, level)
	return w
}

func acquireZlib(level int) *zlib.Writer {
	pool := &zlibPools[level+2]
	if value := pool.Get(); value != nil {
		return value.(*zlib.Writer)
	}
	w, _ := zlib.NewWriterLevel(io.Discard, level)
	return w
}

func (w *encodingWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

type pooledGzip struct {
	*gzip.Writer
	pool *sync.Pool
}

func (p *pooledGzip) Close() error {
	if p.Writer == nil {
		return nil
	}
	err := p.Writer.Close()
	p.pool.Put(p.Writer)
	p.Writer = nil
	return err
}

type pooledZlib struct {
	*zlib.Writer
	pool *sync.Pool
}

func (p *pooledZlib) Close() error {
	if p.Writer == nil {
		return nil
	}
	err := p.Writer.Close()
	p.pool.Put(p.Writer)
	p.Writer = nil
	return err
}
