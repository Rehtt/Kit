package gonet

import (
	"context"
	"github.com/Rehtt/Kit/multiplex"
	"io"
	"log"
	"net"
	"syscall"
	"time"
)

type listenerConfig struct {
	network      string
	addr         string
	tcpMultiplex bool
	config       *net.ListenConfig
	middle
	Handle func(ctx *Context) error
}

type Listener struct {
	listenerConfig
}
type PacketConn struct {
	*Context
}

func Listen(network, addr string, tcpMultiplex ...bool) *Listener {
	return &Listener{
		listenerConfig: config(network, addr, tcpMultiplex...),
	}
}

func ListenPacket(network, addr string, tcpMultiplex ...bool) (*PacketConn, error) {
	l := Listen(network, addr, tcpMultiplex...)
	pc, err := l.config.ListenPacket(context.Background(), network, addr)
	if err != nil {
		return nil, err
	}
	pc.Close()
	return &PacketConn{newContext(pc)}, nil
}
func (p *PacketConn) Close() error {

	if err := p.pconn.Close(); err != nil {
		return err
	}
	delContext(p.Context)
	return nil
}

func config(network, addr string, tcpMultiplex ...bool) listenerConfig {
	var c *net.ListenConfig
	if len(tcpMultiplex) > 0 && tcpMultiplex[0] {
		c = multiplex.NetLister()
	} else {
		c = &net.ListenConfig{}
	}
	return listenerConfig{
		network:      network,
		addr:         addr,
		tcpMultiplex: len(tcpMultiplex) > 0 && tcpMultiplex[0],
		config:       c,
	}
}
func (l *listenerConfig) SetKeepAlive(t time.Duration) {
	l.config.KeepAlive = t
}
func (l *listenerConfig) SetControl(f func(network string, address string, c syscall.RawConn) error) {
	l.config.Control = f
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
	defer delContext(ctx)
	defer conn.Close()
	for {
		if ctx.isDone() {
			return
		}
		n, err := ctx.read.ReadFrom(conn)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			return
		}
		if n == 0 {
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
