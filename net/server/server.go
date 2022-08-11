package server

import (
	"github.com/Rehtt/Kit/buf"
	"log"
	"net"
)

type Engine struct {
	network string
	addr    string
	middle  Middle
	Handle  func(ctx *Context)
}

func New(network, addr string) *Engine {
	return &Engine{
		network: network,
		addr:    addr,
	}
}

func (e *Engine) Run() error {
	listener, err := net.Listen(e.network, e.addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		// 开始处理
		go e.handle(&Context{
			conn:  conn,
			e:     e,
			buf:   buf.NewBuf(),
			write: buf.NewBuf(),
		})
	}
}

// 做中间转换支持中间件
func (e *Engine) handle(ctx *Context) {
	for {
		// 读取包内所有数据
		err := ctx.readBody()
		if err != nil {
			log.Println(err)
			return
		}
		err = e.middle.useMiddleware(ctx, read)
		if err != nil {
			log.Println(err)
			return
		}
		if e.Handle != nil {
			e.Handle(ctx)
		}
	}
}
