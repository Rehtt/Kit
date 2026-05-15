package sse

import (
	"errors"
	"io"
	"net/http"
	"strconv"
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

var ErrInvalidRetry = errors.New("sse: retry must be a single non-negative integer")

type Conn struct {
	*web.Context
	http.Flusher
}

type Event struct {
	Event string
	ID    string
	Retry time.Duration
	Data  string
}

func NewConn(ctx *web.Context, deadline ...time.Time) (*Conn, error) {
	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		return nil, http.ErrNotSupported
	}
	if len(deadline) > 0 {
		if err := ctx.SetWriteDeadline(deadline[0]); err != nil {
			return nil, err
		}
	}

	ctx.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no")
	ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	flusher.Flush()

	return &Conn{ctx, flusher}, nil
}

func (c *Conn) Send(ty ResponseType, data string) error {
	if ty == RETRY {
		if err := validateRetry(data); err != nil {
			return err
		}
	}
	if err := writeField(c, ty, data); err != nil {
		return err
	}
	if _, err := c.WriteString("\n"); err != nil {
		return err
	}
	c.Flush()
	return nil
}

func (c *Conn) SendEvent(event Event) error {
	if event.ID != "" {
		if err := writeField(c, ID, event.ID); err != nil {
			return err
		}
	}
	if event.Event != "" {
		if err := writeField(c, EVENT, event.Event); err != nil {
			return err
		}
	}
	if event.Retry > 0 {
		if err := writeRetry(c, event.Retry); err != nil {
			return err
		}
	}
	if err := writeField(c, DATA, event.Data); err != nil {
		return err
	}
	if _, err := c.WriteString("\n"); err != nil {
		return err
	}
	c.Flush()
	return nil
}

func (c *Conn) Comment(comment string) error {
	if err := writeComment(c, comment); err != nil {
		return err
	}
	if _, err := c.WriteString("\n"); err != nil {
		return err
	}
	c.Flush()
	return nil
}

func (c *Conn) Ping() error {
	return c.Comment("ping")
}

func (c *Conn) LastEventID() string {
	if c.Request == nil {
		return ""
	}
	return c.Request.Header.Get("Last-Event-ID")
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

func writeRetry(w io.StringWriter, retry time.Duration) error {
	_, err := w.WriteString(string(RETRY) + ": " + strconv.FormatInt(retry.Milliseconds(), 10) + "\n")
	return err
}

func validateRetry(data string) error {
	if strings.ContainsAny(data, "\r\n") {
		return ErrInvalidRetry
	}
	retry, err := strconv.ParseInt(data, 10, 64)
	if err != nil || retry < 0 {
		return ErrInvalidRetry
	}
	return nil
}

func writeComment(w io.StringWriter, comment string) error {
	value := strings.ReplaceAll(comment, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	for {
		line, rest, ok := strings.Cut(value, "\n")
		if _, err := w.WriteString(": " + line + "\n"); err != nil {
			return err
		}
		if !ok {
			return nil
		}
		value = rest
	}
}
