package channel

import (
	"github.com/Rehtt/Kit/link"
)

type Chan struct {
	In    chan<- any
	Out   <-chan any
	dlink *link.DLink
}

func New() (c *Chan) {
	in := make(chan any)
	out := make(chan any)
	c = &Chan{
		In:    in,
		Out:   out,
		dlink: link.NewDLink(),
	}
	go func(in, out chan any, c *Chan) {
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

func (c *Chan) Len() int64 {
	return c.dlink.Len()
}
func (c *Chan) Cap() int64 {
	return c.dlink.Cap()
}

func (c *Chan) Close() {
	close(c.In)
}
