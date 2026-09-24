package web

import (
	"context"
	"net/http"
	"time"
)

type Option func(g *GOweb)

// WithRoutes 在 New 的构造阶段批量注册路由。构造完成后只发布一次快照，
// 适合大量初始路由；运行期直接调用注册 API 仍会立即发布。
func WithRoutes(register func(*RouterGroup)) Option {
	return func(g *GOweb) {
		if register != nil {
			register(&g.RouterGroup)
		}
	}
}

// WithContext 设置全局 value chain，仅用作 ctx.Value 的 fallback 源，
func WithContext(ctx context.Context) Option {
	return func(g *GOweb) {
		g.valuesMu.Lock()
		defer g.valuesMu.Unlock()
		g.Context = ctx
	}
}

// WithServer 整体替换底层 *http.Server，Handler 会被强制指向 g。
func WithServer(server *http.Server) Option {
	return func(g *GOweb) {
		if server == nil {
			return
		}
		server.Handler = g
		g.Server = server
	}
}

func WithOnPanic(fn func(*Context, any)) Option {
	return func(g *GOweb) {
		g.onPanic = fn
	}
}

// WithShutdownTimeout 设置 RunContext 的独立关停超时，默认 30 秒。
// timeout 必须大于零，否则 panic；重复配置以最后一次为准。
// 不影响直接调用 Shutdown(ctx) 的超时。
func WithShutdownTimeout(timeout time.Duration) Option {
	if timeout <= 0 {
		panic("[web] WithShutdownTimeout: timeout must be positive")
	}
	return func(g *GOweb) { g.shutdownTimeout = timeout }
}
