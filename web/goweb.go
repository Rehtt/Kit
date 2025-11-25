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

// 内存优化
var contextPool = sync.Pool{
	New: func() any {
		return new(Context)
	},
}

func (g *GOweb) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	c, cancel := context.WithCancel(g.Context)
	ctx := contextPool.Get().(*Context)
	ctx.Request = request
	ctx.Writer = writer
	ctx.Context = c
	ctx.cancel = cancel

	match, handleFunc, grep := g.PathMatch(request.RequestURI, request.Method)
	if handleFunc == nil {
		g.handler404(ctx)
		return
	}

	ctx.param = match

	handleFuncOrder := grep.order
	gp := grep
	for gp != nil {
		for i := range gp.middlewares {
			// 中间件只覆盖顺序靠后的路由
			if gp.middlewares[i].order < handleFuncOrder {
				ctx.runFunc(gp.middlewares[i].HandlerFunc)
			}
		}

		gp = gp.parent
	}

	ctx.runFunc(handleFunc)

	//
	gp = grep
	for gp != nil {
		for i := range gp.footMiddle {
			// 中间件只覆盖顺序靠后的路由
			if gp.footMiddle[i].order < handleFuncOrder {
				ctx.runFunc(gp.footMiddle[i].HandlerFunc)
			}
		}
		gp = gp.parent
	}

	ctx.Stop()
	contextPool.Put(ctx)
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

func New() (g *GOweb) {
	g = new(GOweb)
	g.RouterGroup.globalCount = new(uint32)
	g.Context = context.Background()
	return
}

func (g *GOweb) SetValue(key, value any) {
	g.Context = context.WithValue(g.Context, key, value)
}

func (g *GOweb) GetValue(key any) any {
	return g.Value(key)
}
