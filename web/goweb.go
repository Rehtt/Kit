/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/7/16 下午 04:30
 */

package goweb

import (
	"net/http"
)

type GOweb struct {
	RouterGroup
	routerGroupLock bool

	noRouter HandlerFunc
}

func (g *GOweb) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if !g.routerGroupLock {
		g.routerGroupLock = true
	}

	ctx := &Context{
		Request: request,
		Writer:  writer,
		survive: true,
	}
	match, handleFunc, grep := g.PathMatch(request.RequestURI, request.Method)
	if handleFunc == nil {
		g.handler404(ctx)
		return
	}

	ctx.param = match

	for grep != nil {
		for i := range grep.middlewares {
			ctx.runFunc(grep.middlewares[i])
		}
		grep = grep.parent
	}

	ctx.runFunc(handleFunc)
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

func New() *GOweb {
	return new(GOweb)
}
