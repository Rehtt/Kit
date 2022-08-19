package gonet

import (
	"github.com/Rehtt/Kit/multiplex"
	"net"
	"syscall"
	"time"
)

type listenerConfig struct {
	network      string
	addr         string
	tcpMultiplex bool
	config       *net.ListenConfig
	*middle
	Handle func(ctx *Context) error
}

func (l *listenerConfig) SetKeepAlive(t time.Duration) {
	l.config.KeepAlive = t
}
func (l *listenerConfig) SetControl(f func(network string, address string, c syscall.RawConn) error) {
	l.config.Control = f
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
		middle:       new(middle),
	}
}
