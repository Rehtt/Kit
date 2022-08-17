package server

import (
	"context"
	"github.com/Rehtt/Kit/buf"
	"github.com/Rehtt/Kit/multiplex"
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

func (e *Engine) Run(tcpMultiplex ...bool) (err error) {
	var listener net.Listener
	if len(tcpMultiplex) > 0 && tcpMultiplex[0] {
		listenerConf := multiplex.NetLister()
		listener, err = listenerConf.Listen(context.Background(), e.network, e.addr)
	} else {
		listener, err = net.Listen(e.network, e.addr)
	}
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
