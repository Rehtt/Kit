package web

import (
	"net/http/httptest"
	"testing"
)

func benchHandler(ctx *Context) {}

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

// 静态命中但需要走 radix（动态环境）。
func BenchmarkRadixWalk(b *testing.B) {
	g := benchSetup(b, func(g *GOweb) {
		g.GET("/api/v1/users", benchHandler)
		g.GET("/api/v1/orders", benchHandler)
		g.GET("/api/v2/items", benchHandler)
		g.GET("/u/#id", benchHandler) // 引入 dynamic 让 fast path 失败
	})
	req := httptest.NewRequest("GET", "/api/v1/users", nil)
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
