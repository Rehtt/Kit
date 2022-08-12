/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/7/17 下午 12:10
 */

package goweb

func (g *RouterGroup) GET(path string, handlerFunc HandlerFunc) {
	g.handle(GET, path, handlerFunc)
}
func (g *RouterGroup) POST(path string, handlerFunc HandlerFunc) {
	g.handle(POST, path, handlerFunc)
}
func (g *RouterGroup) PUT(path string, handlerFunc HandlerFunc) {
	g.handle(PUT, path, handlerFunc)
}
func (g *RouterGroup) DELETE(path string, handlerFunc HandlerFunc) {
	g.handle(DELETE, path, handlerFunc)
}
func (g *RouterGroup) HEAD(path string, handlerFunc HandlerFunc) {
	g.handle(HEAD, path, handlerFunc)
}
func (g *RouterGroup) OPTIONS(path string, handlerFunc HandlerFunc) {
	g.handle(OPTIONS, path, handlerFunc)
}
func (g *RouterGroup) CONNECT(path string, handlerFunc HandlerFunc) {
	g.handle(CONNECT, path, handlerFunc)
}
func (g *RouterGroup) Any(path string, handlerFunc HandlerFunc) {
	g.handle(ANY, path, handlerFunc)
}
