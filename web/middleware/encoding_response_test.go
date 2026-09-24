package middleware

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Rehtt/Kit/web"
)

func TestEncodingPreservesRanges(t *testing.T) {
	body := strings.Repeat("0123456789abcdef", 256)
	for _, algo := range []string{"gzip", "deflate"} {
		for _, ranges := range []string{"bytes=0-1023", "bytes=0-1023,2048-3071"} {
			t.Run(algo+"/"+ranges, func(t *testing.T) {
				g := web.New()
				g.Middlewares(Encoding(EncodingOption{MinSize: 1}))
				g.GET("/", func(c *web.Context) {
					http.ServeContent(c.Writer, c.Request, "file.txt", time.Time{}, strings.NewReader(body))
					if err := http.NewResponseController(c.Writer).Flush(); err != nil {
						t.Error(err)
					}
				})
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Accept-Encoding", algo)
				req.Header.Set("Range", ranges)
				rec := httptest.NewRecorder()
				g.ServeHTTP(rec, req)
				res := rec.Result()
				defer res.Body.Close()
				if res.StatusCode != 206 || res.Header.Get("Content-Encoding") != "" {
					t.Fatalf("status=%d headers=%v", res.StatusCode, res.Header)
				}
				if ranges == "bytes=0-1023" {
					if res.Header.Get("Content-Range") != "bytes 0-1023/4096" || rec.Body.String() != body[:1024] || res.ContentLength != 1024 {
						t.Fatalf("invalid range: headers=%v len=%d", res.Header, rec.Body.Len())
					}
				} else {
					mediaType, params, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
					if err != nil || mediaType != "multipart/byteranges" {
						t.Fatalf("multipart type: %s %v", mediaType, err)
					}
					reader := multipart.NewReader(res.Body, params["boundary"])
					for _, offset := range []int{0, 2048} {
						part, err := reader.NextPart()
						if err != nil {
							t.Fatal(err)
						}
						data, err := io.ReadAll(part)
						if err != nil {
							t.Fatal(err)
						}
						if string(data) != body[offset:offset+1024] || part.Header.Get("Content-Range") != fmt.Sprintf("bytes %d-%d/4096", offset, offset+1023) {
							t.Fatalf("invalid part: %v", part.Header)
						}
					}
					if _, err := reader.NextPart(); err != io.EOF {
						t.Fatalf("extra part: %v", err)
					}
				}
			})
		}
	}
}

func TestEncodingSkipsManualRangeAndEarlyFlush(t *testing.T) {
	for _, algo := range []string{"gzip", "deflate"} {
		for _, status := range []int{200, 206} {
			for _, flush := range []bool{false, true} {
				rec := httptest.NewRecorder()
				w := acquireWriter(rec, algo, -1, 1, defaultAllowedTypes)
				defer releaseWriter(w)
				w.Header().Set("Content-Type", "text/plain")
				// Test each exclusion independently: Content-Range on 200, or 206 alone.
				if status == 200 {
					w.Header().Set("Content-Range", "bytes 0-4/10")
				}
				w.WriteHeader(status)
				if flush {
					if err := w.FlushError(); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := w.Write([]byte("hello")); err != nil {
					t.Fatal(err)
				}
				w.finish()
				if rec.Code != status || rec.Header().Get("Content-Encoding") != "" || rec.Body.String() != "hello" {
					t.Fatalf("range changed: status=%d headers=%v body=%q", rec.Code, rec.Header(), rec.Body.String())
				}
			}
		}
	}
}

func TestEncodingLocksFinalStatus(t *testing.T) {
	for _, tc := range []struct {
		name   string
		action func(http.ResponseWriter)
		want   []int
		body   string
	}{
		{"implicit", func(w http.ResponseWriter) { w.Write([]byte("hello")); w.WriteHeader(201) }, []int{200}, "hello"},
		{"empty write", func(w http.ResponseWriter) { w.Write(nil); w.WriteHeader(201) }, []int{200}, ""},
		{"explicit", func(w http.ResponseWriter) { w.WriteHeader(201); w.Write([]byte("hello")); w.WriteHeader(202) }, []int{201}, "hello"},
		{"informational", func(w http.ResponseWriter) {
			w.WriteHeader(103)
			w.WriteHeader(201)
			w.Write([]byte("hello"))
			w.WriteHeader(103)
		}, []int{103, 201}, "hello"},
		{"upgrade", func(w http.ResponseWriter) { w.WriteHeader(101); w.WriteHeader(200) }, []int{101}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := &encodingTraceWriter{header: make(http.Header)}
			g := web.New()
			g.Middlewares(Encoding())
			g.GET("/", func(c *web.Context) { tc.action(c.Writer) })
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Accept-Encoding", "gzip")
			g.ServeHTTP(target, req)
			// Real servers generate an implicit 200 when a handler returns without sending headers.
			if !target.final {
				target.WriteHeader(200)
			}
			if !reflect.DeepEqual(target.statuses, tc.want) || target.body.String() != tc.body {
				t.Fatalf("statuses=%v body=%q", target.statuses, target.body.String())
			}
			if tc.name == "upgrade" && target.Header().Get("Content-Encoding") != "" {
				t.Fatal("compressed upgrade")
			}
		})
	}
}

func TestEncodingInvalidStatusAndBufferedPanic(t *testing.T) {
	for _, code := range []int{99, 1000, 201} {
		g := web.New()
		g.Middlewares(Encoding())
		g.GET("/", func(c *web.Context) { c.Writer.WriteHeader(code); c.WriteString("buffered"); panic("handler failed") })
		rec := runReq(t, g, "GET", "/", "gzip")
		if rec.Code != 500 || strings.Contains(rec.Body.String(), "buffered") || rec.Header().Get("Content-Encoding") != "" {
			t.Fatalf("status=%d body=%q headers=%v", rec.Code, rec.Body.String(), rec.Header())
		}
	}
}

func TestEncodingVaryAcrossNegotiation(t *testing.T) {
	for _, method := range []string{"GET", "HEAD"} {
		for _, accept := range []string{"", "gzip;q=0, deflate;q=0", "gzip", "deflate"} {
			for _, existing := range []string{"", "Origin", "accept-encoding", "*"} {
				for _, encoded := range []bool{false, true} {
					g := web.New()
					g.Middlewares(func(c *web.Context) {
						if existing != "" {
							c.Writer.Header().Set("Vary", existing)
						}
						if encoded {
							c.Writer.Header().Set("Content-Encoding", "custom")
						}
					})
					g.Middlewares(Encoding())
					g.GET("/", func(c *web.Context) { c.WriteString(strings.Repeat("text", 512)) })
					rec := runReq(t, g, method, "/", accept)
					values := strings.Join(rec.Result().Header.Values("Vary"), ", ")
					if existing == "*" {
						if values != "*" {
							t.Fatalf("Vary=%q", values)
						}
						continue
					}
					if strings.Count(strings.ToLower(values), "accept-encoding") != 1 || (existing == "Origin" && !strings.Contains(values, "Origin")) {
						t.Fatalf("method=%s accept=%q encoded=%v Vary=%q", method, accept, encoded, values)
					}
				}
			}
		}
	}
}

type encodingTraceWriter struct {
	header                     http.Header
	statuses                   []int
	final                      bool
	body                       bytes.Buffer
	writes, failWrite, flushes int
	writeErr, flushErr         error
}

func (w *encodingTraceWriter) Header() http.Header { return w.header }
func (w *encodingTraceWriter) WriteHeader(code int) {
	if w.final {
		return
	}
	w.statuses = append(w.statuses, code)
	if code == 101 || code >= 200 {
		w.final = true
	}
}
func (w *encodingTraceWriter) Write(p []byte) (int, error) {
	if !w.final {
		w.WriteHeader(200)
	}
	w.writes++
	if w.failWrite == w.writes {
		return 0, w.writeErr
	}
	return w.body.Write(p)
}
func (w *encodingTraceWriter) FlushError() error {
	w.flushes++
	if !w.final {
		w.WriteHeader(200)
	}
	return w.flushErr
}

func TestEncodingFlushErrorStages(t *testing.T) {
	for _, algo := range []string{"gzip", "deflate"} {
		for _, stage := range []string{"commit", "encoder", "underlying"} {
			t.Run(algo+"/"+stage, func(t *testing.T) {
				failure := errors.New(stage + " failed")
				target := &encodingTraceWriter{header: make(http.Header)}
				switch stage {
				case "commit":
					target.failWrite = 1
					target.writeErr = failure
				case "encoder":
					target.failWrite = 2
					target.writeErr = failure
				case "underlying":
					target.flushErr = failure
				}
				w := acquireWriter(target, algo, -1, 1024, defaultAllowedTypes)
				defer releaseWriter(w)
				w.Header().Set("Content-Type", "text/plain")
				if _, err := w.Write([]byte("buffered text")); err != nil {
					t.Fatal(err)
				}
				err := http.NewResponseController(w).Flush()
				if !errors.Is(err, failure) {
					t.Fatalf("Flush=%v want %v", err, failure)
				}
				wantFlushes := 0
				if stage == "underlying" {
					wantFlushes = 1
				}
				if target.flushes != wantFlushes {
					t.Fatalf("underlying flush calls=%d", target.flushes)
				}
			})
		}
	}
}

func TestEncodingFlushUnsupported(t *testing.T) {
	target := &encodingTraceWriter{header: make(http.Header)}
	// Expose only the required ResponseWriter methods.
	w := acquireWriter(struct{ http.ResponseWriter }{target}, "gzip", -1, 1024, defaultAllowedTypes)
	defer releaseWriter(w)
	if err := http.NewResponseController(w).Flush(); !errors.Is(err, http.ErrNotSupported) {
		t.Fatalf("Flush=%v", err)
	}
}
