package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestPathParamReferencesCleared(t *testing.T) {
	// These subtests are serial: a fresh pool makes the first request use the
	// inline buffer, even if another test previously grew a pooled Context.
	for _, count := range []int{8, 9, 17, 65} {
		for _, outcome := range []string{"success", "panic", "404", "405", "fallback"} {
			t.Run(fmt.Sprintf("params%d/%s", count, outcome), func(t *testing.T) {
				var used *Context
				contextPool = sync.Pool{New: func() any {
					c := new(Context)
					used = c
					return c
				}}
				t.Cleanup(func() {
					contextPool = sync.Pool{New: func() any { return new(Context) }}
				})

				g := New()
				g.OnPanic(func(c *Context, _ any) { c.Writer.WriteHeader(http.StatusInternalServerError) })
				pattern, path := numberedPath("p", count), "/"+numberedValues(count)
				method, status := GET, http.StatusOK
				if outcome == "404" || outcome == "fallback" {
					pattern += "/only"
					path += "/missing"
					status = http.StatusNotFound
				}
				if outcome == "405" {
					method, status = http.MethodPost, http.StatusMethodNotAllowed
				}
				if outcome == "panic" {
					status = http.StatusInternalServerError
				}
				g.GET(pattern, func(c *Context) {
					if outcome == "panic" {
						panic("parameter cleanup")
					}
				})
				if outcome == "fallback" {
					status = http.StatusOK
					g.GET("/#...", func(c *Context) {
						if got := c.GetUrlPathParam("p0"); got != "" {
							t.Errorf("failed branch leaked p0=%q", got)
						}
					})
				}

				rw := httptest.NewRecorder()
				g.ServeHTTP(rw, httptest.NewRequest(method, path, nil))
				if rw.Code != status {
					t.Fatalf("status = %d, want %d", rw.Code, status)
				}
				if len(used.params) != 0 {
					t.Errorf("parameters survived request: %v", used.params)
				}
				for name, buffer := range map[string][]pathParam{
					"inline":   used.paramBuf[:],
					"expanded": used.params[:cap(used.params)],
				} {
					for i, param := range buffer {
						if param != (pathParam{}) {
							t.Errorf("%s buffer[%d] retains %+v", name, i, param)
						}
					}
				}
				if count > maxContextParamCap {
					if used.params != nil {
						t.Error("oversized parameter buffer was retained")
					}
					previous := used
					if next := contextPool.Get().(*Context); next == previous {
						t.Error("oversized Context was returned to the pool")
					}
				}
			})
		}
	}
}

func TestDuplicatePathParamsUseLastValue(t *testing.T) {
	g := New()
	var first, second map[string]string
	var got string
	g.GET("/#id/#id", func(c *Context) {
		got = c.GetUrlPathParam("id")
		first = c.AllUrlPathParam()
		first["id"] = "mutated"
		second = c.AllUrlPathParam()
	})

	runRequest(t, g, "GET", "/first/second")
	if got != "second" {
		t.Fatalf("GetUrlPathParam = %q, want second", got)
	}
	if first["id"] != "mutated" || second["id"] != "second" {
		t.Fatalf("map ownership broken: first=%v second=%v", first, second)
	}
}

func TestFailedParameterBranchClearsValuesBeforeFallback(t *testing.T) {
	g := New()
	g.GET("/#id/only", func(c *Context) {
		t.Fatalf("failed parameter branch should not run")
	})
	g.GET("/#...", func(c *Context) {
		if got := c.GetUrlPathParam("id"); got != "" {
			t.Errorf("stale id parameter = %q", got)
		}
		if got := c.GetUrlPathParam("#"); got != "x" {
			t.Errorf("catch-all = %q, want x", got)
		}
		if got := len(c.params); got != 1 {
			t.Errorf("internal parameter count = %d, want 1", got)
		}
	})

	runRequest(t, g, "GET", "/x")
}

func TestDynamicMatchingNormalizesContinuousSlashes(t *testing.T) {
	g := New()
	g.GET("/files/#...", func(c *Context) {
		if got := c.GetUrlPathParam("#"); got != "a/b" {
			t.Errorf("catch-all = %q, want a/b", got)
		}
	})

	runRequest(t, g, "GET", "///files///a//b///")
}

func numberedPath(prefix string, count int) string {
	var b strings.Builder
	for i := 0; i < count; i++ {
		b.WriteString("/#")
		b.WriteString(prefix)
		b.WriteString(strconv.Itoa(i))
	}
	return b.String()
}

func numberedValues(count int) string {
	var b strings.Builder
	for i := 0; i < count; i++ {
		if i > 0 {
			b.WriteByte('/')
		}
		b.WriteString("v")
		b.WriteString(strconv.Itoa(i))
	}
	return b.String()
}

func TestPathParamBuffersAndPoolReset(t *testing.T) {
	const count = 9
	g := New()
	path := numberedPath("p", count)
	values := "/" + numberedValues(count)
	g.GET(path, func(c *Context) {
		if len(c.params) != count {
			t.Errorf("parameter count = %d, want %d", len(c.params), count)
		}
		if got := c.GetUrlPathParam("p8"); got != "v8" {
			t.Errorf("p8 = %q", got)
		}
	})
	g.GET("/static", func(c *Context) {
		if len(c.params) != 0 {
			t.Errorf("parameters leaked into static request: %v", c.params)
		}
	})

	runRequest(t, g, "GET", values)
	runRequest(t, g, "GET", "/static")
}

func TestMoreThan64PathParamsAreHandled(t *testing.T) {
	const count = 65
	g := New()
	g.GET(numberedPath("p", count), func(c *Context) {
		if len(c.params) != count {
			t.Errorf("parameter count = %d, want %d", len(c.params), count)
		}
		if got := c.GetUrlPathParam("p64"); got != "v64" {
			t.Errorf("p64 = %q", got)
		}
		if cap(c.params) <= maxContextParamCap {
			t.Errorf("parameter buffer cap = %d, want > %d", cap(c.params), maxContextParamCap)
		}
	})

	runRequest(t, g, "GET", "/"+numberedValues(count))
}

func TestPathMatchReturnsIndependentMaps(t *testing.T) {
	g := New()
	g.GET("/#id/#id", func(c *Context) {})

	first, _, _ := g.PathMatch("/a/b", GET)
	first["id"] = "changed"
	second, _, _ := g.PathMatch("/a/b", GET)
	if second["id"] != "b" {
		t.Fatalf("PathMatch map was shared: %v", second)
	}
}

func TestSnapshotChainsAreMethodSpecificAndIndependent(t *testing.T) {
	g := New()
	g.GET("/x", func(c *Context) {})
	g.HeadMiddleware(func(c *Context) {})
	g.POST("/x", func(c *Context) {})

	oldSnapshot := g.host.snapshot.Load()
	oldLeaf := oldSnapshot.flatRoutes["x"]
	getChain := oldLeaf.handlerChains[GET]
	postChain := oldLeaf.handlerChains["POST"]
	if len(getChain) != 1 || len(postChain) != 2 {
		t.Fatalf("unexpected chain lengths: GET=%d POST=%d", len(getChain), len(postChain))
	}

	g.GET("/y", func(c *Context) {})
	newSnapshot := g.host.snapshot.Load()
	newLeaf := newSnapshot.flatRoutes["x"]
	newGetChain := newLeaf.handlerChains[GET]
	if &getChain[0] == &newGetChain[0] {
		t.Fatal("different snapshots share a mutable chain backing array")
	}
}

func TestInFlightRequestKeepsOldHandlerChain(t *testing.T) {
	g := New()
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	var middlewareCalls int
	g.GET("/x", func(c *Context) {
		once.Do(func() { close(started) })
		<-release
	})

	done := make(chan struct{})
	go func() {
		runRequest(t, g, "GET", "/x")
		close(done)
	}()
	<-started
	g.HeadMiddleware(func(c *Context) { middlewareCalls++ })
	g.GET("/y", func(c *Context) {})
	close(release)
	<-done
	if middlewareCalls != 0 {
		t.Fatalf("in-flight request used new middleware chain: %d", middlewareCalls)
	}

	runRequest(t, g, "GET", "/y")
	if middlewareCalls != 1 {
		t.Fatalf("new request did not use new chain: %d", middlewareCalls)
	}
}
