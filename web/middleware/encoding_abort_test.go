package middleware

import (
	"bufio"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Rehtt/Kit/web"
)

// Exercise real HTTP framing as well as decompression: a truncated export must
// fail at the client even if its status was already sent as 200.
func TestPanicInterruptsResponse(t *testing.T) {
	for _, protocol := range []string{"http1", "http2"} {
		for _, coding := range []string{"identity", "gzip", "deflate"} {
			t.Run(protocol+"/"+coding, func(t *testing.T) {
				release := make(chan struct{})
				releaseResponse := sync.OnceFunc(func() { close(release) })
				g := web.New()
				g.Middlewares(Encoding())
				g.GET("/", func(c *web.Context) {
					c.Writer.Header().Set("Content-Type", "text/plain")
					if _, err := c.WriteString(strings.Repeat("row\n", 512)); err != nil {
						t.Error(err)
						return
					}
					if err := http.NewResponseController(c.Writer).Flush(); err != nil {
						t.Error(err)
						return
					}
					// Let the client consume the initial data before resetting the
					// stream, so HTTP/2 scheduling cannot race response setup.
					<-release
					panic("export failed")
				})
				s := httptest.NewUnstartedServer(g)
				if protocol == "http2" {
					s.EnableHTTP2 = true
					s.StartTLS()
				} else {
					s.Start()
				}
				defer s.Close()
				defer releaseResponse()
				client := s.Client()
				client.Timeout = 5 * time.Second
				req, err := http.NewRequest("GET", s.URL, nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Accept-Encoding", coding)
				res, err := client.Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer res.Body.Close()
				if res.StatusCode != http.StatusOK || (protocol == "http2" && res.ProtoMajor != 2) {
					t.Fatalf("unexpected response: %s %s", res.Proto, res.Status)
				}
				var body io.Reader = res.Body
				switch coding {
				case "gzip":
					body, err = gzip.NewReader(res.Body)
				case "deflate":
					body, err = zlib.NewReader(res.Body)
				}
				if err != nil {
					t.Fatal(err)
				}
				prefix := make([]byte, len("row\n"))
				if _, err := io.ReadFull(body, prefix); err != nil || string(prefix) != "row\n" {
					t.Fatalf("initial data=%q err=%v", prefix, err)
				}
				releaseResponse()
				if _, err := io.ReadAll(body); err == nil {
					t.Fatal("partial response appeared complete to the client")
				}
			})
		}
	}
}

type upgradeWriter struct {
	*httptest.ResponseRecorder
	hijacked bool
}

var errUpgradeTest = errors.New("upgrade test hijack")

func (w *upgradeWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	return nil, nil, errUpgradeTest
}

func TestEncodingPreservesUpgradeHijacker(t *testing.T) {
	for _, coding := range []string{"gzip", "deflate"} {
		t.Run(coding, func(t *testing.T) {
			g := web.New()
			g.Middlewares(Encoding())
			g.GET("/", func(c *web.Context) {
				h, ok := c.Writer.(http.Hijacker)
				if !ok {
					t.Error("upgrade handler lost http.Hijacker")
					return
				}
				if _, _, err := h.Hijack(); !errors.Is(err, errUpgradeTest) {
					t.Errorf("Hijack error=%v", err)
				}
			})
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Accept-Encoding", coding)
			req.Header.Set("Connection", "keep-alive, Upgrade")
			req.Header.Set("Upgrade", "websocket")
			w := &upgradeWriter{ResponseRecorder: httptest.NewRecorder()}
			g.ServeHTTP(w, req)
			if !w.hijacked || w.Header().Get("Content-Encoding") != "" || w.Body.Len() != 0 {
				t.Fatalf("hijacked=%v headers=%v body=%q", w.hijacked, w.Header(), w.Body.String())
			}
		})
	}
}
