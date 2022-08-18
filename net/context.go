package gonet

import (
	"context"
	"github.com/Rehtt/Kit/buf"
	"net"
	"sync"
)

type Context struct {
	context context.Context
	close   context.CancelFunc
	read    *buf.Buf
	write   *buf.Buf
	conn    net.Conn
	pconn   net.PacketConn
	flag    uint8
}

const (
	nConn uint8 = iota
	pConn
)

var (
	contextPool = sync.Pool{
		New: func() interface{} {
			return new(Context)
		},
	}
)

func (c *Context) Write(b []byte) (n int, err error) {
	if c.isDone() {
		return 0, c.context.Err()
	}
	if n, err = c.write.Write(b); err != nil {
		return 0, err
	}
	if err = getMiddle(c.context).use(c, write); err != nil {
		return 0, err
	}

	// todo fix
	switch c.flag {
	case nConn:
		_, err = c.write.WriteTo(c.conn)
	case pConn:
		_, err = c.write.WriteTo(c.pconn)
	}

	return
}
func (c *Context) Read(b []byte) (n int, err error) {
	if c.isDone() {
		return 0, c.context.Err()
	}
	return c.read.Read(b)
}
func (c *Context) ReadToBytes() (b []byte, err error) {
	if c.isDone() {
		return nil, c.context.Err()
	}
	b = c.read.ToBytes()
	c.read.Reset()
	return b, nil
}
func (c *Context) Close() error {
	c.close()
	return nil
}
func (c *Context) isDone() bool {
	select {
	case <-c.context.Done():
		return true
	default:
		return false
	}
}

func getMiddle(ctx context.Context) *middle {
	return ctx.Value("&middle").(*middle)
}

func newContext(conn interface{}) *Context {
	c := contextPool.Get().(*Context)
	ctx, cancel := context.WithCancel(context.Background())
	c.context = ctx
	c.close = cancel
	c.read = buf.NewBuf()
	c.write = buf.NewBuf()
	switch conn := conn.(type) {
	case net.Conn:
		c.conn = conn
		c.flag = nConn
	case net.PacketConn:
		c.pconn = conn
		c.flag = pConn
	}
	return c
}
func delContext(ctx *Context) {
	ctx.Close()
	contextPool.Put(ctx)
}
