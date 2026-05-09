package web

import (
	"context"
	"net/http"
)

type Option func(g *GOweb)

// WithContext 设置全局 value chain，仅用作 ctx.Value 的 fallback 源，
func WithContext(ctx context.Context) Option {
	return func(g *GOweb) {
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
