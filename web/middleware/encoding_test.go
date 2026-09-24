package middleware

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
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

func TestEncodingNegotiatesQualityValues(t *testing.T) {
	g := web.New()
	g.HeadMiddleware(Encoding())
	g.GET("/x", func(ctx *web.Context) {
		ctx.Writer.Header().Set("Content-Type", "text/plain")
		ctx.Writer.Write([]byte(strings.Repeat("a", 4096)))
	})
	for _, tc := range []struct {
		accept, want string
	}{
		{"gzip;q=0, deflate;q=0.5", "deflate"},
		{"gzip;q=0, *;q=1", "deflate"},
		{"deflate;q=0.5, gzip;q=0.5", "gzip"},
	} {
		rec := runReq(t, g, "GET", "/x", tc.accept)
		if got := rec.Header().Get("Content-Encoding"); got != tc.want {
			t.Fatalf("Accept-Encoding %q: got %q want %q", tc.accept, got, tc.want)
		}
	}
}

func TestEncodingReadFromPreservesAllData(t *testing.T) {
	g := web.New()
	g.HeadMiddleware(Encoding(EncodingOption{MinSize: 16}))
	body := strings.Repeat("read-from-", 100)
	g.GET("/x", func(ctx *web.Context) {
		ctx.Writer.Header().Set("Content-Type", "text/plain")
		if _, err := ctx.ReadFrom(strings.NewReader(body)); err != nil {
			t.Errorf("ReadFrom: %v", err)
		}
	})
	rec := runReq(t, g, "GET", "/x", "gzip")
	r, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read gzip: %v", err)
	}
	if string(got) != body {
		t.Fatalf("body mismatch: got %d bytes want %d", len(got), len(body))
	}
}

func TestEncodingDetectsContentTypeBeforeCompression(t *testing.T) {
	g := web.New()
	g.HeadMiddleware(Encoding(EncodingOption{MinSize: 16}))
	g.GET("/x", func(ctx *web.Context) {
		ctx.Writer.Write([]byte(strings.Repeat("plain text ", 100)))
	})
	rec := runReq(t, g, "GET", "/x", "gzip")
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Fatalf("Content-Type=%q", got)
	}
}

func TestEncodingPanicDropsBufferedBody(t *testing.T) {
	g := web.New(web.WithOnPanic(func(ctx *web.Context, _ any) {
		http.Error(ctx.Writer, "panic", http.StatusInternalServerError)
	}))
	g.HeadMiddleware(Encoding(EncodingOption{MinSize: 1024}))
	g.GET("/x", func(ctx *web.Context) {
		ctx.Writer.Header().Set("Content-Type", "text/plain")
		ctx.Writer.Write([]byte("buffered body"))
		panic("boom")
	})
	rec := runReq(t, g, "GET", "/x", "gzip")
	if rec.Code != http.StatusInternalServerError || rec.Body.String() != "panic\n" {
		t.Fatalf("status=%d body=%q", rec.Code, rec.Body.String())
	}
}

func TestEncodingRejectsInvalidLevel(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("invalid level should panic")
		}
	}()
	_ = Encoding(EncodingOption{Level: 10})
}

func TestEncodingSniffSmallThreshold(t *testing.T) {
	for _, kind := range []string{"html", "png"} {
		for _, size := range []int{64, 1024} {
			for _, chunk := range []int{1, 1024} {
				t.Run(fmt.Sprintf("%s/size%d/chunk%d", kind, size, chunk), func(t *testing.T) {
					prefix, contentType, encoding := "<!DOCTYPE html><html>", "text/html; charset=utf-8", "gzip"
					if kind == "png" {
						prefix, contentType, encoding = "\x89PNG\r\n\x1a\n", "image/png", ""
					}
					body := []byte(prefix + strings.Repeat("x", size-len(prefix)))
					rec := httptest.NewRecorder()
					w := acquireWriter(rec, "gzip", -1, 1, defaultAllowedTypes)
					defer releaseWriter(w)
					for offset := 0; offset < len(body); {
						end := min(offset+chunk, len(body))
						n, err := w.Write(body[offset:end])
						if err != nil || n != end-offset {
							t.Fatalf("Write = %d, %v", n, err)
						}
						offset = end
						if offset < 512 && w.committed {
							t.Fatal("committed before sniff buffer filled")
						}
					}
					w.finish()
					if got := rec.Header().Get("Content-Type"); got != contentType {
						t.Fatalf("Content-Type = %q", got)
					}
					assertEncodedBody(t, rec, encoding, body)
				})
			}
		}
	}
}

func assertEncodedBody(t *testing.T, rec *httptest.ResponseRecorder, encoding string, want []byte) {
	t.Helper()
	if got := rec.Header().Get("Content-Encoding"); got != encoding {
		t.Fatalf("Content-Encoding = %q, want %q", got, encoding)
	}
	var reader io.Reader = rec.Body
	if encoding == "gzip" {
		r, err := gzip.NewReader(reader)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()
		reader = r
	}
	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("body mismatch: got %d bytes, want %d", len(got), len(want))
	}
}

func TestEncodingEarlyDecision(t *testing.T) {
	for _, tc := range []struct {
		name, ct string
		minSize  int
		flush    bool
		encoding string
	}{
		{"explicit type", "text/plain", 1, false, "gzip"},
		{"flush sniff", "", 1024, true, "gzip"},
		{"finish below minimum", "", 1024, false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			w := acquireWriter(rec, "gzip", -1, tc.minSize, defaultAllowedTypes)
			defer releaseWriter(w)
			if tc.ct != "" {
				w.Header().Set("Content-Type", tc.ct)
			}
			body := []byte("<!DOCTYPE html><html>short")
			if _, err := w.Write(body); err != nil {
				t.Fatal(err)
			}
			if tc.flush {
				w.Flush()
				if !rec.Flushed {
					t.Fatal("Flush not forwarded")
				}
			}
			if (tc.ct != "" || tc.flush) && !w.committed {
				t.Fatal("expected immediate commit")
			}
			w.finish()
			wantType := tc.ct
			if wantType == "" {
				wantType = "text/html; charset=utf-8"
			}
			if got := rec.Result().Header.Get("Content-Type"); got != wantType {
				t.Fatalf("Content-Type = %q", got)
			}
			assertEncodedBody(t, rec, tc.encoding, body)
		})
	}
}
