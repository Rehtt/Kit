package channel

import (
	"sync/atomic"

	"github.com/Rehtt/Kit/link"
)

type Chan[T any] struct {
	In      chan<- T
	Out     <-chan T
	dlink   *link.DLink[T]
	isClose atomic.Bool
}

func New[T any]() (c *Chan[T]) {
	in := make(chan T)
	out := make(chan T)
	c = &Chan[T]{
		In:    in,
		Out:   out,
		dlink: link.NewDLink[T](),
	}
	go func(in, out chan T, c *Chan[T]) {
		defer close(out)
		for {
			value, ok := <-in
			if !ok {
				// 推完全部后退出
				for c.dlink.Len() > 0 {
					out <- c.dlink.Pull()
				}
				return
			}
			c.dlink.Push(value)
			for c.dlink.Len() != 0 {
				select {
				case value, ok = <-in:
					if !ok {
						break
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

func (c *Chan[T]) Len() int64 {
	return c.dlink.Len()
}

func (c *Chan[T]) Cap() int64 {
	return c.dlink.Cap()
}

func (c *Chan[T]) Close() {
	if c.isClose.CompareAndSwap(false, true) {
		close(c.In)
	}
}
