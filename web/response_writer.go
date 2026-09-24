package web

import (
	"bufio"
	"errors"
	"fmt"
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

// ResponseWriterUnwrapper 是可选接口，用于来访问底层的http.ResponseWriter
type ResponseWriterUnwrapper interface {
	Unwrap() http.ResponseWriter
}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
	wrote  bool // final response has been committed
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
	if code < 100 || code > 999 {
		panic(fmt.Sprintf("invalid WriteHeader code %v", code))
	}
	// Informational responses (except 101) do not commit the final response.
	if code >= 100 && code < 200 && code != http.StatusSwitchingProtocols {
		w.ResponseWriter.WriteHeader(code)
		return
	}
	w.ResponseWriter.WriteHeader(code)
	w.wrote = true
	w.status = code
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
	_ = w.FlushError()
}

// FlushError preserves flushing errors and reports unsupported underlying writers.
func (w *responseWriter) FlushError() error {
	err := http.NewResponseController(w.ResponseWriter).Flush()
	// A supported flush can commit headers even when flushing the connection fails.
	if !errors.Is(err, http.ErrNotSupported) && !w.wrote {
		w.wrote = true
		w.status = http.StatusOK
	}
	return err
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
	var prefix int64
	writer := struct{ io.Writer }{w}
	if !w.wrote {
		// Wait for real data before committing; also retain Content-Type sniffing
		// before handing the remainder to a potentially zero-copy ReaderFrom.
		var err error
		prefix, err = io.Copy(writer, io.LimitReader(src, 512))
		if err != nil || prefix < 512 {
			return prefix, err
		}
	}
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		n, err := rf.ReadFrom(src)
		w.size += int(n)
		return prefix + n, err
	}
	n, err := io.Copy(writer, src)
	return prefix + n, err
}

func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
