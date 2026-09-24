package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func benchHandler(ctx *Context) {}

var benchParamSink string

func benchParamHandler(ctx *Context) {
	benchParamSink = ctx.GetUrlPathParam("id")
}

func benchCatchAllHandler(ctx *Context) {
	benchParamSink = ctx.GetUrlPathParam("#")
}

type reusableResponseWriter struct {
	header http.Header
	status int
	size   int
}

func (w *reusableResponseWriter) Header() http.Header { return w.header }

func (w *reusableResponseWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}

func (w *reusableResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.size += len(p)
	return len(p), nil
}

func (w *reusableResponseWriter) reset() {
	w.status = 0
	w.size = 0
	for key := range w.header {
		delete(w.header, key)
	}
}

func benchSetup(b *testing.B, register func(g *GOweb)) *GOweb {
	b.Helper()
	g := New()
	register(g)
	return g
}

// 纯静态路径走 flatRoutes 快速通道；理想情况下 0 分配（不构建 params）。
func BenchmarkStaticFastPath(b *testing.B) {
	g := benchSetup(b, func(g *GOweb) {
		g.GET("/api/v1/users", benchHandler)
		g.GET("/api/v1/orders", benchHandler)
		g.GET("/api/v2/items", benchHandler)
	})
	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		g.ServeHTTP(rec, req)
	}
}

// 动态路径命中 radix；请求不会进入静态 flatRoutes 快速通道。
func BenchmarkRadixWalk(b *testing.B) {
	g := benchSetup(b, func(g *GOweb) {
		g.GET("/u/#id/profile", benchHandler)
		g.GET("/u/#id/orders", benchHandler)
	})
	req := httptest.NewRequest("GET", "/u/12345/profile", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		g.ServeHTTP(rec, req)
	}
}

// 命中 #name 参数。
func BenchmarkParam(b *testing.B) {
	g := benchSetup(b, func(g *GOweb) {
		g.GET("/u/#id", benchHandler)
	})
	req := httptest.NewRequest("GET", "/u/12345", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		g.ServeHTTP(rec, req)
	}
}

// 命中 #... catchAll。
func BenchmarkCatchAll(b *testing.B) {
	g := benchSetup(b, func(g *GOweb) {
		g.GET("/files/#...", benchHandler)
	})
	req := httptest.NewRequest("GET", "/files/a/b/c/d/e", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		g.ServeHTTP(rec, req)
	}
}

// 直接调用 snapshot.match，剥离 ServeHTTP 的开销以观察纯匹配成本。
func BenchmarkSnapshotMatchStatic(b *testing.B) {
	g := benchSetup(b, func(g *GOweb) {
		g.GET("/api/v1/users", benchHandler)
		g.GET("/api/v1/orders", benchHandler)
	})
	snap := g.RouterGroup.host.snapshot.Load()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = snap.match("/api/v1/users", "GET")
	}
}

func BenchmarkSnapshotMatchParam(b *testing.B) {
	g := benchSetup(b, func(g *GOweb) {
		g.GET("/u/#id", benchHandler)
	})
	snap := g.RouterGroup.host.snapshot.Load()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = snap.match("/u/123", "GET")
	}
}

func BenchmarkSnapshotMatchCatchAll(b *testing.B) {
	g := benchSetup(b, func(g *GOweb) {
		g.GET("/files/#...", benchHandler)
	})
	snap := g.RouterGroup.host.snapshot.Load()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = snap.match("/files/a/b/c/d", "GET")
	}
}

// These benchmarks remove ResponseRecorder construction and exercise the
// request path with the parameter API actually used by handlers.
func BenchmarkServeHTTPStaticReusable(b *testing.B) {
	g := benchSetup(b, func(g *GOweb) { g.GET("/api/v1/users", benchHandler) })
	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	rw := &reusableResponseWriter{header: make(http.Header)}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rw.reset()
		g.ServeHTTP(rw, req)
	}
}

func BenchmarkServeHTTPParamReusable(b *testing.B) {
	g := benchSetup(b, func(g *GOweb) { g.GET("/u/#id", benchParamHandler) })
	req := httptest.NewRequest("GET", "/u/12345", nil)
	rw := &reusableResponseWriter{header: make(http.Header)}
	benchParamSink = ""
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rw.reset()
		g.ServeHTTP(rw, req)
	}
	if benchParamSink != "12345" {
		b.Fatalf("parameter = %q, want 12345", benchParamSink)
	}
}

func BenchmarkServeHTTPCatchAllReusable(b *testing.B) {
	g := benchSetup(b, func(g *GOweb) { g.GET("/files/#...", benchCatchAllHandler) })
	req := httptest.NewRequest("GET", "/files/a/b/c/d", nil)
	rw := &reusableResponseWriter{header: make(http.Header)}
	benchParamSink = ""
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rw.reset()
		g.ServeHTTP(rw, req)
	}
	if benchParamSink != "a/b/c/d" {
		b.Fatalf("catch-all parameter = %q, want a/b/c/d", benchParamSink)
	}
}

func BenchmarkServeHTTPMatrix(b *testing.B) {
	for _, tc := range []struct {
		name, pattern, method, path, key, value string
		status                                  int
	}{
		{"static", "/api/v1/users", GET, "/api/v1/users", "", "", http.StatusOK},
		{"param", "/u/#id", GET, "/u/12345", "id", "12345", http.StatusOK},
		{"catchAll", "/files/#...", GET, "/files/a/b/c/d", "#", "a/b/c/d", http.StatusOK},
		{"backtrack", "/u/#id/profile", GET, "/u/static/profile", "id", "static", http.StatusOK},
		{"404", "/u/#id/profile", GET, "/u/12345/missing", "", "", http.StatusNotFound},
		{"405", "/u/#id/profile", http.MethodPost, "/u/12345/profile", "", "", http.StatusMethodNotAllowed},
	} {
		for _, middlewareCount := range []int{0, 5, 20} {
			for _, parallel := range []bool{false, true} {
				mode := "serial"
				if parallel {
					mode = "parallel"
				}
				b.Run(fmt.Sprintf("%s/middleware%d/%s", tc.name, middlewareCount, mode), func(b *testing.B) {
					g := New(WithRoutes(func(r *RouterGroup) {
						for i := 0; i < middlewareCount; i++ {
							r.Middlewares(benchHandler)
						}
						if tc.name == "backtrack" {
							r.GET("/u/static/only", benchHandler)
						}
						r.GET(tc.pattern, func(c *Context) {
							if tc.key != "" {
								if got := c.GetUrlPathParam(tc.key); got != tc.value {
									b.Errorf("parameter %q = %q, want %q", tc.key, got, tc.value)
								}
							}
						})
					}))
					req := httptest.NewRequest(tc.method, tc.path, nil)
					rw := &reusableResponseWriter{header: make(http.Header)}
					g.ServeHTTP(rw, req) // Warm the pool and validate the selected route.
					status := rw.status
					if status == 0 {
						status = http.StatusOK
					}
					if status != tc.status {
						b.Fatalf("status = %d, want %d", status, tc.status)
					}
					b.ReportAllocs()
					b.ResetTimer()
					if parallel {
						b.RunParallel(func(pb *testing.PB) {
							// Writers are worker-local; the request and route snapshot are read-only.
							rw := &reusableResponseWriter{header: make(http.Header)}
							for pb.Next() {
								rw.reset()
								g.ServeHTTP(rw, req)
							}
						})
					} else {
						for i := 0; i < b.N; i++ {
							rw.reset()
							g.ServeHTTP(rw, req)
						}
					}
				})
			}
		}
	}
}

func BenchmarkSnapshotBuild(b *testing.B) {
	for _, routeCount := range []int{100, 1000} {
		for _, middlewareCount := range []int{0, 5, 20} {
			name := fmt.Sprintf("routes%d/middleware%d", routeCount, middlewareCount)
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					_ = New(WithRoutes(func(r *RouterGroup) {
						for j := 0; j < middlewareCount; j++ {
							r.Middlewares(benchHandler)
						}
						for j := 0; j < routeCount; j++ {
							r.GET("/r"+strconv.Itoa(j), benchHandler)
						}
					}))
				}
			})
		}
	}
}
