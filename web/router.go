/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/7/16 下午 05:02
 */

package web

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type nodeKind uint8

const (
	kindStatic nodeKind = iota
	kindParam
	kindCatchAll
)

const (
	paramPrefix     = "#"
	catchAllSegment = "#..."
)

// registry 是 GOweb 的注册中心：在写端持有 mu 与全局序号；在读端持有
// atomic.Pointer 指向只读的 routeSnapshot。所有节点共享同一个 *registry。
//
// 写流程：mu.Lock -> 修改 writeRoot 树 -> publish() 深拷贝出新 snapshot
//
//	-> atomic.Store(snapshot) -> mu.Unlock。
//
// 读流程：snapshot.Load() 拿到不可变快照 -> 直接 match。
type registry struct {
	mu           sync.Mutex
	snapshot     atomic.Pointer[routeSnapshot]
	globalCount  uint32 // 仅在 mu 持有期间被修改
	root         *RouterGroup
	initializing bool
}

// routeSnapshot 是某一时刻路由表的不可变快照；ServeHTTP 完全依赖它做匹配。
type routeSnapshot struct {
	root       *RouterGroup
	flatRoutes map[string]*RouterGroup // normalize 后的纯静态路径 -> 叶子节点
	hasDynamic bool                    // 是否含 #name / #...
}

// RouterGroup 是段级压缩前缀树（radix tree）的节点。
//
// 一条边由若干「路径段」组成，存放于 segments：
//   - kindStatic   : 普通段链，如 ["api","v1","users"]
//   - kindParam    : 单段，形如 ["#name"]
//   - kindCatchAll : 单段 ["#..."]
//
// 同一节点同时最多拥有：
//   - 任意多个静态子节点（按首段分桶到 staticKids）
//   - 至多一个 paramKid（不同名字会冲突）
//   - 至多一个 catchAllKid
//
// 匹配优先级：staticKids > paramKid > catchAllKid。
//
// 写端节点：host != nil；可在该节点上调用注册 API。
// 快照节点：host == nil；只读，注册 API 会 panic（防止误改快照）。
type RouterGroup struct {
	kind nodeKind
	// 路由 / 中间件注册时的全局序号；中间件仅作用于序号比自己更大的注册路由。
	order uint32

	segments []string

	parent      *RouterGroup
	staticKids  map[string]*RouterGroup
	paramKid    *RouterGroup
	catchAllKid *RouterGroup

	method      map[string]HandlerFunc
	methodOrder map[string]uint32
	options     map[string]HandlerOpt
	middlewares []middleware

	// handlerChains is populated only on immutable snapshot nodes. Each method
	// owns a distinct backing array so a snapshot or route cannot mutate another
	// route's execution chain.
	handlerChains map[string][]HandlerFunc

	// Grep 返回的分组使用稳定的逻辑路径。压缩边后续分裂时，
	// 逻辑分组不会因内部节点改写而改变含义。
	routePath []string
	isProxy   bool

	// 写端节点共享的注册中心；快照节点为 nil。
	host *registry
}

type middleware struct {
	HandlerFunc
	order uint32
}

// methodSet 用于 405 时构造 Allow 响应头。
type methodSet map[string]HandlerFunc

func (m methodSet) empty() bool { return len(m) == 0 }

func (m methodSet) headerValue() string {
	if len(m) == 0 {
		return ""
	}
	out := make([]string, 0, len(m))
	for k := range m {
		if k == ANY {
			continue
		}
		out = append(out, k)
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}

// requireHost 取注册中心；写端 API 必须有 host，否则 panic（防止用户对快照 / 裸节点写入）。
func (g *RouterGroup) requireHost() *registry {
	if g.host == nil {
		panic("[web] router not initialized; use web.New() to construct")
	}
	return g.host
}

// Grep 取或创建 path 对应的子节点，便于在该位置挂载中间件 / 子路由。
// 若 path 落在某条压缩边的中段，会自动从该边分裂出对应节点。
func (g *RouterGroup) Grep(path string) *RouterGroup {
	host := g.requireHost()
	host.mu.Lock()
	defer host.mu.Unlock()

	segs := splitSegments(path)
	root := host.root
	full := append(g.baseRoutePath(), segs...)
	if err := dryRunValidate(root, full); err != nil {
		panic(formatPanic(g, path, err))
	}
	_, err := root.position(full)
	if err != nil {
		panic(formatPanic(g, path, err))
	}
	host.publishIfReady(root)
	return &RouterGroup{host: host, routePath: full, isProxy: true}
}

// Middlewares 注册洋葱模型中间件：handler 内 ctx.Next() 控制后续链时机，
// 不调用也会自动继续。仅对其后注册的路由生效。
func (g *RouterGroup) Middlewares(handlers ...HandlerFunc) {
	host := g.requireHost()
	for _, h := range handlers {
		if h == nil {
			panic("[web] Middlewares: nil handler")
		}
	}
	if len(handlers) == 0 {
		return
	}
	host.mu.Lock()
	defer host.mu.Unlock()
	target := g
	if g.isProxy {
		var err error
		target, err = host.root.position(g.routePath)
		if err != nil {
			panic(formatPanic(g, "", err))
		}
	}
	if target.middlewares == nil {
		target.middlewares = make([]middleware, 0, len(handlers)+5)
	}
	for _, h := range handlers {
		host.globalCount++
		target.middlewares = append(target.middlewares, middleware{
			HandlerFunc: h,
			order:       host.globalCount,
		})
	}
	host.publishIfReady(host.root)
}

// HeadMiddleware 等价于 Middlewares(func(ctx){ h(ctx); ctx.Next() })。
func (g *RouterGroup) HeadMiddleware(handlers ...HandlerFunc) {
	if len(handlers) == 0 {
		return
	}
	wrapped := make([]HandlerFunc, len(handlers))
	for i, h := range handlers {
		if h == nil {
			panic("[web] HeadMiddleware: nil handler")
		}
		h := h
		wrapped[i] = func(ctx *Context) {
			h(ctx)
			ctx.Next()
		}
	}
	g.Middlewares(wrapped...)
}

// FootMiddleware 等价于 Middlewares(func(ctx){ ctx.Next(); h(ctx) })。
// 多个 FootMiddleware 之间为 LIFO（洋葱外层后跑）。
func (g *RouterGroup) FootMiddleware(handlers ...HandlerFunc) {
	if len(handlers) == 0 {
		return
	}
	wrapped := make([]HandlerFunc, len(handlers))
	for i, h := range handlers {
		if h == nil {
			panic("[web] FootMiddleware: nil handler")
		}
		h := h
		wrapped[i] = func(ctx *Context) {
			ctx.Next()
			h(ctx)
		}
	}
	g.Middlewares(wrapped...)
}

// splitSegments 切分路径并丢弃空段；"/" 与 "" 均得到空切片。
// 返回的切片有自己的底层数组，调用方修改不会影响输入字符串或其他调用方。
func splitSegments(path string) []string {
	if path == "" {
		return nil
	}
	start, end := 0, len(path)
	for start < end && path[start] == '/' {
		start++
	}
	for end > start && path[end-1] == '/' {
		end--
	}
	if start == end {
		return nil
	}
	s := path[start:end]
	cap := 1
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			cap++
		}
	}
	out := make([]string, 0, cap)
	last := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			if i > last {
				out = append(out, s[last:i])
			}
			last = i + 1
		}
	}
	if len(s) > last {
		out = append(out, s[last:])
	}
	return out
}

// normalizePathKey 把请求/注册路径归一为 flatRoutes 的 key。
// 公共情况（无连续斜杠）不会产生分配。
func normalizePathKey(path string) string {
	n := len(path)
	if n == 0 {
		return ""
	}
	start, end := 0, n
	for start < end && path[start] == '/' {
		start++
	}
	for end > start && path[end-1] == '/' {
		end--
	}
	if start == end {
		return ""
	}
	if !strings.Contains(path[start:end], "//") {
		return path[start:end]
	}
	return strings.Join(splitSegments(path), "/")
}

func commonSegmentPrefix(a, b []string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

func segmentsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (g *RouterGroup) findRoot() *RouterGroup {
	if g.isProxy && g.host != nil && g.host.root != nil {
		return g.host.root
	}
	n := g
	for n.parent != nil {
		n = n.parent
	}
	return n
}

func formatPanic(g *RouterGroup, path string, err error) string {
	return fmt.Sprintf("[web] %s%s: %v", g.completePath(), path, err)
}

// dryRunValidate 在不修改树的前提下校验 segs 的结构合法性，
// 让 position / handle 在错误时不会留下孤儿节点。
func dryRunValidate(start *RouterGroup, segs []string) error {
	node := start
	for i := 0; i < len(segs); {
		seg := segs[i]
		switch {
		case seg == catchAllSegment:
			if i != len(segs)-1 {
				return errors.New("#... 必须出现在路径末尾")
			}
			if node != nil && node.catchAllKid != nil {
				node = node.catchAllKid
			} else {
				node = nil
			}
			i++
		case strings.HasPrefix(seg, paramPrefix):
			if !isValidParamName(seg) {
				return fmt.Errorf("非法参数段: %q", seg)
			}
			if node != nil && node.paramKid != nil {
				if node.paramKid.segments[0] != seg {
					return fmt.Errorf("地址泛匹配重复: 已有 %s, 试图新增 %s",
						node.paramKid.segments[0], seg)
				}
				node = node.paramKid
			} else {
				node = nil
			}
			i++
		default:
			j := i
			for j < len(segs) &&
				!strings.HasPrefix(segs[j], paramPrefix) &&
				segs[j] != catchAllSegment {
				j++
			}
			if node != nil {
				node = walkStaticEdge(node, segs[i:j])
			}
			i = j
		}
	}
	return nil
}

func isValidParamName(seg string) bool {
	// 形如 "#name" 或 "#..."；"#" 只是前缀，参数名至少 1 个非斜杠字符。
	if len(seg) < 2 {
		return false
	}
	if seg == catchAllSegment {
		return true
	}
	for i := 1; i < len(seg); i++ {
		if seg[i] == '/' {
			return false
		}
	}
	return true
}

// walkStaticEdge 沿 node 的静态子节点尽可能走完 segs；走不通返回 nil。
// 仅用于 dry-run 跟踪「假如插入将落到哪个已有节点」，非匹配。
func walkStaticEdge(node *RouterGroup, segs []string) *RouterGroup {
	for len(segs) > 0 {
		if node == nil || node.staticKids == nil {
			return nil
		}
		c, ok := node.staticKids[segs[0]]
		if !ok {
			return nil
		}
		n := len(c.segments)
		if len(segs) < n || !segmentsEqual(segs[:n], c.segments) {
			return nil
		}
		node = c
		segs = segs[n:]
	}
	return node
}

// position 在树上行走/创建到 path 对应的节点；负责必要的边分裂。
// 调用前请先用 dryRunValidate 校验，确保走到这里不会因结构错误中途返回。
//
// 注意：不在此处分配 order；由 handle 在确认 method 不冲突后赋值，
// 这样 Grep / 中间件注册不会篡改已注册路由的 order。
func (g *RouterGroup) position(segs []string) (*RouterGroup, error) {
	node := g
	for i := 0; i < len(segs); {
		seg := segs[i]
		switch {
		case seg == catchAllSegment:
			if i != len(segs)-1 {
				return nil, errors.New("#... 必须出现在路径末尾")
			}
			if node.catchAllKid == nil {
				node.catchAllKid = &RouterGroup{
					kind:     kindCatchAll,
					segments: []string{catchAllSegment},
					parent:   node,
					host:     g.host,
				}
			}
			node = node.catchAllKid
			i++
		case strings.HasPrefix(seg, paramPrefix):
			if node.paramKid != nil {
				if node.paramKid.segments[0] != seg {
					return nil, errors.New("地址泛匹配重复")
				}
			} else {
				node.paramKid = &RouterGroup{
					kind:     kindParam,
					segments: []string{seg},
					parent:   node,
					host:     g.host,
				}
			}
			node = node.paramKid
			i++
		default:
			j := i
			for j < len(segs) &&
				!strings.HasPrefix(segs[j], paramPrefix) &&
				segs[j] != catchAllSegment {
				j++
			}
			node = node.upsertStatic(segs[i:j])
			i = j
		}
	}
	return node, nil
}

// upsertStatic 把一串纯静态段插入到 g 之下；必要时对既有边做段级分裂。
func (g *RouterGroup) upsertStatic(segs []string) *RouterGroup {
	if len(segs) == 0 {
		return g
	}
	if g.staticKids == nil {
		g.staticKids = make(map[string]*RouterGroup, 4)
	}
	first := segs[0]
	child, ok := g.staticKids[first]
	if !ok {
		node := &RouterGroup{
			kind:     kindStatic,
			segments: append([]string(nil), segs...),
			parent:   g,
			host:     g.host,
		}
		g.staticKids[first] = node
		return node
	}
	common := commonSegmentPrefix(child.segments, segs)
	if common == len(child.segments) {
		if common == len(segs) {
			return child
		}
		return child.upsertStatic(segs[common:])
	}

	// 需要分裂：把 child 既有载荷搬到新建的 inner 节点（承接旧后缀），
	// 把 child 自身降级为前缀节点。
	inner := &RouterGroup{
		kind:        kindStatic,
		segments:    append([]string(nil), child.segments[common:]...),
		parent:      child,
		staticKids:  child.staticKids,
		paramKid:    child.paramKid,
		catchAllKid: child.catchAllKid,
		method:      child.method,
		methodOrder: child.methodOrder,
		options:     child.options,
		middlewares: child.middlewares,
		order:       child.order,
		host:        child.host,
	}
	for _, c := range inner.staticKids {
		c.parent = inner
	}
	if inner.paramKid != nil {
		inner.paramKid.parent = inner
	}
	if inner.catchAllKid != nil {
		inner.catchAllKid.parent = inner
	}
	child.segments = append([]string(nil), child.segments[:common]...)
	child.staticKids = map[string]*RouterGroup{inner.segments[0]: inner}
	child.paramKid = nil
	child.catchAllKid = nil
	child.method = nil
	child.methodOrder = nil
	child.options = nil
	child.middlewares = nil
	child.order = 0

	if common == len(segs) {
		return child
	}
	return child.upsertStatic(segs[common:])
}

// completePath 上溯所有祖先 segments 拼出完整路径。根节点返回 ""。
func (g *RouterGroup) completePath() string {
	if g.isProxy {
		return pathFromSegments(g.routePath)
	}
	if g.parent == nil {
		return ""
	}
	var total int
	for n := g; n.parent != nil; n = n.parent {
		for _, s := range n.segments {
			total += len(s) + 1
		}
	}
	var groups [][]string
	for n := g; n.parent != nil; n = n.parent {
		groups = append(groups, n.segments)
	}
	var sb strings.Builder
	sb.Grow(total)
	for i := len(groups) - 1; i >= 0; i-- {
		for _, s := range groups[i] {
			sb.WriteByte('/')
			sb.WriteString(s)
		}
	}
	return sb.String()
}

func (g *RouterGroup) handle(method, path string, handlerFunc HandlerFunc, opt ...HandlerOpt) {
	if handlerFunc == nil {
		panic("[web] handler must not be nil: " + method + " " + path)
	}
	host := g.requireHost()
	host.mu.Lock()
	defer host.mu.Unlock()

	root := host.root
	segs := splitSegments(path)
	full := append(g.baseRoutePath(), segs...)
	if err := dryRunValidate(root, full); err != nil {
		panic(formatPanic(g, path, err))
	}
	leaf, err := root.position(full)
	if err != nil {
		panic(formatPanic(g, path, err))
	}

	if method == ANY {
		if len(leaf.method) > 0 {
			panic(formatPanic(g, path, errors.New("已存在其它方法，无法注册 ANY")))
		}
	} else {
		if _, ok := leaf.method[ANY]; ok {
			panic(formatPanic(g, path, errors.New("该路由 ANY 方法冲突")))
		}
		if _, ok := leaf.method[method]; ok {
			panic(formatPanic(g, path, fmt.Errorf("该路由 method 重复: %s", method)))
		}
	}
	if leaf.method == nil {
		leaf.method = make(map[string]HandlerFunc, 4)
	}
	if leaf.methodOrder == nil {
		leaf.methodOrder = make(map[string]uint32, 4)
	}
	leaf.method[method] = handlerFunc
	if len(opt) > 0 {
		if leaf.options == nil {
			leaf.options = make(map[string]HandlerOpt)
		}
		leaf.options[method] = opt[0]
	}
	host.globalCount++
	leaf.methodOrder[method] = host.globalCount
	leaf.order = host.globalCount

	host.publishIfReady(root)
}

func pathFromSegments(segs []string) string {
	if len(segs) == 0 {
		return ""
	}
	var b strings.Builder
	for _, seg := range segs {
		b.WriteByte('/')
		b.WriteString(seg)
	}
	return b.String()
}

func (g *RouterGroup) baseRoutePath() []string {
	if g.isProxy {
		return append([]string(nil), g.routePath...)
	}
	if g.parent == nil {
		return nil
	}
	return splitSegments(g.completePath())
}

// PathMatch 是公开 API：基于已发布的快照做匹配。当前 RouterGroup 不在
// 任何 registry 下（裸节点 / 快照节点）时返回 (nil, nil, nil)。
//
// ServeHTTP 不直接走这里，而是直接使用 snapshot.match 以拿到 Allow 信息。
func (g *RouterGroup) PathMatch(path, method string) (params map[string]string, handle HandlerFunc, grep *RouterGroup) {
	if g.host == nil {
		return nil, nil, nil
	}
	snap := g.host.snapshot.Load()
	if snap == nil {
		return nil, nil, nil
	}
	if idx := strings.IndexByte(path, '?'); idx >= 0 {
		path = path[:idx]
	}
	m, h, leaf, _ := snap.match(path, method)
	return m, h, leaf
}

// match 在快照上为请求查找处理器。
//
// 返回 4 元组：
//   - params：动态参数表，纯静态命中时为 nil。
//   - handle：命中的 handler；若路径找到但 method 不匹配，handle 为 nil 但 allowed 非空（用于 405）。
//   - leaf  ：命中的节点；middleware 遍历需要它的 parent 链。
//   - allowed：仅当路径找到、method 全部不匹配时填充该叶子的方法集。
func (snap *routeSnapshot) match(path, method string) (params map[string]string, handle HandlerFunc, leaf *RouterGroup, allowed methodSet) {
	var scratch [contextParamBufferSize]pathParam
	matched, handle, leaf, allowed := snap.matchInto(path, method, scratch[:0])
	if len(matched) > 0 {
		params = make(map[string]string, len(matched))
		for _, param := range matched {
			params[param.key] = param.value
		}
	}
	return params, handle, leaf, allowed
}

// matchInto 是请求路径匹配的无 map 版本。params 通常直接指向 Context
// 内置的 8 项缓冲；回溯只会在需要时扩容，并在失败分支清除字符串引用。
func (snap *routeSnapshot) matchInto(path, method string, params []pathParam) (matched []pathParam, handle HandlerFunc, leaf *RouterGroup, allowed methodSet) {
	if snap == nil || snap.root == nil {
		return params, nil, nil, nil
	}
	path = normalizePathKey(path)

	// 纯静态快速通道：命中即返回；纯静态环境且未命中可直接判 404。
	if snap.flatRoutes != nil {
		if l, ok := snap.flatRoutes[path]; ok {
			handle = l.method[method]
			if handle == nil {
				handle = l.method[ANY]
			}
			leaf = l
			if handle == nil {
				allowed = methodSet(l.method)
			}
			return params, handle, leaf, allowed
		}
		if !snap.hasDynamic {
			return params, nil, nil, nil
		}
	}

	node, params, ok := matchPathNormalized(snap.root, path, 0, params)
	if !ok {
		return params, nil, nil, nil
	}
	leaf = node
	handle = node.method[method]
	if handle == nil {
		handle = node.method[ANY]
	}
	if handle == nil {
		allowed = methodSet(node.method)
	}
	return params, handle, leaf, allowed
}

// pathSegment returns the next normalized path segment without constructing a
// substring. next is positioned at the beginning of the following segment.
func pathSegment(path string, pos int) (start, end, next int) {
	start = pos
	end = pos
	for end < len(path) && path[end] != '/' {
		end++
	}
	next = end
	if next < len(path) {
		next++
	}
	return start, end, next
}

func matchStaticEdge(child *RouterGroup, path string, pos int) (next int, ok bool) {
	next = pos
	for _, segment := range child.segments {
		if next >= len(path) {
			return pos, false
		}
		start, end, after := pathSegment(path, next)
		if path[start:end] != segment {
			return pos, false
		}
		next = after
	}
	return next, true
}

func clearPathParams(params []pathParam) {
	for i := range params {
		params[i] = pathParam{}
	}
}

// matchPathNormalized walks a normalized path by byte positions. Static
// edges are compared directly against the path and dynamic values retain
// substrings of that path, so the common request path does not need a segment
// slice, strings.Join, or parameter map.
func matchPathNormalized(node *RouterGroup, path string, pos int, params []pathParam) (*RouterGroup, []pathParam, bool) {
	if pos == len(path) {
		if len(node.method) > 0 {
			return node, params, true
		}
		if node.catchAllKid != nil && len(node.catchAllKid.method) > 0 {
			params = append(params, pathParam{key: paramPrefix, value: path[pos:]})
			return node.catchAllKid, params, true
		}
		return nil, params, false
	}

	start, end, next := pathSegment(path, pos)
	if child, ok := node.staticKids[path[start:end]]; ok {
		if edgeNext, edgeOK := matchStaticEdge(child, path, pos); edgeOK {
			if leaf, out, matched := matchPathNormalized(child, path, edgeNext, params); matched {
				return leaf, out, true
			} else {
				params = out
			}
		}
	}

	if node.paramKid != nil {
		base := len(params)
		params = append(params, pathParam{
			key:   node.paramKid.segments[0][1:],
			value: path[start:end],
		})
		if leaf, out, matched := matchPathNormalized(node.paramKid, path, next, params); matched {
			return leaf, out, true
		} else {
			clearPathParams(out[base:])
			params = out[:base]
		}
	}

	if node.catchAllKid != nil && len(node.catchAllKid.method) > 0 {
		params = append(params, pathParam{key: paramPrefix, value: path[pos:]})
		return node.catchAllKid, params, true
	}
	return nil, params, false
}

// matchPath 按静态 > 参数 > 通配的优先级做深度优先匹配；
// 某个高优先级分支在后续段失败时回溯到下一个候选分支。
func matchPath(node *RouterGroup, segs []string, index int, params map[string]string) (*RouterGroup, map[string]string, bool) {
	if index == len(segs) {
		if len(node.method) > 0 {
			return node, params, true
		}
		if node.catchAllKid != nil {
			if len(node.catchAllKid.method) == 0 {
				return nil, nil, false
			}
			if params == nil {
				params = make(map[string]string, 4)
			}
			params[paramPrefix] = ""
			return node.catchAllKid, params, true
		}
		return nil, nil, false
	}

	if child, ok := node.staticKids[segs[index]]; ok {
		n := len(child.segments)
		if index+n <= len(segs) && segmentsEqual(segs[index:index+n], child.segments) {
			if leaf, out, matched := matchPath(child, segs, index+n, params); matched {
				return leaf, out, true
			}
		}
	}

	if node.paramKid != nil {
		name := node.paramKid.segments[0][1:]
		if params == nil {
			params = make(map[string]string, 4)
		}
		old, existed := params[name]
		params[name] = segs[index]
		if leaf, out, matched := matchPath(node.paramKid, segs, index+1, params); matched {
			return leaf, out, true
		}
		if existed {
			params[name] = old
		} else {
			delete(params, name)
		}
	}

	if node.catchAllKid != nil && len(node.catchAllKid.method) > 0 {
		if params == nil {
			params = make(map[string]string, 4)
		}
		params[paramPrefix] = strings.Join(segs[index:], "/")
		return node.catchAllKid, params, true
	}
	return nil, nil, false
}

// publish 深拷贝写端树为只读快照并 atomic 替换。整个过程必须在 mu.Lock 中调用。
func (r *registry) publish(writeRoot *RouterGroup) {
	newRoot := cloneNode(writeRoot, nil)
	compileHandlerChains(newRoot)

	flatRoutes := make(map[string]*RouterGroup)
	var hasDynamic bool
	var walk func(n *RouterGroup, dyn bool, path []string)
	walk = func(n *RouterGroup, dyn bool, path []string) {
		if n.kind == kindParam || n.kind == kindCatchAll {
			dyn = true
			hasDynamic = true
		}
		if n.parent != nil && len(n.segments) > 0 {
			path = append(append([]string(nil), path...), n.segments...)
		}
		if len(n.method) > 0 && !dyn {
			flatRoutes[strings.Join(path, "/")] = n
		}
		for _, c := range n.staticKids {
			walk(c, dyn, path)
		}
		if n.paramKid != nil {
			walk(n.paramKid, dyn, path)
		}
		if n.catchAllKid != nil {
			walk(n.catchAllKid, dyn, path)
		}
	}
	walk(newRoot, false, nil)

	r.snapshot.Store(&routeSnapshot{
		root:       newRoot,
		flatRoutes: flatRoutes,
		hasDynamic: hasDynamic,
	})
}

// compileHandlerChains materializes the exact middleware chain for every
// registered method in a snapshot. The write tree remains order-aware, while
// request handling can select an immutable slice directly.
func compileHandlerChains(root *RouterGroup) {
	if root == nil {
		return
	}
	var walk func(*RouterGroup)
	walk = func(node *RouterGroup) {
		if len(node.method) > 0 {
			node.handlerChains = make(map[string][]HandlerFunc, len(node.method))
			for method, handler := range node.method {
				methodOrder := node.methodOrder[method]
				var ancestors []*RouterGroup
				for ancestor := node; ancestor != nil; ancestor = ancestor.parent {
					ancestors = append(ancestors, ancestor)
				}

				chainLen := 1
				for i := len(ancestors) - 1; i >= 0; i-- {
					for _, middleware := range ancestors[i].middlewares {
						if middleware.order < methodOrder {
							chainLen++
						}
					}
				}
				chain := make([]HandlerFunc, 0, chainLen)
				for i := len(ancestors) - 1; i >= 0; i-- {
					for _, middleware := range ancestors[i].middlewares {
						if middleware.order < methodOrder {
							chain = append(chain, middleware.HandlerFunc)
						}
					}
				}
				chain = append(chain, handler)
				node.handlerChains[method] = chain
			}
		}

		for _, child := range node.staticKids {
			walk(child)
		}
		if node.paramKid != nil {
			walk(node.paramKid)
		}
		if node.catchAllKid != nil {
			walk(node.catchAllKid)
		}
	}
	walk(root)
}

func (r *registry) publishIfReady(writeRoot *RouterGroup) {
	if !r.initializing {
		r.publish(writeRoot)
	}
}

// cloneNode 深拷贝节点（含 staticKids/paramKid/catchAllKid 子树），重新建立 parent 链。
// 新节点 host 设为 nil，禁止运行期被改写。
func cloneNode(orig, newParent *RouterGroup) *RouterGroup {
	if orig == nil {
		return nil
	}
	n := &RouterGroup{
		kind:     orig.kind,
		segments: append([]string(nil), orig.segments...),
		parent:   newParent,
		order:    orig.order,
		host:     nil,
	}
	if orig.method != nil {
		n.method = make(map[string]HandlerFunc, len(orig.method))
		for k, v := range orig.method {
			n.method[k] = v
		}
	}
	if orig.methodOrder != nil {
		n.methodOrder = make(map[string]uint32, len(orig.methodOrder))
		for k, v := range orig.methodOrder {
			n.methodOrder[k] = v
		}
	}
	if orig.options != nil {
		n.options = make(map[string]HandlerOpt, len(orig.options))
		for k, v := range orig.options {
			n.options[k] = v
		}
	}
	if len(orig.middlewares) > 0 {
		n.middlewares = append([]middleware(nil), orig.middlewares...)
	}
	if orig.staticKids != nil {
		n.staticKids = make(map[string]*RouterGroup, len(orig.staticKids))
		for k, child := range orig.staticKids {
			n.staticKids[k] = cloneNode(child, n)
		}
	}
	if orig.paramKid != nil {
		n.paramKid = cloneNode(orig.paramKid, n)
	}
	if orig.catchAllKid != nil {
		n.catchAllKid = cloneNode(orig.catchAllKid, n)
	}
	return n
}

// findNode resolves a logical route path without modifying the compressed
// write tree. It is used by read-only helpers on Grep proxies.
func findNode(root *RouterGroup, segs []string) *RouterGroup {
	node := root
	for i := 0; i < len(segs); {
		if child, ok := node.staticKids[segs[i]]; ok {
			n := len(child.segments)
			if i+n > len(segs) || !segmentsEqual(segs[i:i+n], child.segments) {
				return nil
			}
			node = child
			i += n
			continue
		}
		if node.paramKid != nil && segs[i] == node.paramKid.segments[0] {
			node = node.paramKid
			i++
			continue
		}
		if node.catchAllKid != nil && segs[i] == catchAllSegment && i+1 == len(segs) {
			return node.catchAllKid
		}
		return nil
	}
	return node
}

// BottomNodeList 返回不含任何子节点的叶子节点（写端视图）。
//
// 注意：仅在「注册阶段已完成」或调用方自行同步时使用；与并发注册混用会产生竞态。
func (g *RouterGroup) BottomNodeList() []*RouterGroup {
	if g.host != nil {
		g.host.mu.Lock()
		defer g.host.mu.Unlock()
	}
	if g.isProxy {
		if target := findNode(g.host.root, g.routePath); target != nil {
			g = target
		}
	}
	var out []*RouterGroup
	g.walkLeaves(&out)
	return out
}

func (g *RouterGroup) walkLeaves(out *[]*RouterGroup) {
	if len(g.staticKids) == 0 && g.paramKid == nil && g.catchAllKid == nil {
		if g.parent != nil {
			*out = append(*out, g)
		}
		return
	}
	for _, c := range g.staticKids {
		c.walkLeaves(out)
	}
	if g.paramKid != nil {
		g.paramKid.walkLeaves(out)
	}
	if g.catchAllKid != nil {
		g.catchAllKid.walkLeaves(out)
	}
}

type RouterInfo struct {
	Method string
	Path   string
	Option HandlerOpt
}

// List 列出所有已注册的路由（包括非叶子节点上的）；输出按 (path, method) 字典序排序。
func (g *RouterGroup) List() (infos []RouterInfo) {
	if g.host != nil {
		g.host.mu.Lock()
		defer g.host.mu.Unlock()
	}
	if g.isProxy {
		if target := findNode(g.host.root, g.routePath); target != nil {
			g = target
		}
	}
	type entry struct {
		method, path string
		opt          HandlerOpt
	}
	var entries []entry
	g.walkAll(func(n *RouterGroup) {
		if len(n.method) == 0 {
			return
		}
		p := n.completePath()
		if p == "" {
			p = "/"
		}
		for m := range n.method {
			entries = append(entries, entry{method: m, path: p, opt: n.options[m]})
		}
	})
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].path != entries[j].path {
			return entries[i].path < entries[j].path
		}
		return entries[i].method < entries[j].method
	})
	infos = make([]RouterInfo, len(entries))
	for i, e := range entries {
		infos[i] = RouterInfo{
			Method: e.method,
			Path:   e.path,
			Option: e.opt,
		}
	}
	return
}

func (g *RouterGroup) walkAll(fn func(*RouterGroup)) {
	fn(g)
	for _, c := range g.staticKids {
		c.walkAll(fn)
	}
	if g.paramKid != nil {
		g.paramKid.walkAll(fn)
	}
	if g.catchAllKid != nil {
		g.catchAllKid.walkAll(fn)
	}
}
