package middleware

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Rehtt/Kit/web"
)

func runReq(t *testing.T, g *web.GOweb, method, path, accept string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	if accept != "" {
		req.Header.Set("Accept-Encoding", accept)
	}
	g.ServeHTTP(rec, req)
	return rec
}

// 大于阈值 + 允许的 Content-Type → 启用 gzip。
func TestEncodingGzip(t *testing.T) {
	g := web.New()
	g.HeadMiddleware(Encoding())
	body := strings.Repeat("a", 4096)
	g.GET("/x", func(ctx *web.Context) {
		ctx.Writer.Header().Set("Content-Type", "text/plain")
		ctx.Writer.Write([]byte(body))
	})

	rec := runReq(t, g, "GET", "/x", "gzip")
	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding: %q", got)
	}
	if vary := rec.Header().Get("Vary"); !strings.Contains(vary, "Accept-Encoding") {
		t.Fatalf("Vary 缺失: %q", vary)
	}
	r, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read gzip: %v", err)
	}
	if string(got) != body {
		t.Fatalf("解压不一致")
	}
}

// 小于阈值的响应不压缩。
func TestEncodingBelowMinSize(t *testing.T) {
	g := web.New()
	g.HeadMiddleware(Encoding(EncodingOption{MinSize: 1024}))
	g.GET("/x", func(ctx *web.Context) {
		ctx.Writer.Header().Set("Content-Type", "text/plain")
		ctx.Writer.Write([]byte("tiny"))
	})

	rec := runReq(t, g, "GET", "/x", "gzip")
	if rec.Header().Get("Content-Encoding") != "" {
		t.Fatalf("小响应不应压缩，但 CE=%q", rec.Header().Get("Content-Encoding"))
	}
	if rec.Body.String() != "tiny" {
		t.Fatalf("body 错: %q", rec.Body.String())
	}
}

// 不在白名单的 Content-Type 不压缩。
func TestEncodingContentTypeFiltered(t *testing.T) {
	g := web.New()
	g.HeadMiddleware(Encoding())
	g.GET("/x", func(ctx *web.Context) {
		ctx.Writer.Header().Set("Content-Type", "image/png")
		ctx.Writer.Write(bytes.Repeat([]byte("\x89PNG"), 1024))
	})
	rec := runReq(t, g, "GET", "/x", "gzip")
	if rec.Header().Get("Content-Encoding") != "" {
		t.Fatalf("image/png 不应被压缩")
	}
}

// HEAD 请求不压缩。
func TestEncodingSkipsHead(t *testing.T) {
	g := web.New()
	g.HeadMiddleware(Encoding())
	g.GET("/x", func(ctx *web.Context) {
		ctx.Writer.Header().Set("Content-Type", "text/plain")
		ctx.Writer.Write([]byte(strings.Repeat("a", 4096)))
	})
	rec := runReq(t, g, "HEAD", "/x", "gzip")
	if rec.Header().Get("Content-Encoding") != "" {
		t.Fatalf("HEAD 不应被压缩")
	}
}

// deflate 走 zlib 通路。
func TestEncodingDeflate(t *testing.T) {
	g := web.New()
	g.HeadMiddleware(Encoding())
	body := strings.Repeat("hello ", 1000)
	g.GET("/x", func(ctx *web.Context) {
		ctx.Writer.Header().Set("Content-Type", "text/plain")
		ctx.Writer.Write([]byte(body))
	})

	rec := runReq(t, g, "GET", "/x", "deflate")
	if got := rec.Header().Get("Content-Encoding"); got != "deflate" {
		t.Fatalf("Content-Encoding: %q", got)
	}
	r, err := zlib.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("zlib reader: %v", err)
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != body {
		t.Fatalf("解压不一致")
	}
}

// 关键回归：连续 Write 不应丢数据（修复 defer Close 的 bug）。
func TestEncodingMultipleWrites(t *testing.T) {
	g := web.New()
	g.HeadMiddleware(Encoding(EncodingOption{MinSize: 16}))
	g.GET("/x", func(ctx *web.Context) {
		ctx.Writer.Header().Set("Content-Type", "text/plain")
		for i := 0; i < 100; i++ {
			ctx.Writer.Write([]byte("0123456789"))
		}
	})

	rec := runReq(t, g, "GET", "/x", "gzip")
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("应启用 gzip")
	}
	r, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got) != 1000 {
		t.Fatalf("数据丢失：want 1000B, got %dB", len(got))
	}
}

// Accept-Encoding 不支持时透传。
func TestEncodingNoAcceptHeader(t *testing.T) {
	g := web.New()
	g.HeadMiddleware(Encoding())
	g.GET("/x", func(ctx *web.Context) {
		ctx.Writer.Header().Set("Content-Type", "text/plain")
		ctx.Writer.Write([]byte("plain"))
	})
	rec := runReq(t, g, "GET", "/x", "")
	if rec.Header().Get("Content-Encoding") != "" {
		t.Fatalf("无 Accept-Encoding 不应压缩")
	}
	if rec.Body.String() != "plain" {
		t.Fatalf("got %q", rec.Body.String())
	}
}

// 状态码经 WriteHeader 后能正确传到底层。
func TestEncodingPreservesStatusCode(t *testing.T) {
	g := web.New()
	g.HeadMiddleware(Encoding())
	g.GET("/x", func(ctx *web.Context) {
		ctx.Writer.Header().Set("Content-Type", "text/plain")
		ctx.Writer.WriteHeader(http.StatusAccepted)
		ctx.Writer.Write([]byte(strings.Repeat("a", 4096)))
	})
	rec := runReq(t, g, "GET", "/x", "gzip")
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status: %d", rec.Code)
	}
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("应启用 gzip")
	}
}
