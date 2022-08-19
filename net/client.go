package gonet

import (
	"github.com/Rehtt/Kit/multiplex"
	"net"
)

type Dialer struct {
	dialer       *net.Dialer
	TcpMultiplex bool
	*middle
	*Context
}

func Dial(network, addr, laddr string, tcpMultiplex ...bool) (*Dialer, error) {
	var dial = new(Dialer)
	dial.TcpMultiplex = len(tcpMultiplex) > 0 && tcpMultiplex[0]
	err := dial.Dial(network, addr, laddr)
	return dial, err
}
func (d *Dialer) Dial(network, addr, laddr string) (err error) {

	var l net.Addr
	if laddr != "" {
		switch network {
		case "tcp", "tcp4", "tcp6":
			l, err = net.ResolveTCPAddr(network, laddr)
		case "udp", "udp4", "udp6":
			l, err = net.ResolveUDPAddr(network, laddr)
		}
		if err != nil {
			return err
		}
	}

	if d.TcpMultiplex {
		d.dialer = multiplex.NetDialer()
		d.dialer.LocalAddr = l
	} else {
		d.dialer = new(net.Dialer)
		d.dialer.LocalAddr = l
	}
	d.middle = new(middle)
	conn, err := d.dialer.Dial(network, addr)
	if err != nil {
		return err
	}
	c := newContext(conn)
	c.setValue(RemoteAddr, conn.RemoteAddr())
	c.setValue(LocalAddr, conn.LocalAddr())
	c.setValue(Middle, d.middle)
	d.Context = c
	return
}
