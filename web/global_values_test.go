package web

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestConcurrentGlobalValues(t *testing.T) {
	g := New()
	g.SetValue("config", 0)
	g.GET("/", func(c *Context) {
		if _, ok := c.GetContextValue("config").(int); !ok {
			t.Error("request lost its global value")
		}
	})
	const workers, updates = 4, 200
	start := make(chan struct{})
	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(key int) {
			defer wg.Done()
			<-start
			for i := 0; i < updates; i++ {
				g.SetValue(key, i)
				g.SetValue("config", i)
				if got := g.GetValue(key); got != i {
					t.Errorf("worker %d: got %v, want %d", key, got, i)
				}
				_ = g.Value("config")
				_, _ = g.Deadline()
				_ = g.Done()
				_ = g.Err()
				g.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
			}
		}(worker)
	}
	close(start)
	wg.Wait()
	for worker := 0; worker < workers; worker++ {
		if got := g.GetValue(worker); got != updates-1 {
			t.Errorf("lost update for worker %d: %v", worker, got)
		}
	}
}

func TestGlobalValueSnapshotAndContextMethods(t *testing.T) {
	type key string
	parent, cancel := context.WithDeadline(context.WithValue(context.Background(), key("base"), "value"), time.Now().Add(time.Hour))
	defer cancel()
	g := New(WithContext(parent))
	g.SetValue(key("config"), "before")
	g.GET("/", func(c *Context) {
		g.SetValue(key("config"), "after")
		if got := c.Value(key("config")); got != "before" {
			t.Errorf("in-flight value changed: %v", got)
		}
		if got := c.Value(key("base")); got != "value" {
			t.Errorf("WithContext value lost: %v", got)
		}
		c.SetContextValue(key("config"), "local")
		if got := c.Value(key("config")); got != "local" {
			t.Errorf("request override lost: %v", got)
		}
	})
	g.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if got := g.GetValue(key("config")); got != "after" {
		t.Fatalf("global value=%v", got)
	}
	wantDeadline, _ := parent.Deadline()
	if deadline, ok := g.Deadline(); !ok || !deadline.Equal(wantDeadline) {
		t.Fatal("global context deadline changed")
	}
	if g.Done() != parent.Done() || g.Err() != nil {
		t.Fatal("global context cancellation changed")
	}
	cancel()
	if g.Err() != context.Canceled {
		t.Fatal("global context no longer follows its parent")
	}
}
