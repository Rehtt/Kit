package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNetworkErrorPanicWrites500(t *testing.T) {
	for _, failure := range []error{
		context.DeadlineExceeded,
		fmt.Errorf("backend failed: %w", context.DeadlineExceeded),
		&net.OpError{Op: "dial", Net: "tcp", Err: context.DeadlineExceeded},
	} {
		t.Run(failure.Error(), func(t *testing.T) {
			g := New()
			g.GET("/", func(c *Context) { panic(failure) })
			rec := httptest.NewRecorder()
			g.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
			if rec.Code != http.StatusInternalServerError || rec.Body.String() != "Internal Server Error\n" {
				t.Fatalf("status=%d body=%q", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestPanicAbortCleansUpContext(t *testing.T) {
	for _, tc := range []struct {
		name  string
		write bool
		hook  func(*Context, any)
	}{
		{name: "committed default", write: true},
		{name: "committed custom", write: true, hook: func(*Context, any) {}},
		{name: "hook requests abort", hook: func(*Context, any) { panic(http.ErrAbortHandler) }},
		{name: "hook fails", hook: func(*Context, any) { panic("hook failed") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := New(WithOnPanic(tc.hook))
			var requestContext context.Context
			g.GET("/", func(c *Context) {
				requestContext = c.Context
				if tc.write {
					c.WriteString("partial")
				}
				panic("handler failed")
			})
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				g.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
			}()
			if recovered != http.ErrAbortHandler {
				t.Fatalf("panic=%v, want ErrAbortHandler", recovered)
			}
			if requestContext == nil || requestContext.Err() != context.Canceled {
				t.Fatal("aborted request context was not canceled")
			}
		})
	}
}
