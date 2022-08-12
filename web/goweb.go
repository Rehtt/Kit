/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/7/16 下午 04:30
 */

package goweb

import (
	"context"
	"github.com/Rehtt/Kit/util"
	"net/http"
)

type GOweb struct {
	RouterGroup
	noRouter HandlerFunc
	values   map[interface{}]interface{}
}

func (g *GOweb) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	ctx := &Context{
		Request: request,
		Writer:  writer,
		survive: true,
		Context: context.Background(),
		values:  util.DeepCopy(g.values).(map[interface{}]interface{}),
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
func (g *GOweb) SetKeyValue(key, value interface{}) {
	if g.values == nil {
		g.values = make(map[interface{}]interface{})
	}
	g.values[key] = value
}
func New() (g *GOweb) {
	return new(GOweb)
}
