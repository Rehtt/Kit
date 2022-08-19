package gonet

import (
	"context"
	"log"
	"net"
)

type Listener struct {
	listenerConfig
}
type PacketConn struct {
	ctx  map[net.Addr]*Context
	conn net.PacketConn
	l    *Listener
}

func Listen(network, addr string, tcpMultiplex ...bool) *Listener {
	return &Listener{
		listenerConfig: config(network, addr, tcpMultiplex...),
	}
}

func (l *Listener) Run() error {
	ll, err := l.config.Listen(context.Background(), l.network, l.addr)
	if err != nil {
		return err
	}
	for {
		conn, err := ll.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go l.handle(conn)
	}
}

func (l *Listener) handle(conn net.Conn) {
	ctx := newContext(conn)
	ctx.setValue(Middle, l.middle)
	ctx.setValue(RemoteAddr, conn.RemoteAddr())
	ctx.setValue(LocalAddr, conn.LocalAddr())
	defer delContext(ctx)
	defer conn.Close()
	for {
		if ctx.isDone() {
			return
		}
		err := ctx.readAll()
		if err != nil {
			log.Println(err)
			return
		}
		if err = l.middle.use(ctx, read); err != nil {
			log.Println(err)
			return
		}
		if err = l.Handle(ctx); err != nil {
			log.Println(err)
			return
		}
	}
}

func ListenPacket(network, addr string, tcpMultiplex ...bool) (*PacketConn, error) {
	l := Listen(network, addr, tcpMultiplex...)
	pc, err := l.config.ListenPacket(context.Background(), network, addr)
	if err != nil {
		return nil, err
	}
	return &PacketConn{
		ctx:  make(map[net.Addr]*Context),
		conn: pc,
		l:    l,
	}, nil
}

func (p *PacketConn) Close(addr net.Addr) error {
	if ctx, ok := p.ctx[addr]; ok {
		delContext(ctx)
	}
	return nil
}
func (p *PacketConn) addOrLoadContext(addr net.Addr) *Context {
	ctx, ok := p.ctx[addr]
	if !ok {
		ctx = newContext(p)
		ctx.setValue(Middle, p.l.middle)
		ctx.setValue(RemoteAddr, addr)
		ctx.setValue(LocalAddr, p.conn.LocalAddr())
		p.ctx[addr] = ctx
	}
	return ctx
}

func (p *PacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	ctx := p.addOrLoadContext(addr)
	return ctx.Write(b)
}

func (p *PacketConn) ReadToBytes() (data []byte, addr net.Addr, err error) {
	ctx, err := p.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	return ctx.read.ToBytes(), ctx.RemoteAddr(), nil
}

func (p *PacketConn) ReadAll() (*Context, error) {
	var (
		tmp  = make([]byte, 512)
		addr net.Addr
		err  error
		n    int
		ctx  *Context
	)

	for {
		n, addr, err = p.conn.ReadFrom(tmp)
		if err != nil {
			return nil, err
		}
		if ctx == nil {
			ctx = p.addOrLoadContext(addr)
		}
		ctx.read.WriteBytes(tmp[:n])
		if n < 512 {
			break
		}
	}
	err = p.l.middle.use(ctx, read)
	return ctx, err
}

//func (p *PacketConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
//	var data *buf.Buf
//	_, data, addr, err = p.ReadAll()
//	if err != nil {
//		return
//	}
//	n, err = data.Read(b)
//	return
//}
