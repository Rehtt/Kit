package web

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type readFailure struct{ err error }

func (r readFailure) Read([]byte) (int, error) { return 0, r.err }

type readerFromRecorder struct {
	*httptest.ResponseRecorder
	copied int64
	calls  int
}

func (w *readerFromRecorder) ReadFrom(r io.Reader) (int64, error) {
	w.calls++
	n, err := io.Copy(w.ResponseRecorder, r)
	w.copied += n
	return n, err
}

func TestReadFromDoesNotCommitWithoutData(t *testing.T) {
	failure := errors.New("read failed")
	for _, fast := range []bool{false, true} {
		for _, wantErr := range []error{nil, failure} {
			rec := httptest.NewRecorder()
			var underlying http.ResponseWriter = rec
			fastWriter := &readerFromRecorder{ResponseRecorder: rec}
			if fast {
				underlying = fastWriter
			}
			w := &responseWriter{}
			w.reset(underlying)
			var src io.Reader = strings.NewReader("")
			if wantErr != nil {
				src = readFailure{wantErr}
			}
			n, err := w.ReadFrom(src)
			if n != 0 || !errors.Is(err, wantErr) {
				t.Fatalf("ReadFrom = %d, %v", n, err)
			}
			if w.Written() || w.Status() != 0 || w.Size() != 0 || fastWriter.calls != 0 {
				t.Fatalf("empty read changed state: %+v, fast calls=%d", w, fastWriter.calls)
			}
			w.WriteHeader(http.StatusCreated)
			if rec.Code != http.StatusCreated {
				t.Fatalf("status = %d", rec.Code)
			}
		}
	}
}

func TestReadFromTracksBodyAndPreservesFastPath(t *testing.T) {
	body := strings.Repeat("plain text ", 200)
	for _, fast := range []bool{false, true} {
		for _, explicit := range []bool{false, true} {
			rec := httptest.NewRecorder()
			fastWriter := &readerFromRecorder{ResponseRecorder: rec}
			var underlying http.ResponseWriter = rec
			if fast {
				underlying = fastWriter
			}
			w := &responseWriter{}
			w.reset(underlying)
			wantStatus := http.StatusOK
			if explicit {
				wantStatus = http.StatusCreated
				w.WriteHeader(wantStatus)
			}
			n, err := w.ReadFrom(strings.NewReader(body))
			if err != nil || n != int64(len(body)) || w.Size() != len(body) || !w.Written() || w.Status() != wantStatus || rec.Body.String() != body {
				t.Fatalf("fast=%v explicit=%v: n=%d err=%v status=%d size=%d", fast, explicit, n, err, w.Status(), w.Size())
			}
			if !explicit && rec.Result().Header.Get("Content-Type") != "text/plain; charset=utf-8" {
				t.Fatal("lost content sniffing")
			}
			if fast {
				wantCopied := int64(len(body))
				if !explicit {
					wantCopied -= 512
				}
				if fastWriter.calls != 1 || fastWriter.copied != wantCopied {
					t.Fatalf("fast path calls=%d bytes=%d", fastWriter.calls, fastWriter.copied)
				}
			}
		}
	}
}

func TestUncommittedResponsePanicRecovers500(t *testing.T) {
	for _, action := range []func(*Context){
		func(c *Context) { c.Writer.WriteHeader(99) },
		func(c *Context) { c.Writer.WriteHeader(1000) },
		func(c *Context) { _, _ = c.ReadFrom(strings.NewReader("")); panic("after empty copy") },
	} {
		g := New()
		g.GET("/", action)
		rec := httptest.NewRecorder()
		g.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d", rec.Code)
		}
	}
}

type plainResponseWriter struct{ header http.Header }

func (w *plainResponseWriter) Header() http.Header       { return w.header }
func (*plainResponseWriter) WriteHeader(int)             {}
func (*plainResponseWriter) Write(p []byte) (int, error) { return len(p), nil }

type errorFlusher struct {
	*plainResponseWriter
	err   error
	calls int
}

func (w *errorFlusher) FlushError() error { w.calls++; return w.err }

type unwrapWriter struct{ http.ResponseWriter }

func (w unwrapWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func TestResponseWriterFlushError(t *testing.T) {
	failure := errors.New("flush failed")
	for _, wantErr := range []error{nil, failure, http.ErrNotSupported} {
		underlying := &plainResponseWriter{header: make(http.Header)}
		flusher := &errorFlusher{plainResponseWriter: underlying, err: wantErr}
		var target http.ResponseWriter = flusher
		if wantErr == http.ErrNotSupported {
			target = underlying
		}
		w := &responseWriter{}
		w.reset(unwrapWriter{target})
		err := http.NewResponseController(w).Flush()
		if !errors.Is(err, wantErr) {
			t.Fatalf("Flush = %v, want %v", err, wantErr)
		}
		if w.Written() != (wantErr != http.ErrNotSupported) {
			t.Fatalf("Written = %v after %v", w.Written(), err)
		}
		if wantErr == http.ErrNotSupported {
			if w.Status() != 0 {
				t.Fatalf("unsupported status = %d", w.Status())
			}
		} else if flusher.calls != 1 || w.Status() != http.StatusOK {
			t.Fatalf("calls=%d status=%d", flusher.calls, w.Status())
		}
	}
}

func TestResponseWriterFailedWriteHeaderDoesNotCommit(t *testing.T) {
	w := &responseWriter{}
	w.reset(panicHeaderWriter{&plainResponseWriter{header: make(http.Header)}})
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected panic")
			}
		}()
		w.WriteHeader(http.StatusCreated)
	}()
	if w.Written() || w.Status() != 0 {
		t.Fatalf("failed WriteHeader changed state: %+v", w)
	}
}

type panicHeaderWriter struct{ *plainResponseWriter }

func (panicHeaderWriter) WriteHeader(int) { panic("header failed") }
