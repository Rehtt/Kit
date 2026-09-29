package sse

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Rehtt/Kit/web"
	"github.com/Rehtt/Kit/web/middleware"
)

type deadlineWriter struct {
	*httptest.ResponseRecorder
	deadline time.Time
	err      error
}

func (w *deadlineWriter) SetWriteDeadline(deadline time.Time) error {
	w.deadline = deadline
	return w.err
}

func TestNewConnWriteDeadline(t *testing.T) {
	failure := errors.New("deadline failed")
	explicit := time.Now().Add(time.Minute)
	for _, supplied := range []bool{false, true} {
		for _, deadlineErr := range []error{nil, http.ErrNotSupported, fmt.Errorf("wrapped: %w", http.ErrNotSupported), failure} {
			t.Run(fmt.Sprintf("explicit=%v/error=%v", supplied, deadlineErr), func(t *testing.T) {
				w := &deadlineWriter{ResponseRecorder: httptest.NewRecorder(), deadline: explicit, err: deadlineErr}
				ctx := &web.Context{Writer: unwrapResponseWriter{w}}
				var deadlines []time.Time
				if supplied {
					deadlines = []time.Time{explicit}
				}
				conn, err := NewConn(ctx, deadlines...)
				wantErr := deadlineErr
				if !supplied && errors.Is(deadlineErr, http.ErrNotSupported) {
					wantErr = nil
				}
				if !errors.Is(err, wantErr) || (conn == nil) != (wantErr != nil) {
					t.Fatalf("conn=%v err=%v want=%v", conn, err, wantErr)
				}
				if w.deadline.IsZero() == supplied || supplied && !w.deadline.Equal(explicit) {
					t.Fatalf("deadline=%v explicit=%v", w.deadline, supplied)
				}
				if w.Flushed != (wantErr == nil) {
					t.Fatalf("flushed=%v error=%v", w.Flushed, wantErr)
				}
			})
		}
	}
}

func TestNewConnPreservesCORSPolicy(t *testing.T) {
	for _, origin := range []string{"", "https://app.example", "*"} {
		g := web.New()
		g.Middlewares(func(c *web.Context) {
			if origin != "" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
				c.Writer.Header().Set("Vary", "Origin")
			}
		})
		g.GET("/", func(c *web.Context) {
			if _, err := NewConn(c); err != nil {
				t.Error(err)
			}
		})
		w := httptest.NewRecorder()
		g.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
		h := w.Result().Header
		if h.Get("Access-Control-Allow-Origin") != origin {
			t.Fatalf("origin=%q headers=%v", origin, h)
		}
		if origin != "" && (h.Get("Access-Control-Allow-Credentials") != "true" || h.Get("Vary") != "Origin") {
			t.Fatalf("CORS policy changed: %v", h)
		}
	}
}

// A real HTTP server over an in-memory connection exercises write deadlines
// without requiring a TCP listener. Accept remains blocked until Close.
type pipeListener struct {
	conn      net.Conn
	accept    sync.Once
	closeOnce sync.Once
	done      chan struct{}
}

func (l *pipeListener) Accept() (net.Conn, error) {
	var conn net.Conn
	l.accept.Do(func() { conn = l.conn })
	if conn != nil {
		return conn, nil
	}
	<-l.done
	return nil, net.ErrClosed
}

func (l *pipeListener) Close() error {
	l.closeOnce.Do(func() { close(l.done) })
	return nil
}

func (l *pipeListener) Addr() net.Addr { return l.conn.LocalAddr() }

func TestSSEOutlivesServerWriteTimeout(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		for _, coding := range []string{"identity", "gzip"} {
			t.Run(fmt.Sprintf("explicit=%v/%s", explicit, coding), func(t *testing.T) {
				g := web.New()
				g.Server.WriteTimeout = 100 * time.Millisecond
				g.Middlewares(middleware.Encoding())
				result := make(chan error, 1)
				g.GET("/", func(c *web.Context) {
					var deadlines []time.Time
					if explicit {
						deadlines = []time.Time{time.Now().Add(g.Server.WriteTimeout)}
					}
					conn, err := NewConn(c, deadlines...)
					if err == nil {
						time.Sleep(2 * g.Server.WriteTimeout)
						err = conn.Send(DATA, "still alive")
					}
					result <- err
				})
				server, client := net.Pipe()
				listener := &pipeListener{conn: server, done: make(chan struct{})}
				done := make(chan error, 1)
				go func() { done <- g.Server.Serve(listener) }()
				t.Cleanup(func() {
					client.Close()
					g.Server.Close()
					<-done
				})
				client.SetDeadline(time.Now().Add(5 * time.Second))
				req, _ := http.NewRequest("GET", "http://example/", nil)
				req.Close = true
				req.Header.Set("Accept-Encoding", coding)
				if err := req.Write(client); err != nil {
					t.Fatal(err)
				}
				res, err := http.ReadResponse(bufio.NewReader(client), req)
				if err != nil {
					t.Fatal(err)
				}
				_, readErr := io.ReadAll(res.Body)
				res.Body.Close()
				select {
				case err := <-result:
					if explicit {
						var timeout net.Error
						if !errors.As(err, &timeout) || !timeout.Timeout() {
							t.Fatalf("expected explicit deadline timeout, got %v", err)
						}
					} else if err != nil || readErr != nil {
						t.Fatalf("send=%v read=%v", err, readErr)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("handler did not finish")
				}
			})
		}
	}
}
