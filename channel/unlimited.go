package channel

import (
	"github.com/Rehtt/Kit/link"
)

type channel struct {
	In    chan<- interface{}
	Out   <-chan interface{}
	dlink *link.DLink
}

func New() (c *channel) {
	in := make(chan interface{})
	out := make(chan interface{})
	c = &channel{
		In:    in,
		Out:   out,
		dlink: link.NewDLink(),
	}
	go func(in, out chan interface{}, c *channel) {
		defer close(out)
		for {
			value, ok := <-in
			if !ok {
				return
			}
			c.dlink.Push(value)
			for c.dlink.Len() != 0 {
				select {
				case value, ok = <-in:
					if !ok {
						continue
					}
					c.dlink.Push(value)
				case out <- c.dlink.Peek():
					c.dlink.Pull()
				}
			}
		}
	}(in, out, c)
	return
}

func (c *channel) Len() int64 {
	return c.dlink.Len()
}
func (c *channel) Cap() int64 {
	return c.dlink.Cap()
}

func (c *channel) Close() {
	close(c.In)
}
