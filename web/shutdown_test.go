package web

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestShutdownTimeoutOption(t *testing.T) {
	if got := New().shutdownTimeout; got != 30*time.Second {
		t.Fatalf("default = %v", got)
	}
	if got := New(WithShutdownTimeout(time.Second), WithShutdownTimeout(2*time.Second)).shutdownTimeout; got != 2*time.Second {
		t.Fatalf("override = %v", got)
	}
	for _, d := range []time.Duration{0, -time.Second} {
		t.Run(d.String(), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			New(WithShutdownTimeout(d))
		})
	}
}

func TestRunContextActiveRequest(t *testing.T) {
	for _, mode := range []string{"cancel", "deadline", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			entered, release, stopping := make(chan struct{}), make(chan struct{}), make(chan struct{})
			timeout := 2 * time.Second
			if mode == "timeout" {
				timeout = 50 * time.Millisecond
			}
			g := New(WithShutdownTimeout(timeout))
			g.Server.RegisterOnShutdown(func() { close(stopping) })
			g.GET("/", func(c *Context) { close(entered); <-release; _, _ = c.Writer.Write([]byte("done")) })
			// BaseContext exposes the bound listener address before requests are accepted.
			address := make(chan string, 1)
			g.Server.BaseContext = func(l net.Listener) context.Context { address <- l.Addr().String(); return context.Background() }
			ctx, cancel := context.WithCancel(context.Background())
			if mode == "deadline" {
				cancel()
				ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			}
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- g.RunContext(ctx, "127.0.0.1:0") }()
			released := false
			defer func() {
				if !released {
					close(release)
				}
				_ = g.Server.Close()
			}()
			var addr string
			select {
			case addr = <-address:
			case <-time.After(3 * time.Second):
				t.Fatal("listen timeout")
			}
			client := &http.Client{Timeout: 5 * time.Second}
			response := make(chan error, 1)
			go func() {
				res, err := client.Get("http://" + addr + "/")
				if err == nil {
					body, readErr := io.ReadAll(res.Body)
					res.Body.Close()
					err = readErr
					if err == nil && string(body) != "done" {
						err = errors.New("response body mismatch")
					}
				}
				response <- err
			}()
			select {
			case <-entered:
			case <-time.After(3 * time.Second):
				t.Fatal("request did not enter")
			}
			if mode != "deadline" {
				cancel()
			}
			select {
			case <-stopping:
			case <-time.After(3 * time.Second):
				t.Fatal("shutdown did not start")
			}
			if mode == "timeout" {
				select {
				case err := <-done:
					if !errors.Is(err, context.DeadlineExceeded) {
						t.Fatalf("shutdown = %v", err)
					}
				case <-time.After(3 * time.Second):
					t.Fatal("shutdown did not time out")
				}
			}
			close(release)
			released = true
			select {
			case err := <-response:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("response timeout")
			}
			if mode != "timeout" {
				select {
				case err := <-done:
					if !errors.Is(err, http.ErrServerClosed) {
						t.Fatalf("shutdown = %v", err)
					}
				case <-time.After(3 * time.Second):
					t.Fatal("shutdown timeout")
				}
			}
		})
	}
}
