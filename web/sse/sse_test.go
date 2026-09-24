package sse

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Rehtt/Kit/web"
)

func TestConnSendFormatsMultilineData(t *testing.T) {
	rw := &stringResponseWriter{header: make(http.Header)}
	conn := &Conn{
		Context: &web.Context{Writer: rw},
		Flusher: rw,
	}

	if err := conn.Send(DATA, "a\r\nb\rc"); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	const want = "data: a\ndata: b\ndata: c\n\n"
	if rw.body != want {
		t.Fatalf("body = %q, want %q", rw.body, want)
	}
	if !rw.flushed {
		t.Fatal("Send() did not flush")
	}
}

func TestConnSendReturnsWriteError(t *testing.T) {
	wantErr := errors.New("write failed")
	rw := &stringResponseWriter{header: make(http.Header), writeErr: wantErr}
	conn := &Conn{
		Context: &web.Context{Writer: rw},
		Flusher: rw,
	}

	if err := conn.Send(DATA, "a"); !errors.Is(err, wantErr) {
		t.Fatalf("Send() error = %v, want %v", err, wantErr)
	}
	if rw.flushed {
		t.Fatal("Send() flushed after write error")
	}
}

func TestConnSendRejectsInvalidRetry(t *testing.T) {
	rw := &stringResponseWriter{header: make(http.Header)}
	conn := &Conn{
		Context: &web.Context{Writer: rw},
		Flusher: rw,
	}

	if err := conn.Send(RETRY, "abc"); !errors.Is(err, ErrInvalidRetry) {
		t.Fatalf("Send() error = %v, want %v", err, ErrInvalidRetry)
	}
	if rw.body != "" {
		t.Fatalf("body = %q, want empty", rw.body)
	}
	if rw.flushed {
		t.Fatal("Send() flushed after invalid retry")
	}
}

func TestConnSendEventFormatsStandardEvent(t *testing.T) {
	rw := &stringResponseWriter{header: make(http.Header)}
	conn := &Conn{
		Context: &web.Context{Writer: rw},
		Flusher: rw,
	}

	err := conn.SendEvent(Event{
		Event: "update",
		ID:    "42",
		Retry: 3 * time.Second,
		Data:  "a\nb",
	})
	if err != nil {
		t.Fatalf("SendEvent() error = %v", err)
	}

	const want = "id: 42\nevent: update\nretry: 3000\ndata: a\ndata: b\n\n"
	if rw.body != want {
		t.Fatalf("body = %q, want %q", rw.body, want)
	}
	if !rw.flushed {
		t.Fatal("SendEvent() did not flush")
	}
}

func TestConnSendEventSendsEmptyDataEvent(t *testing.T) {
	rw := &stringResponseWriter{header: make(http.Header)}
	conn := &Conn{
		Context: &web.Context{Writer: rw},
		Flusher: rw,
	}

	if err := conn.SendEvent(Event{}); err != nil {
		t.Fatalf("SendEvent() error = %v", err)
	}

	const want = "data: \n\n"
	if rw.body != want {
		t.Fatalf("body = %q, want %q", rw.body, want)
	}
}

func TestConnCommentFormatsHeartbeat(t *testing.T) {
	rw := &stringResponseWriter{header: make(http.Header)}
	conn := &Conn{
		Context: &web.Context{Writer: rw},
		Flusher: rw,
	}

	if err := conn.Comment("a\r\nb"); err != nil {
		t.Fatalf("Comment() error = %v", err)
	}

	const want = ": a\n: b\n\n"
	if rw.body != want {
		t.Fatalf("body = %q, want %q", rw.body, want)
	}
	if !rw.flushed {
		t.Fatal("Comment() did not flush")
	}
}

func TestConnLastEventID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	req.Header.Set("Last-Event-ID", "42")
	conn := &Conn{
		Context: &web.Context{Request: req},
	}

	if got := conn.LastEventID(); got != "42" {
		t.Fatalf("LastEventID() = %q, want %q", got, "42")
	}
}

func TestNewConnSetsSSEHeaders(t *testing.T) {
	rw := &stringResponseWriter{header: make(http.Header)}
	ctx := &web.Context{Writer: rw}

	conn, err := NewConn(ctx)
	if err != nil {
		t.Fatalf("NewConn() error = %v", err)
	}
	if conn == nil {
		t.Fatal("NewConn() returned nil conn")
	}

	tests := map[string]string{
		"Content-Type":                "text/event-stream; charset=utf-8",
		"Cache-Control":               "no-cache",
		"X-Accel-Buffering":           "no",
		"Access-Control-Allow-Origin": "*",
	}
	for key, want := range tests {
		if got := rw.header.Get(key); got != want {
			t.Fatalf("header %s = %q, want %q", key, got, want)
		}
	}
	if !rw.flushed {
		t.Fatal("NewConn() did not flush")
	}
}

type stringResponseWriter struct {
	header   http.Header
	body     string
	writeErr error
	flushed  bool
}

func (w *stringResponseWriter) Header() http.Header {
	return w.header
}

func (w *stringResponseWriter) Write(b []byte) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	w.body += string(b)
	return len(b), nil
}

func (w *stringResponseWriter) WriteHeader(int) {}

func (w *stringResponseWriter) WriteString(s string) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	w.body += s
	return len(s), nil
}

func (w *stringResponseWriter) Flush() {
	w.flushed = true
}

// FlushError-only writers and unwrap-only wrappers must work through the controller.
type flushErrorWriter struct {
	http.ResponseWriter
	err   error
	calls int
}

func (w *flushErrorWriter) FlushError() error { w.calls++; return w.err }

type unwrapResponseWriter struct{ http.ResponseWriter }

func (w unwrapResponseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

type legacyAndErrorFlusher struct {
	err                     error
	errorCalls, legacyCalls int
}

func (f *legacyAndErrorFlusher) Flush()            { f.legacyCalls++ }
func (f *legacyAndErrorFlusher) FlushError() error { f.errorCalls++; return f.err }

func TestNewConnFlushErrors(t *testing.T) {
	failure := errors.New("initial flush failed")
	for _, wrapped := range []bool{false, true} {
		for _, wantErr := range []error{nil, failure, http.ErrNotSupported} {
			rec := httptest.NewRecorder()
			// Hide the recorder's Flusher, then optionally add only FlushError.
			var writer http.ResponseWriter = struct{ http.ResponseWriter }{rec}
			f := &flushErrorWriter{ResponseWriter: writer, err: wantErr}
			if wantErr != http.ErrNotSupported {
				writer = f
			}
			if wrapped {
				writer = unwrapResponseWriter{writer}
			}
			conn, err := NewConn(&web.Context{Writer: writer})
			if !errors.Is(err, wantErr) {
				t.Fatalf("NewConn=%v, want %v", err, wantErr)
			}
			if (conn == nil) != (wantErr != nil) {
				t.Fatalf("conn=%v err=%v", conn, err)
			}
			if wantErr != http.ErrNotSupported && f.calls != 1 {
				t.Fatalf("flush calls=%d", f.calls)
			}
		}
	}
}

func TestConnOperationsReturnFlushError(t *testing.T) {
	operations := map[string]func(*Conn) error{
		"send":    func(c *Conn) error { return c.Send(DATA, "hello") },
		"event":   func(c *Conn) error { return c.SendEvent(Event{Data: "hello"}) },
		"comment": func(c *Conn) error { return c.Comment("hello") },
		"ping":    func(c *Conn) error { return c.Ping() },
		"flush":   func(c *Conn) error { return c.FlushError() },
	}
	for name, operation := range operations {
		for _, direct := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/direct=%v", name, direct), func(t *testing.T) {
				failure := errors.New("flush failed")
				rw := &stringResponseWriter{header: make(http.Header)}
				f := &legacyAndErrorFlusher{err: failure}
				ctx := &web.Context{Writer: rw}
				var conn *Conn
				var controllerWriter *flushErrorWriter
				if direct {
					conn = &Conn{Context: ctx, Flusher: f}
				} else {
					controllerWriter = &flushErrorWriter{ResponseWriter: rw}
					ctx.Writer = unwrapResponseWriter{controllerWriter}
					var err error
					conn, err = NewConn(ctx)
					if err != nil {
						t.Fatal(err)
					}
					controllerWriter.err = failure
				}
				if err := operation(conn); !errors.Is(err, failure) {
					t.Fatalf("operation error=%v", err)
				}
				if direct && (f.errorCalls != 1 || f.legacyCalls != 0) {
					t.Fatalf("error calls=%d legacy calls=%d", f.errorCalls, f.legacyCalls)
				}
				if !direct && controllerWriter.calls != 2 {
					t.Fatalf("flush calls=%d", controllerWriter.calls)
				}
			})
		}
	}
}

func TestNewConnDetectsUnsupportedWebWriter(t *testing.T) {
	g := web.New()
	g.GET("/", func(c *web.Context) {
		conn, err := NewConn(c)
		if conn != nil || !errors.Is(err, http.ErrNotSupported) {
			t.Errorf("NewConn=%v, %v", conn, err)
		}
		if c.Writer.(web.ResponseWriter).Written() {
			t.Error("unsupported flush committed response")
		}
	})
	rec := httptest.NewRecorder()
	g.ServeHTTP(struct{ http.ResponseWriter }{rec}, httptest.NewRequest("GET", "/", nil))
}
