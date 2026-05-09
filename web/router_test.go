package web

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func runRequest(t *testing.T, g *GOweb, method, path string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	g.ServeHTTP(rec, req)
	return rec.Body.String()
}

func TestStaticRoute(t *testing.T) {
	g := New()
	g.GET("/api/users", func(ctx *Context) { ctx.Writer.Write([]byte("users")) })
	if got := runRequest(t, g, "GET", "/api/users"); got != "users" {
		t.Fatalf("got %q", got)
	}
}

func TestRootRoute(t *testing.T) {
	g := New()
	g.GET("/", func(ctx *Context) { ctx.Writer.Write([]byte("root")) })
	if got := runRequest(t, g, "GET", "/"); got != "root" {
		t.Fatalf("got %q", got)
	}
}

func TestParamRoute(t *testing.T) {
	g := New()
	g.GET("/api/#name/info", func(ctx *Context) {
		ctx.Writer.Write([]byte(ctx.GetUrlPathParam("name")))
	})
	if got := runRequest(t, g, "GET", "/api/alice/info"); got != "alice" {
		t.Fatalf("got %q", got)
	}
	if got := runRequest(t, g, "GET", "/api/alice"); got != "404 page not found\n" {
		t.Fatalf("expected 404, got %q", got)
	}
}

func TestCatchAllRoute(t *testing.T) {
	g := New()
	g.GET("/files/#...", func(ctx *Context) {
		ctx.Writer.Write([]byte(ctx.GetUrlPathParam("#")))
	})
	if got := runRequest(t, g, "GET", "/files/a/b/c"); got != "a/b/c" {
		t.Fatalf("got %q", got)
	}
}

func TestEdgeSplitOnLaterRegistration(t *testing.T) {
	g := New()
	g.GET("/api/v1/users", func(ctx *Context) { ctx.Writer.Write([]byte("u1")) })
	g.GET("/api/v1/orders", func(ctx *Context) { ctx.Writer.Write([]byte("o1")) })
	g.GET("/api/v2", func(ctx *Context) { ctx.Writer.Write([]byte("v2")) })
	g.GET("/api", func(ctx *Context) { ctx.Writer.Write([]byte("api")) })

	cases := map[string]string{
		"/api/v1/users":  "u1",
		"/api/v1/orders": "o1",
		"/api/v2":        "v2",
		"/api":           "api",
	}
	for path, want := range cases {
		if got := runRequest(t, g, "GET", path); got != want {
			t.Fatalf("path %s want %q got %q", path, want, got)
		}
	}
}

func TestGrepInsideCompressedEdgeSplits(t *testing.T) {
	// /api/v1/users 先注册形成压缩边 ["api","v1","users"]，再 Grep("/api")
	// 必须能从压缩边中分裂出独立的 /api 节点，并保留原路由可达。
	g := New()
	var seen []string
	g.GET("/api/v1/users", func(ctx *Context) { seen = append(seen, "u") })
	api := g.Grep("/api")
	api.HeadMiddleware(func(ctx *Context) { seen = append(seen, "mw") })
	api.GET("/v2", func(ctx *Context) { seen = append(seen, "v2") })

	// /api/v1/users 注册早于中间件 -> 不触发 mw（保持原 order 语义）
	seen = nil
	runRequest(t, g, "GET", "/api/v1/users")
	if !equalStrings(seen, []string{"u"}) {
		t.Fatalf("/api/v1/users got %v", seen)
	}
	// /api/v2 注册晚于中间件 -> 应触发 mw，证明分裂出的 /api 节点上的中间件确实生效
	seen = nil
	runRequest(t, g, "GET", "/api/v2")
	if !equalStrings(seen, []string{"mw", "v2"}) {
		t.Fatalf("/api/v2 got %v", seen)
	}
}

func TestGrepMiddlewareOrderOnlyAfter(t *testing.T) {
	g := New()
	api := g.Grep("/api")
	var calls []string
	api.GET("/before", func(ctx *Context) { calls = append(calls, "before") })
	api.HeadMiddleware(func(ctx *Context) { calls = append(calls, "mw") })
	api.GET("/after", func(ctx *Context) { calls = append(calls, "after") })

	calls = nil
	runRequest(t, g, "GET", "/api/before")
	if !equalStrings(calls, []string{"before"}) {
		t.Fatalf("/before: got %v", calls)
	}

	calls = nil
	runRequest(t, g, "GET", "/api/after")
	if !equalStrings(calls, []string{"mw", "after"}) {
		t.Fatalf("/after: got %v", calls)
	}
}

func TestFootMiddleware(t *testing.T) {
	g := New()
	var calls []string
	g.FootMiddleware(func(ctx *Context) { calls = append(calls, "foot") })
	g.GET("/x", func(ctx *Context) { calls = append(calls, "h") })
	runRequest(t, g, "GET", "/x")
	if !equalStrings(calls, []string{"h", "foot"}) {
		t.Fatalf("got %v", calls)
	}
}

func TestList(t *testing.T) {
	g := New()
	g.GET("/", func(ctx *Context) {})
	g.GET("/api", func(ctx *Context) {})
	g.GET("/api/users", func(ctx *Context) {})
	g.POST("/api/users", func(ctx *Context) {})
	g.GET("/files/#...", func(ctx *Context) {})

	methods, paths := g.List()
	pairs := make([]string, len(methods))
	for i := range methods {
		pairs[i] = methods[i] + " " + paths[i]
	}
	sort.Strings(pairs)
	want := []string{
		"GET /",
		"GET /api",
		"GET /api/users",
		"GET /files/#...",
		"POST /api/users",
	}
	if !equalStrings(pairs, want) {
		t.Fatalf("got %v\nwant %v", pairs, want)
	}
}

func TestStaticOnlyUsesFlatRoutes(t *testing.T) {
	g := New()
	g.GET("/a", func(ctx *Context) { ctx.Writer.Write([]byte("a")) })
	g.GET("/api/users", func(ctx *Context) { ctx.Writer.Write([]byte("au")) })
	g.GET("/api/orders", func(ctx *Context) { ctx.Writer.Write([]byte("ao")) })

	snap := g.RouterGroup.host.snapshot.Load()
	if snap == nil {
		t.Fatalf("expected snapshot to be published")
	}
	if snap.hasDynamic {
		t.Fatalf("expected hasDynamic=false for purely static routes")
	}
	if got := len(snap.flatRoutes); got != 3 {
		t.Fatalf("expected 3 flat entries, got %d", got)
	}
	if got := runRequest(t, g, "GET", "/a"); got != "a" {
		t.Fatalf("got %q", got)
	}
	if got := runRequest(t, g, "GET", "/api/users"); got != "au" {
		t.Fatalf("got %q", got)
	}
	if got := runRequest(t, g, "GET", "/missing"); !strings.Contains(got, "404") {
		t.Fatalf("expected 404, got %q", got)
	}
}

func TestStaticPriorityOverDynamic(t *testing.T) {
	g := New()
	g.GET("/users", func(ctx *Context) { ctx.Writer.Write([]byte("static")) })
	g.GET("/#name", func(ctx *Context) { ctx.Writer.Write([]byte("dyn:" + ctx.GetUrlPathParam("name"))) })
	snap := g.RouterGroup.host.snapshot.Load()
	if !snap.hasDynamic {
		t.Fatalf("expected hasDynamic=true")
	}
	if got := runRequest(t, g, "GET", "/users"); got != "static" {
		t.Fatalf("got %q", got)
	}
	if got := runRequest(t, g, "GET", "/foo"); got != "dyn:foo" {
		t.Fatalf("got %q", got)
	}
}

func TestRootCatchAllFallback(t *testing.T) {
	g := New()
	g.GET("/api/test", func(ctx *Context) { ctx.Writer.Write([]byte("api")) })
	g.GET("/#...", func(ctx *Context) { ctx.Writer.Write([]byte("fb:" + ctx.GetUrlPathParam("#"))) })

	if got := runRequest(t, g, "GET", "/api/test"); got != "api" {
		t.Fatalf("got %q", got)
	}
	if got := runRequest(t, g, "GET", "/foo/bar"); got != "fb:foo/bar" {
		t.Fatalf("got %q", got)
	}
	if got := runRequest(t, g, "GET", "/api"); got != "fb:api" {
		t.Fatalf("got %q", got)
	}
}

func TestQueryStringStripped(t *testing.T) {
	g := New()
	g.GET("/x", func(ctx *Context) { ctx.Writer.Write([]byte("ok")) })
	if got := runRequest(t, g, "GET", "/x?a=1&b=2"); got != "ok" {
		t.Fatalf("got %q", got)
	}
}

func TestAnyHandler(t *testing.T) {
	g := New()
	g.Any("/a", func(ctx *Context) { ctx.Writer.Write([]byte(ctx.Request.Method)) })
	for _, m := range []string{"GET", "POST", "PUT", "DELETE"} {
		if got := runRequest(t, g, m, "/a"); got != m {
			t.Fatalf("method %s got %q", m, got)
		}
	}
}

func TestMethodConflict(t *testing.T) {
	g := New()
	g.GET("/x", func(ctx *Context) {})
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on duplicate GET")
		}
	}()
	g.GET("/x", func(ctx *Context) {})
}

func TestAnyConflict(t *testing.T) {
	g := New()
	g.Any("/x", func(ctx *Context) {})
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on GET after Any")
		}
	}()
	g.GET("/x", func(ctx *Context) {})
}

func TestParamNameConflict(t *testing.T) {
	g := New()
	g.GET("/x/#a", func(ctx *Context) {})
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on conflicting param name")
		}
	}()
	g.GET("/x/#b", func(ctx *Context) {})
}

func TestCatchAllNotAtEnd(t *testing.T) {
	g := New()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when #... is not last")
		}
	}()
	g.GET("/x/#.../y", func(ctx *Context) {})
}

func TestCompletePathAfterSplit(t *testing.T) {
	g := New()
	g.GET("/a/b/c", func(ctx *Context) {})
	g.GET("/a/b/d", func(ctx *Context) {})
	leaves := g.RouterGroup.BottomNodeList()
	got := make([]string, 0, len(leaves))
	for _, n := range leaves {
		got = append(got, n.completePath())
	}
	sort.Strings(got)
	want := []string{"/a/b/c", "/a/b/d"}
	if !equalStrings(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestSplitPreservesExistingFlatRoute(t *testing.T) {
	g := New()
	g.GET("/api/v1/users", func(ctx *Context) { ctx.Writer.Write([]byte("u1")) })
	g.GET("/api/v2", func(ctx *Context) { ctx.Writer.Write([]byte("v2")) })
	if got := runRequest(t, g, "GET", "/api/v1/users"); got != "u1" {
		t.Fatalf("got %q", got)
	}
	if got := runRequest(t, g, "GET", "/api/v2"); got != "v2" {
		t.Fatalf("got %q", got)
	}
}

func TestNoRoute(t *testing.T) {
	g := New()
	g.NoRoute(func(ctx *Context) { ctx.Writer.Write([]byte("custom-404")) })
	g.GET("/a", func(ctx *Context) { ctx.Writer.Write([]byte("a")) })
	if got := runRequest(t, g, "GET", "/missing"); got != "custom-404" {
		t.Fatalf("got %q", got)
	}
	if got := runRequest(t, g, "GET", "/a"); got != "a" {
		t.Fatalf("got %q", got)
	}
}

// 回归: Grep 不应改写已注册路由的 order，
// 否则在「先注册路由 / 再注册中间件 / 再 Grep 同一路径」时
// 会误把中间件应用到注册时间更早的路由上。
func TestGrepDoesNotBumpOrderOfExistingRoute(t *testing.T) {
	g := New()
	var calls []string
	g.GET("/api", func(ctx *Context) { calls = append(calls, "h") })
	g.HeadMiddleware(func(ctx *Context) { calls = append(calls, "mw") })
	_ = g.Grep("/api")

	calls = nil
	runRequest(t, g, "GET", "/api")
	if !equalStrings(calls, []string{"h"}) {
		t.Fatalf("mw 注册晚于 /api，不应触发；got %v", calls)
	}
}

// 回归: /#... 兜底路由必须能匹配 /（无段路径）。
func TestRootCatchAllMatchesEmptyPath(t *testing.T) {
	g := New()
	g.GET("/#...", func(ctx *Context) {
		ctx.Writer.Write([]byte("fb:" + ctx.GetUrlPathParam("#")))
	})
	if got := runRequest(t, g, "GET", "/"); got != "fb:" {
		t.Fatalf("/ 应被 /#... 兜底；got %q", got)
	}
}

// 回归: 节点存在但未登记 method 时，应回落到该节点下挂的 catchAll。
func TestCatchAllFallbackOnIntermediateNode(t *testing.T) {
	g := New()
	g.GET("/a/#...", func(ctx *Context) {
		ctx.Writer.Write([]byte("ca:" + ctx.GetUrlPathParam("#")))
	})
	g.GET("/a/b", func(ctx *Context) { ctx.Writer.Write([]byte("ab")) })

	if got := runRequest(t, g, "GET", "/a/b"); got != "ab" {
		t.Fatalf("got %q", got)
	}
	if got := runRequest(t, g, "GET", "/a/c"); got != "ca:c" {
		t.Fatalf("got %q", got)
	}
	// /a 自身没注册，应该回落到 /a/#... (#="")。
	if got := runRequest(t, g, "GET", "/a"); got != "ca:" {
		t.Fatalf("got %q", got)
	}
}

func TestParamMultipleSegments(t *testing.T) {
	g := New()
	g.GET("/u/#id/p/#pid", func(ctx *Context) {
		ctx.Writer.Write([]byte(ctx.GetUrlPathParam("id") + "/" + ctx.GetUrlPathParam("pid")))
	})
	if got := runRequest(t, g, "GET", "/u/42/p/9"); got != "42/9" {
		t.Fatalf("got %q", got)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ---- 生产级行为：404 / 405 / panic 不泄漏 Context ----

// recoveredHandler 让 net/http 在测试里把 panic 当 500 处理，避免污染测试输出。
func runWithRecover(t *testing.T, g *GOweb, method, path string) (rec *httptest.ResponseRecorder) {
	t.Helper()
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	defer func() { _ = recover() }()
	g.ServeHTTP(rec, req)
	return rec
}

func TestContextNotLeakedOn404(t *testing.T) {
	g := New()
	g.GET("/known", func(ctx *Context) { ctx.Writer.Write([]byte("ok")) })
	for i := 0; i < 200; i++ {
		runRequest(t, g, "GET", "/missing")
	}
	for i := 0; i < 200; i++ {
		runRequest(t, g, "GET", "/known")
	}
	// 简单 sanity：goroutine 数稳定（不严格断言数值，只防漏 goroutine）。
	runtime.GC()
	if n := runtime.NumGoroutine(); n > 50 {
		t.Fatalf("unexpected goroutine count: %d", n)
	}
}

func TestContextNotLeakedOn405(t *testing.T) {
	g := New()
	g.GET("/x", func(ctx *Context) { ctx.Writer.Write([]byte("ok")) })
	for i := 0; i < 100; i++ {
		rec := runRequest(t, g, "POST", "/x")
		_ = rec
	}
}

func TestContextNotLeakedOnPanic(t *testing.T) {
	g := New()
	g.GET("/boom", func(ctx *Context) { panic("kaboom") })
	for i := 0; i < 50; i++ {
		runWithRecover(t, g, "GET", "/boom")
	}
	runtime.GC()
}

// ---- 客户端断开能传到 handler 的 ctx ----
func TestClientCancelPropagates(t *testing.T) {
	g := New()
	gotDone := make(chan struct{}, 1)
	g.GET("/wait", func(ctx *Context) {
		select {
		case <-ctx.Done():
			gotDone <- struct{}{}
		case <-time.After(time.Second):
			t.Errorf("expected ctx done after client cancel")
		}
	})

	rec := httptest.NewRecorder()
	clientCtx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest("GET", "/wait", nil).WithContext(clientCtx)
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	g.ServeHTTP(rec, req)
	select {
	case <-gotDone:
	case <-time.After(time.Second):
		t.Fatalf("ctx.Done() did not fire")
	}
}

// ---- 405 + Allow 头 ----
func Test405WithAllowHeader(t *testing.T) {
	g := New()
	g.GET("/x", func(ctx *Context) {})
	g.PUT("/x", func(ctx *Context) {})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/x", nil)
	g.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	allow := rec.Header().Get("Allow")
	if allow != "GET, PUT" {
		t.Fatalf("Allow header want %q got %q", "GET, PUT", allow)
	}
}

// strict_405 语义：路径匹配后只看本节点的方法集，不再回落到祖先的 catchAll。
func Test405StrictDoesNotFallthroughCatchAll(t *testing.T) {
	g := New()
	g.GET("/api", func(ctx *Context) { ctx.Writer.Write([]byte("ok")) })
	g.GET("/#...", func(ctx *Context) { ctx.Writer.Write([]byte("fb")) })
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api", nil)
	g.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("/api POST 应严格 405；got %d body=%q", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Allow"); got != "GET" {
		t.Fatalf("Allow want GET got %q", got)
	}
}

// ---- URL 解码 ----
func TestURLDecodedPath(t *testing.T) {
	g := New()
	g.GET("/hello world", func(ctx *Context) { ctx.Writer.Write([]byte("h1")) })
	g.GET("/中文", func(ctx *Context) { ctx.Writer.Write([]byte("h2")) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/hello%20world", nil)
	g.ServeHTTP(rec, req)
	if rec.Body.String() != "h1" {
		t.Fatalf("expected h1, got %q", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/%E4%B8%AD%E6%96%87", nil)
	g.ServeHTTP(rec, req)
	if rec.Body.String() != "h2" {
		t.Fatalf("expected h2, got %q", rec.Body.String())
	}
}

// ---- 并发安全 (通过 -race 时该用例真正发力) ----
func TestRegisterDuringServe_Race(t *testing.T) {
	g := New()
	g.GET("/seed", func(ctx *Context) { ctx.Writer.Write([]byte("seed")) })

	var stop atomic.Bool
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; !stop.Load() && i < 500; i++ {
			path := "/r" + strconv.Itoa(i)
			func() {
				defer func() { _ = recover() }()
				g.GET(path, func(ctx *Context) { ctx.Writer.Write([]byte("ok")) })
			}()
		}
	}()

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; !stop.Load() && j < 500; j++ {
				runRequest(t, g, "GET", "/seed")
			}
		}()
	}

	time.Sleep(50 * time.Millisecond)
	stop.Store(true)
	wg.Wait()
}

// ---- position 原子性：错误注册不应留孤儿节点 ----
func TestPositionAtomicityOnError(t *testing.T) {
	g := New()
	g.GET("/healthy", func(ctx *Context) {})

	func() {
		defer func() { _ = recover() }()
		// /b/c 还不存在；非法的 #... 不在末尾。注册前 dry-run 会报错。
		g.GET("/b/c/#.../x", func(ctx *Context) {})
	}()

	_, paths := g.List()
	for _, p := range paths {
		if p == "/b" || p == "/b/c" || strings.HasPrefix(p, "/b/c/") {
			t.Fatalf("不应有孤儿条目: %s", p)
		}
	}
}

// ---- ANY 对称 ----
func TestAnyAfterGetPanics(t *testing.T) {
	g := New()
	g.GET("/x", func(ctx *Context) {})
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Any 在已有具体 method 的路由上应 panic")
		}
	}()
	g.Any("/x", func(ctx *Context) {})
}

// ---- List 输出确定 ----
func TestListSortedDeterministic(t *testing.T) {
	g := New()
	g.GET("/c", func(ctx *Context) {})
	g.POST("/b", func(ctx *Context) {})
	g.GET("/a", func(ctx *Context) {})
	g.GET("/b", func(ctx *Context) {})

	for i := 0; i < 50; i++ {
		ms, ps := g.List()
		want := []string{"GET /a", "GET /b", "POST /b", "GET /c"}
		got := make([]string, len(ms))
		for i := range ms {
			got[i] = ms[i] + " " + ps[i]
		}
		if !equalStrings(got, want) {
			t.Fatalf("iter %d: got %v want %v", i, got, want)
		}
	}
}

// ---- nil handler 拒绝 ----
func TestNilHandlerPanics(t *testing.T) {
	g := New()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("nil handler 应 panic")
		}
	}()
	g.GET("/x", nil)
}

func TestNilMiddlewarePanics(t *testing.T) {
	g := New()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("nil middleware 应 panic")
		}
	}()
	g.HeadMiddleware(nil)
}

// ---- 注册热更新：旧请求看旧表，新请求看新表 ----
func TestHotRegisterAfterServe(t *testing.T) {
	g := New()
	g.GET("/a", func(ctx *Context) { ctx.Writer.Write([]byte("a")) })

	if got := runRequest(t, g, "GET", "/b"); !strings.Contains(got, "404") {
		t.Fatalf("/b 应未注册；got %q", got)
	}
	g.GET("/b", func(ctx *Context) { ctx.Writer.Write([]byte("b")) })
	if got := runRequest(t, g, "GET", "/b"); got != "b" {
		t.Fatalf("/b 注册后应可达；got %q", got)
	}
}

// ---- Middlewares + ctx.Next()：手写洋葱模型 ----
func TestMiddlewaresExplicitNext(t *testing.T) {
	g := New()
	var calls []string
	g.Middlewares(func(ctx *Context) {
		calls = append(calls, "pre1")
		ctx.Next()
		calls = append(calls, "post1")
	}, func(ctx *Context) {
		calls = append(calls, "pre2")
		ctx.Next()
		calls = append(calls, "post2")
	})
	g.GET("/x", func(ctx *Context) { calls = append(calls, "h") })

	runRequest(t, g, "GET", "/x")
	want := []string{"pre1", "pre2", "h", "post2", "post1"}
	if !equalStrings(calls, want) {
		t.Fatalf("洋葱模型顺序错误\n got: %v\nwant: %v", calls, want)
	}
}

// ---- 未显式 Next() 的中间件不阻断后续链 ----
func TestMiddlewaresWithoutNextStillFlows(t *testing.T) {
	g := New()
	var calls []string
	g.Middlewares(func(ctx *Context) {
		calls = append(calls, "mw")
		// 故意不调用 ctx.Next()，框架应自动继续
	})
	g.GET("/x", func(ctx *Context) { calls = append(calls, "h") })

	runRequest(t, g, "GET", "/x")
	if !equalStrings(calls, []string{"mw", "h"}) {
		t.Fatalf("隐式继续应使 handler 仍执行；got %v", calls)
	}
}

// ---- 多层级中间件按 root → leaf 顺序执行 ----
func TestMiddlewareRootToLeafOrder(t *testing.T) {
	g := New()
	var calls []string
	g.HeadMiddleware(func(ctx *Context) { calls = append(calls, "root") })
	api := g.Grep("/api")
	api.HeadMiddleware(func(ctx *Context) { calls = append(calls, "api") })
	api.GET("/x", func(ctx *Context) { calls = append(calls, "h") })

	runRequest(t, g, "GET", "/api/x")
	want := []string{"root", "api", "h"}
	if !equalStrings(calls, want) {
		t.Fatalf("root→leaf 顺序错误\n got: %v\nwant: %v", calls, want)
	}
}

// ---- 多个 FootMiddleware：洋葱模型下后置段为 LIFO ----
func TestMultipleFootLIFO(t *testing.T) {
	g := New()
	var calls []string
	g.FootMiddleware(func(ctx *Context) { calls = append(calls, "f1") })
	g.FootMiddleware(func(ctx *Context) { calls = append(calls, "f2") })
	g.GET("/x", func(ctx *Context) { calls = append(calls, "h") })

	runRequest(t, g, "GET", "/x")
	// f1 是外层（先注册）→ Next() 完成后才跑 → 在 f2 之后
	want := []string{"h", "f2", "f1"}
	if !equalStrings(calls, want) {
		t.Fatalf("Foot 应为 LIFO\n got: %v\nwant: %v", calls, want)
	}
}

// ---- 在中间件中 Stop()：链立即短路，handler 不执行 ----
func TestStopAbortsChain(t *testing.T) {
	g := New()
	var calls []string
	g.Middlewares(func(ctx *Context) {
		calls = append(calls, "mw")
		ctx.Stop()
		ctx.Next() // 显式调用，验证短路：后续仍不应执行
	})
	g.GET("/x", func(ctx *Context) { calls = append(calls, "h") })

	runRequest(t, g, "GET", "/x")
	if !equalStrings(calls, []string{"mw"}) {
		t.Fatalf("Stop 后链应短路；got %v", calls)
	}
}

// ---- Head + Foot 混合：注册顺序 head, foot, head, foot ----
func TestHeadFootMixed(t *testing.T) {
	g := New()
	var calls []string
	g.HeadMiddleware(func(ctx *Context) { calls = append(calls, "H1") })
	g.FootMiddleware(func(ctx *Context) { calls = append(calls, "F1") })
	g.HeadMiddleware(func(ctx *Context) { calls = append(calls, "H2") })
	g.FootMiddleware(func(ctx *Context) { calls = append(calls, "F2") })
	g.GET("/x", func(ctx *Context) { calls = append(calls, "h") })

	runRequest(t, g, "GET", "/x")
	// 链：[H1, F1(Next;f1), H2, F2(Next;f2), h]
	// 展开：H1 -> F1.Next -> H2 -> F2.Next -> h -> F2.post -> F1.post
	want := []string{"H1", "H2", "h", "F2", "F1"}
	if !equalStrings(calls, want) {
		t.Fatalf("混合顺序错误\n got: %v\nwant: %v", calls, want)
	}
}

// ---- Middlewares 拒绝 nil ----
func TestMiddlewaresNilPanics(t *testing.T) {
	g := New()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("nil middleware 应 panic")
		}
	}()
	g.Middlewares(nil)
}

// ---- OnPanic 钩子被调用 + 默认行为写 500 ----
func TestOnPanicHookInvoked(t *testing.T) {
	g := New()
	var captured any
	g.OnPanic(func(ctx *Context, rec any) {
		captured = rec
		http.Error(ctx.Writer, "boom-handled", http.StatusBadGateway)
	})
	g.GET("/x", func(ctx *Context) { panic("boom") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	g.ServeHTTP(rec, req)

	if captured != "boom" {
		t.Fatalf("OnPanic 未拿到 panic 值: %v", captured)
	}
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("OnPanic 自定义状态码失效: %d", rec.Code)
	}
}

// ---- 默认 OnPanic：未自定义时 header 没发就写 500 ----
func TestDefaultOnPanicWrites500(t *testing.T) {
	g := New()
	g.GET("/x", func(ctx *Context) { panic("boom") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	g.ServeHTTP(rec, req) // 不应再向上抛

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("默认 panic 未写 500: %d body=%q", rec.Code, rec.Body.String())
	}
}

// ---- http.ErrAbortHandler 应当继续上抛（stdlib 协议）----
func TestErrAbortHandlerStillPropagates(t *testing.T) {
	g := New()
	g.GET("/x", func(ctx *Context) { panic(http.ErrAbortHandler) })
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	defer func() {
		r := recover()
		if r != http.ErrAbortHandler {
			t.Fatalf("ErrAbortHandler 应被透传，got %v", r)
		}
	}()
	g.ServeHTTP(rec, req)
}

// ---- HEAD 自动回落到 GET ----
func TestHeadFallbackToGet(t *testing.T) {
	g := New()
	g.GET("/x", func(ctx *Context) { ctx.Writer.Write([]byte("body")) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/x", nil)
	g.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("HEAD 应回落 GET 200，got %d", rec.Code)
	}
}

// ---- HEAD 自动回落不影响显式注册的方法集 ----
func TestHeadDoesNotOverrideAllow(t *testing.T) {
	g := New()
	g.POST("/x", func(ctx *Context) {})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/x", nil)
	g.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("没有 GET 时 HEAD 应 405，got %d", rec.Code)
	}
}

// ---- ResponseWriter 状态跟踪 ----
func TestResponseWriterTracksStatusAndSize(t *testing.T) {
	g := New()
	g.GET("/x", func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusTeapot)
		ctx.Writer.Write([]byte("hello"))
		// 重复 WriteHeader 应被吞掉
		ctx.Writer.WriteHeader(http.StatusBadGateway)

		rw, ok := ctx.Writer.(ResponseWriter)
		if !ok {
			t.Fatalf("Writer 应实现 web.ResponseWriter")
		}
		if rw.Status() != http.StatusTeapot {
			t.Fatalf("status 跟踪错: %d", rw.Status())
		}
		if rw.Size() != 5 {
			t.Fatalf("size 跟踪错: %d", rw.Size())
		}
		if !rw.Written() {
			t.Fatalf("Written 应为 true")
		}
	})
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, httptest.NewRequest("GET", "/x", nil))
	if rec.Code != http.StatusTeapot {
		t.Fatalf("第二次 WriteHeader 不应生效: %d", rec.Code)
	}
}

// ---- SetValue 通过 Context.Value 回退查询（不再依赖 mergedContext）----
func TestSetValueFallback(t *testing.T) {
	g := New()
	g.SetValue("k", "v")
	g.GET("/x", func(ctx *Context) {
		got := ctx.GetContextValue("k")
		if got != "v" {
			t.Fatalf("ctx.Value fallback 失效: got %v", got)
		}
		// SetContextValue 优先级高于全局 fallback
		ctx.SetContextValue("k", "override")
		if v := ctx.GetContextValue("k"); v != "override" {
			t.Fatalf("局部 set 应覆盖全局: got %v", v)
		}
	})
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, httptest.NewRequest("GET", "/x", nil))
}

// ---- Server 默认超时已注入 ----
func TestServerDefaultTimeouts(t *testing.T) {
	g := New()
	if g.Server == nil {
		t.Fatalf("Server 不应为 nil")
	}
	if g.Server.ReadHeaderTimeout == 0 || g.Server.ReadTimeout == 0 ||
		g.Server.WriteTimeout == 0 || g.Server.IdleTimeout == 0 {
		t.Fatalf("默认超时未注入: %+v", g.Server)
	}
	if g.Server.Handler == nil {
		t.Fatalf("Handler 应自动绑定为 g 自己")
	}
}

// ---- WithServer 整体替换 ----
func TestWithServer(t *testing.T) {
	custom := &http.Server{ReadHeaderTimeout: time.Second}
	g := New(WithServer(custom))
	if g.Server != custom {
		t.Fatalf("WithServer 未替换 Server")
	}
	if g.Server.Handler != g {
		t.Fatalf("Handler 应自动绑定")
	}
}

// ---- RunContext 在 ctx 取消时优雅关停 ----
func TestRunContextShutdown(t *testing.T) {
	g := New()
	g.GET("/ping", func(c *Context) { c.Writer.Write([]byte("pong")) })

	// 绑定到 ephemeral port 以避免端口冲突。
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- g.RunContext(ctx, addr) }()

	// 等监听就绪
	deadline := time.Now().Add(2 * time.Second)
	for {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			_ = conn.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("server 未就绪: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("RunContext 退出错: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Shutdown 超时")
	}
}

// ---- ResponseWriter Flusher 透传 ----
func TestResponseWriterFlusherPassthrough(t *testing.T) {
	g := New()
	g.GET("/x", func(ctx *Context) {
		f, ok := ctx.Writer.(http.Flusher)
		if !ok {
			t.Fatalf("Flusher 接口应可断言")
		}
		ctx.Writer.Write([]byte("part1"))
		f.Flush()
	})
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, httptest.NewRequest("GET", "/x", nil))
	if rec.Body.String() != "part1" {
		t.Fatalf("got %q", rec.Body.String())
	}
}
