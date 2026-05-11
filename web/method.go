/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/7/17 下午 12:10
 */

package web

func (g *RouterGroup) GET(path string, handlerFunc HandlerFunc, opt ...HandlerOpt) {
	g.handle(GET, path, handlerFunc, opt...)
}

func (g *RouterGroup) POST(path string, handlerFunc HandlerFunc, opt ...HandlerOpt) {
	g.handle(POST, path, handlerFunc, opt...)
}

func (g *RouterGroup) PUT(path string, handlerFunc HandlerFunc, opt ...HandlerOpt) {
	g.handle(PUT, path, handlerFunc, opt...)
}

func (g *RouterGroup) DELETE(path string, handlerFunc HandlerFunc, opt ...HandlerOpt) {
	g.handle(DELETE, path, handlerFunc, opt...)
}

func (g *RouterGroup) HEAD(path string, handlerFunc HandlerFunc, opt ...HandlerOpt) {
	g.handle(HEAD, path, handlerFunc, opt...)
}

func (g *RouterGroup) OPTIONS(path string, handlerFunc HandlerFunc, opt ...HandlerOpt) {
	g.handle(OPTIONS, path, handlerFunc, opt...)
}

func (g *RouterGroup) CONNECT(path string, handlerFunc HandlerFunc, opt ...HandlerOpt) {
	g.handle(CONNECT, path, handlerFunc, opt...)
}

func (g *RouterGroup) Any(path string, handlerFunc HandlerFunc, opt ...HandlerOpt) {
	g.handle(ANY, path, handlerFunc, opt...)
}
