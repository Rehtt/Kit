package web

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequestCancellationLifecycle(t *testing.T) {
	for _, mode := range []string{"normal", "done", "stop", "between", "pre-canceled", "deadline", "panic"} {
		t.Run(mode, func(t *testing.T) {
			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			if mode == "pre-canceled" {
				cancel()
			}
			if mode == "deadline" {
				var stop context.CancelFunc
				parent, stop = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
				defer stop()
			}
			var captured context.Context
			var done <-chan struct{}
			calls := 0
			g := New()
			g.OnPanic(func(*Context, any) {})
			g.Middlewares(func(c *Context) {
				captured = c.Context
				if mode == "done" {
					done = c.Done()
					select {
					case <-done:
						t.Error("premature cancellation")
					default:
					}
				}
				if mode == "stop" {
					c.Stop()
				}
				if mode == "between" {
					cancel()
				}
				if mode == "panic" {
					panic("test")
				}
			})
			g.GET("/", func(c *Context) { calls++ })
			g.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil).WithContext(parent))
			want := 0
			if mode == "normal" || mode == "done" {
				want = 1
			}
			if calls != want {
				t.Fatalf("calls=%d want %d", calls, want)
			}
			if captured != nil && captured.Err() != context.Canceled {
				t.Fatal("request context not cleaned up")
			}
			if done != nil {
				select {
				case <-done:
				default:
					t.Fatal("Done not closed after return")
				}
			}
		})
	}
}
