package client

import (
	"github.com/Rehtt/Kit/buf"
	"github.com/Rehtt/Kit/multiplex"
	"log"
	"net"
	"sync"
	"time"
)

type Context struct {
	network    string
	remoteAddr string
	localAddr  string
	Middle     Middle
	conn       net.Conn
	buf        *buf.Buf
	write      *buf.Buf
	lock       sync.WaitGroup
	OnResponse func(ctx *Context)
	net.Dialer
}

func New(network, laddr, raddr string, tcpMultiplex ...bool) (e *Context, err error) {
	e = &Context{
		network:    network,
		remoteAddr: raddr,
		localAddr:  laddr,
		write:      buf.NewBuf(),
		buf:        buf.NewBuf(),
	}
	m := true
	if len(tcpMultiplex) != 0 {
		m = tcpMultiplex[0]
	}
	if m {
		dial := multiplex.NetDialer()
		if e.localAddr != "" {
			var l net.Addr
			switch e.network {
			case "tcp", "tcp4", "tcp6":
				l, err = net.ResolveTCPAddr(e.network, e.localAddr)
			case "udp", "udp4", "udp6":
				l, err = net.ResolveUDPAddr(e.network, e.localAddr)
			}
			if err != nil {
				return nil, err
			}
			dial.LocalAddr = l
		}

		e.Dialer = *dial

	} else {
		e.Dialer = net.Dialer{}
	}
	return e, nil
}
func (ctx *Context) Dial() (err error) {
	ctx.conn, err = ctx.Dialer.Dial(ctx.network, ctx.remoteAddr)
	if err != nil {
		return err
	}
	go func(ctx *Context) {
		defer ctx.ConnClose()
		for {
			// 读取包内所有数据
			err = ctx.readBody()
			if err != nil {
				log.Println(err)
				return
			}
			err = ctx.Middle.useMiddleware(ctx, read)
			if err != nil {
				log.Println(err)
				return
			}
			if ctx.OnResponse != nil {
				ctx.OnResponse(ctx)
			}
		}
	}(ctx)
	ctx.lock.Add(1)
	return
}
func (ctx *Context) Body() *buf.Buf {
	return ctx.buf
}
func (ctx *Context) Read(b []byte) (n int, err error) {
	return ctx.buf.Read(b)
}
func (ctx *Context) Write(b []byte) (n int, err error) {
	n, err = ctx.write.Write(b)
	if err != nil {
		return 0, err
	}
	if err = ctx.Middle.useMiddleware(ctx, write); err != nil {
		return 0, err
	}
	_, err = ctx.conn.Write(ctx.write.ToBytes())
	ctx.write.Reset()
	return
}
func (ctx *Context) readBody() error {
	// 读取包内所有数据
	ctx.buf.Reset()
	var tmp = make([]byte, 512)
	for {
		n, err := ctx.conn.Read(tmp)
		if err != nil || n == 0 {
			return err
		}
		ctx.buf.WriteBytes(tmp[:n])
		if n < 512 {
			break
		}
	}
	return nil
}
func (ctx *Context) close() error {
	ctx.buf.Free()
	return ctx.conn.Close()
}

func (ctx *Context) RemoteAddr() net.Addr {
	return ctx.conn.RemoteAddr()
}
func (ctx *Context) SetDeadline(t time.Time) error {
	return ctx.conn.SetDeadline(t)
}
func (ctx *Context) SetReadDeadline(t time.Time) error {
	return ctx.conn.SetReadDeadline(t)
}
func (ctx *Context) SetWriteDeadline(t time.Time) error {
	return ctx.conn.SetWriteDeadline(t)
}
func (ctx *Context) ConnClose() error {
	ctx.lock.Done()
	return ctx.conn.Close()
}
func (ctx *Context) Wait() {
	ctx.lock.Wait()
}
