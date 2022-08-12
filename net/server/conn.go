package server

import (
	"github.com/Rehtt/Kit/buf"
	"net"
	"time"
)

const (
	write = iota
	read
)

type Context struct {
	conn  net.Conn
	buf   *buf.Buf
	write *buf.Buf
	e     *Engine
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
	err = ctx.Send()
	return
}
func (ctx *Context) Send() (err error) {
	if err = ctx.e.middle.useMiddleware(ctx, write); err != nil {
		return err
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
	return ctx.conn.Close()
}
