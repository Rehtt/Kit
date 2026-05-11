package web

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"

	kitstrings "github.com/Rehtt/Kit/strings"
)

// ResponseWriter 在 http.ResponseWriter 之上跟踪 status/size/written，
// 并透传 Flusher/Hijacker/Pusher/ReaderFrom。
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	http.Hijacker
	http.Pusher
	io.ReaderFrom

	Status() int
	Size() int
	Written() bool
}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
	wrote  bool
}

func (w *responseWriter) reset(rw http.ResponseWriter) {
	w.ResponseWriter = rw
	w.status = 0
	w.size = 0
	w.wrote = false
}

func (w *responseWriter) WriteHeader(code int) {
	if w.wrote {
		return
	}
	w.wrote = true
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.wrote = true
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

func (w *responseWriter) WriteString(s string) (int, error) {
	if !w.wrote {
		w.wrote = true
		w.status = http.StatusOK
	}
	if sw, ok := w.ResponseWriter.(io.StringWriter); ok {
		n, err := sw.WriteString(s)
		w.size += n
		return n, err
	}
	n, err := w.ResponseWriter.Write(kitstrings.UnsafeStringToBytes(s))
	w.size += n
	return n, err
}

func (w *responseWriter) Status() int   { return w.status }
func (w *responseWriter) Size() int     { return w.size }
func (w *responseWriter) Written() bool { return w.wrote }

func (w *responseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		if !w.wrote {
			w.wrote = true
			w.status = http.StatusOK
		}
		f.Flush()
	}
}

var errNotHijackable = errors.New("[web] underlying ResponseWriter does not implement http.Hijacker")

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		w.wrote = true
		return h.Hijack()
	}
	return nil, nil, errNotHijackable
}

func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (w *responseWriter) ReadFrom(src io.Reader) (int64, error) {
	if !w.wrote {
		w.wrote = true
		w.status = http.StatusOK
	}
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		n, err := rf.ReadFrom(src)
		w.size += int(n)
		return n, err
	}
	n, err := io.Copy(struct{ io.Writer }{w.ResponseWriter}, src)
	w.size += int(n)
	return n, err
}
