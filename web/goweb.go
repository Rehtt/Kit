/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/7/16 下午 04:30
 */

package goweb

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

var (
	// 内存优化
	contextPool = sync.Pool{
		New: func() interface{} {
			return new(Context)
		},
	}
)

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
	for grep != nil {
		for i := range grep.middlewares {
			// 中间件只覆盖顺序靠后的路由
			if grep.middlewares[i].order < handleFuncOrder {
				ctx.runFunc(grep.middlewares[i].HandlerFunc)
			}
		}

		grep = grep.parent
	}

	ctx.runFunc(handleFunc)
	contextPool.Put(ctx)
}

func (g *GOweb) NoRoute(handlerFunc HandlerFunc) {
	g.noRouter = handlerFunc
}
func (g *GOweb) handler404(ctx *Context) {
	ctx.Writer.WriteHeader(http.StatusNotFound)
	if g.noRouter != nil {
		g.noRouter(ctx)
	} else {
		http.NotFound(ctx.Writer, ctx.Request)
	}
}
func New() (g *GOweb) {
	g = new(GOweb)
	g.Context = context.Background()
	return
}
func (g *GOweb) SetValue(key, value interface{}) {
	g.Context = context.WithValue(g.Context, key, value)
}
func (g *GOweb) GetValue(key interface{}) interface{} {
	return g.Value(key)
}
