/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/7/16 下午 05:02
 */

package goweb

import (
	"strings"
	"sync/atomic"
)

type RouterGroup struct {
	index       uint32
	middlewares []middleware
	path        string
	method      map[string]HandlerFunc
	child       map[string]*RouterGroup
	parent      *RouterGroup

	// 记录路由添加顺序
	order       uint32  // 添加顺序
	globalCount *uint32 // 全局计数
}
type middleware struct {
	HandlerFunc
	order uint32
}

func (g *RouterGroup) Grep(path string) *RouterGroup {
	return g.position(path)
}
func (g *RouterGroup) Middleware(handlers ...HandlerFunc) {
	if len(g.middlewares) == 0 {
		g.middlewares = make([]middleware, 0, len(handlers)+5)
	}

	for i := range handlers {
		g.middlewares = append(g.middlewares, middleware{
			HandlerFunc: handlers[i],
			order:       atomic.AddUint32(g.globalCount, 1), // 记录中间件添加时的位置
		})
	}
}

func (g *RouterGroup) position(path string) *RouterGroup {

	for _, p := range strings.Split(path, "/") {
		if p == "" {
			continue
		}

		if _, ok := g.child[p]; ok {
			g = g.child[p]
			continue
		}
		if g.child == nil {
			g.child = make(map[string]*RouterGroup, 10)
		}
		if p[0] == '#' {
			for child := range g.child {
				if child[0] == '#' {
					panic(path + " 地址泛匹配重复")
				}
			}
		}

		g.child[p] = &RouterGroup{
			path:        p,
			parent:      g,
			index:       g.index + 1,
			globalCount: g.globalCount,
		}
		g = g.child[p]
	}

	if g.globalCount == nil {
		g.globalCount = new(uint32)
	}
	// 记录添加路由顺序
	g.order = atomic.AddUint32(g.globalCount, 1)

	return g
}

func (g *RouterGroup) completePath() string {
	completePath := make([]string, g.index)
	for g.index != 0 {
		completePath[g.index-1] = g.path
		g = g.parent
	}
	return "/" + strings.Join(completePath, "/")
}
func (g *RouterGroup) handle(method string, path string, handlerFunc HandlerFunc) {
	g = g.position(path)
	if method == ANY && len(g.method) > 1 {
		if _, ok := g.method[ANY]; ok {
			panic(g.completePath() + "该路由any方法冲突")
		}
	}
	if _, ok := g.method[method]; ok {
		panic(g.completePath() + "该路由method重复")
	}
	if g.method == nil {
		g.method = make(map[string]HandlerFunc, 5)
	}
	g.method[method] = handlerFunc
}

func (g *RouterGroup) PathMatch(path, method string) (match map[string]string, handle HandlerFunc, grep *RouterGroup) {
	var ok bool
	for _, p := range strings.Split(path, "/") {
		if p == "" {
			continue
		}
		if _, ok = g.child[p]; ok {
			g = g.child[p]
			continue
		}
		var find bool
		for child := range g.child {
			if child[0] == '#' {
				if match == nil {
					match = make(map[string]string)
				}
				match[child[1:]] = p
				g = g.child[child]
				find = true
				break
			}
		}
		if !find {
			return
		}
	}

	handle, ok = g.method[method]
	if !ok {
		for m := range g.method {
			if m == ANY {
				handle = g.method[m]
			}
		}
	}
	grep = g
	return
}
