package web

import (
	"context"
	"fmt"
	"math"
	"net/http/httptest"
	"reflect"
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

func TestGlobalMapCompatibility(t *testing.T) {
	g := New()
	g.Context = context.WithValue(context.Background(), "nil", "parent")
	g.SetValue("nil", nil)
	g.SetValue("key", 0)
	old := g.globalContext()
	for i := 1; i <= 1000; i++ {
		g.SetValue("key", i)
	}
	if old.Value("key") != 0 || g.Value("key") != 1000 || g.Value("missing") != nil || g.Value("nil") != nil {
		t.Fatal("snapshot values changed")
	}
	snapshot := g.Context.(*globalValues)
	if len(snapshot.values) != 2 {
		t.Fatal("retained update history")
	}
	if _, ok := snapshot.Context.(*globalValues); ok {
		t.Fatal("retained old snapshot")
	}
	// External wrappers must remain visible and must not be flattened.
	g.Context = context.WithValue(g.Context, "key", "external")
	g.SetValue("next", 1)
	if g.Value("key") != "external" {
		t.Fatal("external wrapper lost")
	}
}

func TestGlobalSpecialKeys(t *testing.T) {
	type boxed struct{ X any }
	for _, key := range []any{nil, []int{1}, boxed{[]int{1}}} {
		t.Run(fmt.Sprintf("%T", key), func(t *testing.T) {
			g := New()
			if key == nil || !reflect.TypeOf(key).Comparable() {
				defer func() {
					if recover() == nil {
						t.Fatal("invalid key accepted")
					}
				}()
				g.SetValue(key, 1)
			} else {
				// WithValue accepts this key; equality can panic only when queried.
				g.SetValue(key, 1)
				defer func() {
					if recover() == nil {
						t.Fatal("standard equality panic lost")
					}
				}()
				_ = g.Value(key)
			}
		})
	}
	for _, key := range []any{math.NaN(), complex(math.NaN(), 0), boxed{math.NaN()}} {
		g := New()
		g.SetValue("ordinary", 1)
		g.SetValue(key, 2)
		g.SetValue("after", 3)
		if g.Value(key) != nil || g.Value("ordinary") != 1 || g.Value("after") != 3 {
			t.Fatal("non-reflexive key semantics changed")
		}
	}
	var pointer *int
	g := New()
	g.SetValue(pointer, 42)
	if g.Value(pointer) != 42 {
		t.Fatal("typed nil key lost")
	}
}
