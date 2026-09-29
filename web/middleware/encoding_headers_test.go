package middleware

import (
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

func TestEncodingFreezesResponseHeaders(t *testing.T) {
	for _, coding := range []string{"identity", "gzip", "deflate"} {
		for _, layers := range []int{1, 2} {
			for _, size := range []int{16, 2048} {
				for _, action := range []string{"header", "write", "empty", "flush"} {
					t.Run(fmt.Sprintf("%s/layers=%d/size=%d/%s", coding, layers, size, action), func(t *testing.T) {
						g := web.New()
						for i := 0; i < layers; i++ {
							g.Middlewares(Encoding())
						}
						body := strings.Repeat("x", size)
						g.GET("/", func(c *web.Context) {
							h := c.Writer.Header()
							h.Set("Content-Type", "text/plain")
							h.Set("Cache-Control", "private")
							h["X-Frozen"] = []string{"before"}
							switch action {
							case "header":
								c.Writer.WriteHeader(201)
							case "write":
								c.WriteString(body)
							case "empty":
								c.Write(nil)
							case "flush":
								if err := http.NewResponseController(c.Writer).Flush(); err != nil {
									t.Error(err)
								}
							}
							// Include mutations through retained maps and value slices.
							h.Set("Content-Type", "application/octet-stream")
							h.Set("Cache-Control", "public, max-age=3600")
							h["X-Frozen"][0] = "after"
							h.Set("Content-Encoding", "bogus")
							c.Writer.WriteHeader(202)
							if action != "write" {
								c.WriteString(body)
							}
						})
						res := runReq(t, g, "GET", "/", coding).Result()
						defer res.Body.Close()
						wantStatus := 200
						if action == "header" {
							wantStatus = 201
						}
						if res.StatusCode != wantStatus || res.Header.Get("Content-Type") != "text/plain" ||
							res.Header.Get("Cache-Control") != "private" || res.Header.Get("X-Frozen") != "before" {
							t.Fatalf("status=%d headers=%v", res.StatusCode, res.Header)
						}
						wantEncoding := ""
						if coding != "identity" && (size >= defaultMinSize || action == "flush") {
							wantEncoding = coding
						}
						if got := res.Header.Get("Content-Encoding"); got != wantEncoding {
							t.Fatalf("encoding=%q want=%q", got, wantEncoding)
						}
						assertDecodedBody(t, res, body)
					})
				}
			}
		}
	}
}

func assertDecodedBody(t *testing.T, res *http.Response, want string) {
	t.Helper()
	var reader io.Reader = res.Body
	var decoder io.ReadCloser
	var err error
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		decoder, err = gzip.NewReader(res.Body)
	case "deflate":
		decoder, err = zlib.NewReader(res.Body)
	}
	if err != nil {
		t.Fatal(err)
	}
	if decoder != nil {
		defer decoder.Close()
		reader = decoder
	}
	body, err := io.ReadAll(reader)
	if err != nil || string(body) != want {
		t.Fatalf("decoded bytes=%d want=%d err=%v", len(body), len(want), err)
	}
}

func TestEncodingPreservesTrailers(t *testing.T) {
	for _, coding := range []string{"identity", "gzip", "deflate"} {
		for _, layers := range []int{1, 2} {
			for _, size := range []int{16, 2048} {
				t.Run(fmt.Sprintf("%s/layers=%d/size=%d", coding, layers, size), func(t *testing.T) {
					g := web.New()
					for i := 0; i < layers; i++ {
						g.Middlewares(Encoding())
					}
					body := strings.Repeat("x", size)
					g.GET("/", func(c *web.Context) {
						h := c.Writer.Header()
						h.Set("Content-Type", "text/plain")
						h.Set("Trailer", "X-Checksum, X-Removed")
						h.Set("X-Removed", "before")
						h.Set(http.TrailerPrefix+"X-Removed-Prefix", "before")
						c.WriteString(body)
						h.Set("X-Checksum", "done")
						h.Del("X-Removed")
						h.Del(http.TrailerPrefix + "X-Removed-Prefix")
						h.Set(http.TrailerPrefix+"X-Late", "late value")
					})
					// Read through EOF to check actual HTTP trailer framing.
					server := httptest.NewServer(g)
					defer server.Close()
					req, _ := http.NewRequest("GET", server.URL, nil)
					req.Header.Set("Accept-Encoding", coding)
					res, err := server.Client().Do(req)
					if err != nil {
						t.Fatal(err)
					}
					defer res.Body.Close()
					assertDecodedBody(t, res, body)
					if res.Trailer.Get("X-Checksum") != "done" || res.Trailer.Get("X-Late") != "late value" ||
						res.Trailer.Get("X-Removed") != "" || res.Trailer.Get("X-Removed-Prefix") != "" {
						t.Fatalf("trailers=%v", res.Trailer)
					}
					if res.Header.Get("X-Checksum") != "" || res.Header.Get("X-Late") != "" {
						t.Fatalf("trailers leaked into headers: %v", res.Header)
					}
				})
			}
		}
	}
}

type headerTraceWriter struct {
	encodingTraceWriter
	headers []http.Header
}

func (w *headerTraceWriter) WriteHeader(code int) {
	w.headers = append(w.headers, w.Header().Clone())
	w.encodingTraceWriter.WriteHeader(code)
}

func TestEncodingInformationalHeadersRemainMutable(t *testing.T) {
	target := &headerTraceWriter{encodingTraceWriter: encodingTraceWriter{header: make(http.Header)}}
	w := acquireWriter(target, "gzip", -1, defaultMinSize, defaultAllowedTypes)
	defer releaseWriter(w)
	w.Header().Set("X-Phase", "early")
	w.WriteHeader(103)
	w.Header().Set("X-Phase", "final")
	w.WriteHeader(201)
	w.Header().Set("X-Phase", "too late")
	w.Write([]byte("hello"))
	w.finish()
	if len(target.headers) != 2 || target.headers[0].Get("X-Phase") != "early" || target.headers[1].Get("X-Phase") != "final" {
		t.Fatalf("headers=%v", target.headers)
	}
}

func TestNestedEncodingAfterOuterResponseStarted(t *testing.T) {
	for _, coding := range []string{"gzip", "deflate"} {
		for _, prefix := range []string{"", "prefix\n"} {
			g := web.New()
			g.Middlewares(Encoding())
			g.Middlewares(func(c *web.Context) {
				c.Writer.Header().Set("Content-Type", "text/plain")
				c.Writer.WriteHeader(201)
				c.WriteString(prefix)
				c.Writer = encodingUnwrapWriter{c.Writer}
			})
			g.Middlewares(Encoding())
			body := strings.Repeat("row\n", 512)
			g.GET("/", func(c *web.Context) { c.WriteString(body) })
			res := runReq(t, g, "GET", "/", coding).Result()
			if res.StatusCode != 201 || res.Header.Get("Content-Encoding") != coding {
				t.Fatalf("status=%d headers=%v", res.StatusCode, res.Header)
			}
			assertDecodedBody(t, res, prefix+body)
			res.Body.Close()
		}
	}
}
