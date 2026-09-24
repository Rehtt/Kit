package middleware

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/Rehtt/Kit/web"
)

type encodingUnwrapWriter struct{ http.ResponseWriter }

func (w encodingUnwrapWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func TestEncodingSkipsCommittedResponse(t *testing.T) {
	body := strings.Repeat("original data\n", 256)
	for _, coding := range []string{"gzip", "deflate"} {
		for _, action := range []string{"header", "flush", "body", "content-length"} {
			for _, wrapped := range []bool{false, true} {
				t.Run(coding+"/"+action+"/wrapped="+strconv.FormatBool(wrapped), func(t *testing.T) {
					g := web.New()
					prefix := ""
					g.Middlewares(func(c *web.Context) {
						c.Writer.Header().Set("Content-Type", "text/plain")
						switch action {
						case "header":
							c.Writer.WriteHeader(201)
						case "flush":
							if err := http.NewResponseController(c.Writer).Flush(); err != nil {
								t.Error(err)
							}
						case "body":
							prefix = "prefix\n"
							c.WriteString(prefix)
						case "content-length":
							c.Writer.Header().Set("Content-Length", strconv.Itoa(len(body)))
							c.Writer.WriteHeader(200)
						}
						if wrapped {
							c.Writer = encodingUnwrapWriter{c.Writer}
						}
					})
					g.Middlewares(Encoding())
					g.GET("/", func(c *web.Context) { c.WriteString(body) })
					rec := runReq(t, g, "GET", "/", coding)
					res := rec.Result()
					defer res.Body.Close()
					if got := res.Header.Get("Content-Encoding"); got != "" || rec.Body.String() != prefix+body {
						t.Fatalf("encoding=%q body bytes=%d", got, rec.Body.Len())
					}
					wantStatus := http.StatusOK
					if action == "header" {
						wantStatus = http.StatusCreated
					}
					if res.StatusCode != wantStatus {
						t.Fatalf("status=%d want=%d", res.StatusCode, wantStatus)
					}
					if action == "content-length" && res.ContentLength != int64(len(body)) {
						t.Fatalf("content length=%d", res.ContentLength)
					}
				})
			}
		}
	}
}

func TestNestedEncodingPanicRestoresHeaders(t *testing.T) {
	for _, coding := range []string{"gzip", "deflate"} {
		for _, location := range []string{"handler", "after inner encoding"} {
			for _, custom := range []bool{false, true} {
				t.Run(coding+"/"+location+"/custom="+strconv.FormatBool(custom), func(t *testing.T) {
					g := web.New()
					want := "Internal Server Error\n"
					if custom {
						want = "custom error\n"
						g.OnPanic(func(c *web.Context, _ any) { http.Error(c.Writer, "custom error", 500) })
					}
					g.Middlewares(Encoding())
					g.Middlewares(func(c *web.Context) {
						c.Next()
						if location == "after inner encoding" {
							panic("outer handler failed")
						}
					})
					group := g.Grep("/api")
					group.Middlewares(Encoding())
					group.GET("/", func(c *web.Context) {
						c.Writer.Header().Set("Content-Type", "text/plain")
						c.WriteString(strings.Repeat("row\n", 512))
						if location == "handler" {
							panic("export failed")
						}
					})
					rec := runReq(t, g, "GET", "/api/", coding)
					if rec.Code != 500 || rec.Header().Get("Content-Encoding") != "" || rec.Body.String() != want {
						t.Fatalf("status=%d encoding=%q body=%q", rec.Code, rec.Header().Get("Content-Encoding"), rec.Body.String())
					}
				})
			}
		}
	}
}

func TestNestedEncodingSuccess(t *testing.T) {
	body := strings.Repeat("row\n", 512)
	for _, coding := range []string{"gzip", "deflate"} {
		t.Run(coding, func(t *testing.T) {
			g := web.New()
			g.Middlewares(Encoding())
			group := g.Grep("/api")
			group.Middlewares(Encoding())
			group.GET("/", func(c *web.Context) { c.WriteString(body) })
			rec := runReq(t, g, "GET", "/api/", coding)
			if rec.Code != 200 || rec.Header().Get("Content-Encoding") != coding {
				t.Fatalf("status=%d headers=%v", rec.Code, rec.Header())
			}
			var reader io.ReadCloser
			var err error
			if coding == "gzip" {
				reader, err = gzip.NewReader(rec.Body)
			} else {
				reader, err = zlib.NewReader(rec.Body)
			}
			if err != nil {
				t.Fatal(err)
			}
			defer reader.Close()
			got, err := io.ReadAll(reader)
			if err != nil || string(got) != body {
				t.Fatalf("decoded bytes=%d err=%v", len(got), err)
			}
		})
	}
}
