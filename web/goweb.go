/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/7/16 下午 04:30
 */

package web

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// 默认 Server 超时；用户可通过 g.Server 字段覆盖或 WithServer 替换。
const (
	defaultReadHeaderTimeout = 5 * time.Second
	defaultReadTimeout       = 30 * time.Second
	defaultWriteTimeout      = 30 * time.Second
	defaultIdleTimeout       = 60 * time.Second
	defaultShutdownTimeout   = 30 * time.Second
)

type GOweb struct {
	RouterGroup

	Server          *http.Server
	noRouter        HandlerFunc
	onPanic         func(*Context, any)
	shutdownTimeout time.Duration

	context.Context
}

var contextPool = sync.Pool{
	New: func() any {
		return new(Context)
	},
}

func (g *GOweb) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var snap *routeSnapshot
	if g.host != nil {
		snap = g.host.snapshot.Load()
	}

	rctx, cancel := context.WithCancel(request.Context())

	ctx := contextPool.Get().(*Context)
	ctx.rw.reset(writer)
	ctx.Request = request
	ctx.Writer = &ctx.rw
	ctx.Context = rctx
	ctx.cancel = cancel
	ctx.values = g.Context
	ctx.param = nil
	ctx.index = 0
	if ctx.handlers != nil {
		ctx.handlers = ctx.handlers[:0]
	}

	defer func() {
		rec := recover()
		// ErrAbortHandler 是 stdlib 内部协议，必须继续向上抛。
		if rec != nil && rec != http.ErrAbortHandler {
			g.runPanicHandler(ctx, rec)
		}

		if ctx.cancel != nil {
			ctx.cancel()
		}
		ctx.Request = nil
		ctx.Writer = nil
		ctx.Context = nil
		ctx.cancel = nil
		ctx.values = nil
		ctx.param = nil
		ctx.rw.reset(nil)
		if ctx.handlers != nil {
			// 清空 handler 引用，避免 pool 长期持有闭包。
			for i := range ctx.handlers {
				ctx.handlers[i] = nil
			}
			ctx.handlers = ctx.handlers[:0]
		}
		ctx.index = 0
		contextPool.Put(ctx)

		if rec == http.ErrAbortHandler {
			panic(rec)
		}
	}()

	if snap == nil {
		http.NotFound(ctx.Writer, request)
		return
	}

	method := request.Method
	// 用 URL.Path（已解码），不用 RequestURI。
	params, handleFunc, leaf, allowed := snap.match(request.URL.Path, method)
	matchedMethod := method
	if handleFunc != nil && leaf.method[method] == nil {
		matchedMethod = ANY
	}

	// HEAD 未注册时回落 GET（RFC 9110 §9.3.2）。
	if handleFunc == nil && method == http.MethodHead && leaf != nil {
		if h := leaf.method[http.MethodGet]; h != nil {
			handleFunc = h
			matchedMethod = http.MethodGet
			allowed = nil
		} else if h := leaf.method[ANY]; h != nil {
			handleFunc = h
			matchedMethod = ANY
			allowed = nil
		}
	}

	if handleFunc == nil {
		if !allowed.empty() {
			ctx.Writer.Header().Set("Allow", allowed.headerValue())
			ctx.Writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		g.handler404(ctx)
		return
	}

	ctx.param = params

	handleFuncOrder := leaf.methodOrder[matchedMethod]

	// 收集祖先链 leaf→root，再倒序展开为 root→leaf 拼 chain。栈数组避免堆分配。
	var stackBuf [16]*RouterGroup
	ancestors := stackBuf[:0]
	for gp := leaf; gp != nil; gp = gp.parent {
		ancestors = append(ancestors, gp)
	}

	chain := ctx.handlers
	for i := len(ancestors) - 1; i >= 0; i-- {
		gp := ancestors[i]
		for j := range gp.middlewares {
			if gp.middlewares[j].order < handleFuncOrder {
				chain = append(chain, gp.middlewares[j].HandlerFunc)
			}
		}
	}
	chain = append(chain, handleFunc)

	ctx.handlers = chain
	ctx.index = -1
	ctx.Next()
}

// runPanicHandler 调用用户钩子；钩子自身 panic 也要 recover，避免污染上层 defer。
func (g *GOweb) runPanicHandler(ctx *Context, rec any) {
	defer func() { _ = recover() }()
	if g.onPanic != nil {
		g.onPanic(ctx, rec)
		return
	}
	defaultOnPanic(ctx, rec)
}

// defaultOnPanic 打印堆栈；header 未发时尽力写 500。
func defaultOnPanic(ctx *Context, rec any) {
	const stackSize = 64 << 10
	buf := make([]byte, stackSize)
	n := runtime.Stack(buf, false)
	if ctx != nil && ctx.Request != nil {
		log.Printf("[web] panic recovered %s %s: %v\n%s",
			ctx.Request.Method, ctx.Request.URL.Path, rec, buf[:n])
	} else {
		log.Printf("[web] panic recovered: %v\n%s", rec, buf[:n])
	}

	if ne := netErrFromRecover(rec); ne != nil {
		return
	}
	if ctx == nil || ctx.rw.ResponseWriter == nil || ctx.rw.Written() {
		return
	}
	http.Error(ctx.rw.ResponseWriter, "Internal Server Error", http.StatusInternalServerError)
}

func netErrFromRecover(rec any) error {
	if err, ok := rec.(error); ok {
		var ne net.Error
		if errors.As(err, &ne) {
			return ne
		}
	}
	return nil
}

// OnPanic 注册自定义 panic 钩子；nil 恢复默认。
func (g *GOweb) OnPanic(fn func(*Context, any)) {
	g.onPanic = fn
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
	g.shutdownTimeout = defaultShutdownTimeout
	reg := &registry{}
	g.RouterGroup.host = reg
	reg.root = &g.RouterGroup
	reg.initializing = true
	g.Context = context.Background()
	g.Server = &http.Server{
		Handler:           g,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		ReadTimeout:       defaultReadTimeout,
		WriteTimeout:      defaultWriteTimeout,
		IdleTimeout:       defaultIdleTimeout,
	}
	for _, opt := range opts {
		opt(g)
	}
	// Option 可能把 Server 置 nil 或换掉，保证最终 Handler 指向自己。
	if g.Server == nil {
		g.Server = &http.Server{
			Handler:           g,
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			ReadTimeout:       defaultReadTimeout,
			WriteTimeout:      defaultWriteTimeout,
			IdleTimeout:       defaultIdleTimeout,
		}
	} else if g.Server.Handler == nil {
		g.Server.Handler = g
	}
	reg.initializing = false
	reg.publish(&g.RouterGroup)
	return
}

func (g *GOweb) SetValue(key, value any) {
	g.Context = context.WithValue(g.Context, key, value)
}

func (g *GOweb) GetValue(key any) any {
	return g.Value(key)
}

// Run 阻塞直到 Server 关闭。正常关停返回 http.ErrServerClosed。
func (g *GOweb) Run(addr string) error {
	g.Server.Addr = addr
	return g.Server.ListenAndServe()
}

func (g *GOweb) RunTLS(addr, certFile, keyFile string) error {
	g.Server.Addr = addr
	return g.Server.ListenAndServeTLS(certFile, keyFile)
}

// RunContext 在 ctx 取消或到期时使用独立的关停窗口（默认 30 秒）。
// WithShutdownTimeout 可配置该窗口；超时返回关停错误，不强制关闭活跃连接。
// 关停成功后返回服务退出结果，通常为 http.ErrServerClosed。
func (g *GOweb) RunContext(ctx context.Context, addr string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	g.Server.Addr = addr
	errCh := make(chan error, 1)
	go func() { errCh <- g.Server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), g.shutdownTimeout)
		shutdownErr := g.Server.Shutdown(shutdownCtx)
		cancel()
		if shutdownErr != nil {
			return shutdownErr
		}
		return <-errCh
	case err := <-errCh:
		return err
	}
}

// Shutdown 使用调用方 ctx 控制关停，不使用 WithShutdownTimeout 配置。
func (g *GOweb) Shutdown(ctx context.Context) error {
	if g.Server == nil {
		return nil
	}
	return g.Server.Shutdown(ctx)
}
