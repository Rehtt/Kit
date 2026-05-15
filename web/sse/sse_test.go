package sse

import (
	"errors"
	"net/http"
	"testing"

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
