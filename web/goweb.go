/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/7/16 下午 04:30
 */

package web

import (
	"context"
	"net/http"
	"sync"
)

type GOweb struct {
	RouterGroup
	noRouter HandlerFunc
	context.Context
}

// 内存优化：复用 *Context，由 ServeHTTP 在 defer 中重置并归还。
var contextPool = sync.Pool{
	New: func() any {
		return new(Context)
	},
}

// mergedContext 把 request.Context()（用于客户端断开取消）与 GOweb.Context
// （用户通过 SetValue 配置的全局 value chain）合并：取消语义来自前者，
// Value 查找未命中时回退到后者。
type mergedContext struct {
	context.Context
	values context.Context
}

func (m *mergedContext) Value(key any) any {
	if v := m.Context.Value(key); v != nil {
		return v
	}
	if m.values == nil {
		return nil
	}
	return m.values.Value(key)
}

func (g *GOweb) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var snap *routeSnapshot
	if g.host != nil {
		snap = g.host.snapshot.Load()
	}

	parentCtx := request.Context()
	if g.Context != nil {
		parentCtx = &mergedContext{Context: parentCtx, values: g.Context}
	}
	rctx, cancel := context.WithCancel(parentCtx)

	ctx := contextPool.Get().(*Context)
	ctx.Request = request
	ctx.Writer = writer
	ctx.Context = rctx
	ctx.cancel = cancel
	ctx.param = nil

	defer func() {
		rec := recover()
		if ctx.cancel != nil {
			ctx.cancel()
		}
		ctx.Request = nil
		ctx.Writer = nil
		ctx.Context = nil
		ctx.cancel = nil
		ctx.param = nil
		contextPool.Put(ctx)
		if rec != nil {
			panic(rec)
		}
	}()

	if snap == nil {
		http.NotFound(writer, request)
		return
	}

	// 用 URL.Path（已解码），而不是 RequestURI（含百分号编码与 query）。
	params, handleFunc, leaf, allowed := snap.match(request.URL.Path, request.Method)
	if handleFunc == nil {
		if !allowed.empty() {
			writer.Header().Set("Allow", allowed.headerValue())
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		g.handler404(ctx)
		return
	}

	ctx.param = params

	handleFuncOrder := leaf.order
	for gp := leaf; gp != nil; gp = gp.parent {
		for i := range gp.middlewares {
			if gp.middlewares[i].order < handleFuncOrder {
				ctx.runFunc(gp.middlewares[i].HandlerFunc)
			}
		}
	}

	ctx.runFunc(handleFunc)

	for gp := leaf; gp != nil; gp = gp.parent {
		for i := range gp.footMiddle {
			if gp.footMiddle[i].order < handleFuncOrder {
				ctx.runFunc(gp.footMiddle[i].HandlerFunc)
			}
		}
	}
}

func (g *GOweb) NoRoute(handlerFunc HandlerFunc) {
	g.noRouter = handlerFunc
}

func (g *GOweb) handler404(ctx *Context) {
	if g.noRouter != nil {
		g.noRouter(ctx)
	} else {
		http.NotFound(ctx.Writer, ctx.Request)
	}
}

func New(opts ...Option) (g *GOweb) {
	g = new(GOweb)
	reg := &registry{}
	g.RouterGroup.host = reg
	g.Context = context.Background()
	for _, opt := range opts {
		opt(g)
	}
	// 初始化空快照，避免首个请求看到 nil。
	reg.publish(&g.RouterGroup)
	return
}

func (g *GOweb) SetValue(key, value any) {
	g.Context = context.WithValue(g.Context, key, value)
}

func (g *GOweb) GetValue(key any) any {
	return g.Value(key)
}

func (g *GOweb) Run(addr string) error {
	return http.ListenAndServe(addr, g)
}

func (g *GOweb) RunTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, g)
}
