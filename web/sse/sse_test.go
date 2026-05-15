package sse

import (
	"errors"
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
