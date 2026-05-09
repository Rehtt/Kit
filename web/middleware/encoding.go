package middleware

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"strings"

	"github.com/Rehtt/Kit/web"
)

type EncodeingWriter struct {
	http.ResponseWriter
	writer io.WriteCloser
}

func (w *EncodeingWriter) Write(b []byte) (int, error) {
	if w.writer == nil {
		return w.ResponseWriter.Write(b)
	}
	defer w.writer.Close()
	return w.writer.Write(b)
}

func Encodeing() web.HandlerFunc {
	return func(c *web.Context) {
		encoding := c.Request.Header.Get("Accept-Encoding")
		if encoding == "" {
			return
		}

		switch {
		case strings.Contains(encoding, "gzip"):
			c.Writer.Header().Set("Content-Encoding", "gzip")
			w := gzip.NewWriter(c.Writer)
			c.Writer = &EncodeingWriter{c.Writer, w}
		case strings.Contains(encoding, "deflate"):
			c.Writer.Header().Set("Content-Encoding", "deflate")
			w := zlib.NewWriter(c.Writer)
			c.Writer = &EncodeingWriter{c.Writer, w}
		}
	}
}
