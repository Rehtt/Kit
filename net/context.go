package gonet

import (
	"context"
	"github.com/Rehtt/Kit/buf"
	"net"
	"sync"
)

type Context struct {
	context  context.Context
	close    context.CancelFunc
	read     *buf.Buf
	write    *buf.Buf
	conn     interface{}
	readFlag bool
}

const (
	cacheSize = 512
)
const (
	RemoteAddr = "&remoteAddr"
	LocalAddr  = "&localAddr"
	Middle     = "&middle"
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

	switch conn := c.conn.(type) {
	case net.Conn:
		_, err = c.write.WriteTo(conn)
	case *PacketConn:
		_, err = conn.conn.WriteTo(b, c.RemoteAddr())
	}
	c.write.Reset()

	return
}
func (c *Context) Read(b []byte) (n int, err error) {
	if c.isDone() {
		return 0, c.context.Err()
	}
	if c.read.Len() == 0 {
		if err = c.readAll(); err != nil {
			return 0, err
		}
	}
	return c.read.Read(b)
}
func (c *Context) ReadToBytes() (b []byte, err error) {
	if c.isDone() {
		return nil, c.context.Err()
	}
	if c.read.Len() == 0 {
		if err = c.readAll(); err != nil {
			return nil, err
		}
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

func (c *Context) setValue(key string, value interface{}) {
	c.context = context.WithValue(c.context, key, value)
}

func getMiddle(ctx context.Context) *middle {
	return ctx.Value(Middle).(*middle)
}
func (c *Context) LocalAddr() net.Addr {
	return c.context.Value(LocalAddr).(net.Addr)
}
func (c *Context) RemoteAddr() net.Addr {
	return c.context.Value(RemoteAddr).(net.Addr)
}

func newContext(conn interface{}) *Context {
	c := contextPool.Get().(*Context)
	ctx, cancel := context.WithCancel(context.Background())
	c.context = ctx
	c.close = cancel
	c.read = buf.NewBuf()
	c.write = buf.NewBuf()
	c.conn = conn
	return c
}
func delContext(ctx *Context) {
	ctx.Close()
	contextPool.Put(ctx)
}

// todo Deadline
//func (c *Context) SetReadDeadline(t time.Time)error {
//
//}
//func (c *Context) SetDeadline(t time.Time)error {
//
//
//}
//func (c *Context) SetWriteDeadline(t time.Time)error {
//
//
//}

func (c *Context) readAll() (err error) {
	var tmp = make([]byte, cacheSize)
	var n int
	switch conn := c.conn.(type) {
	case net.Conn:
		for {
			n, err = conn.Read(tmp)
			if err != nil {
				return err
			}
			c.read.WriteBytes(tmp[:n])

			if n < cacheSize {
				break
			}
		}
		//case *PacketConn:
		//	for {
		//		n, conn.addr, err = conn.conn.ReadFrom(tmp)
		//		if err != nil {
		//			return err
		//		}
		//		c.read.WriteBytes(tmp[:n])
		//		if n < cacheSize {
		//			break
		//		}
		//	}
	}
	return getMiddle(c.context).use(c, read)
}
