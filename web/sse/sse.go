package sse

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Rehtt/Kit/web"
)

type ResponseType string

const (
	EVENT ResponseType = "event"
	ID    ResponseType = "id"
	RETRY ResponseType = "retry"
	DATA  ResponseType = "data"
)

type Conn struct {
	*web.Context
	http.Flusher
}

func NewConn(ctx *web.Context, deadline ...time.Time) (*Conn, error) {
	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		return nil, http.ErrNotSupported
	}
	var d time.Time
	if len(deadline) > 0 {
		d = deadline[0]
	}
	if err := ctx.SetWriteDeadline(d); err != nil {
		return nil, err
	}

	ctx.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	flusher.Flush()

	return &Conn{ctx, flusher}, nil
}

func (c *Conn) Send(ty ResponseType, data string) error {
	if err := writeField(c, ty, data); err != nil {
		return err
	}
	if _, err := c.WriteString("\n"); err != nil {
		return err
	}
	c.Flush()
	return nil
}

func writeField(w io.StringWriter, ty ResponseType, data string) error {
	value := strings.ReplaceAll(data, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	for {
		line, rest, ok := strings.Cut(value, "\n")
		if _, err := w.WriteString(string(ty) + ": " + line + "\n"); err != nil {
			return err
		}
		if !ok {
			return nil
		}
		value = rest
	}
}
