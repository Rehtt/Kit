package link

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Node[T any] struct {
	pre   *Node[T]
	next  *Node[T]
	Value T
}
type DLink[T any] struct {
	top    *Node[T]
	bottom *Node[T]
	cap    atomic.Int64
	len    atomic.Int64
	// 自动扩容
	AutoLen bool
	// 返回被循环链表覆盖的值
	OnCover  func(value T)
	nodePool sync.Pool
}

// 双向循环链表
// *非并发安全*
func NewDLink[T any]() *DLink[T] {
	return &DLink[T]{
		AutoLen: true,
		nodePool: sync.Pool{
			New: func() any {
				return new(Node[T])
			},
		},
	}
}

func (l *DLink[T]) Size(size int64) error {
	if size > l.Cap() {
		l.AddNode(size - l.Cap())
	} else if size < l.Cap() {
		err := l.DelNode(l.Cap() - size)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddNode 扩容
func (l *DLink[T]) AddNode(n int64) {
	if l.Cap() == 0 {
		l.top = l.newNode()
		l.top.next = l.top
		l.top.pre = l.top
		l.cap.Add(1)
		n -= 1
	}
	index := l.top.pre
	for i := int64(0); i < n; i++ {
		index.next = l.newNode()
		index.next.pre = index
		index = index.next
		l.cap.Add(1)
	}
	index.next = l.top
	l.top.pre = index
}

// DelNode 缩容
func (l *DLink[T]) DelNode(n int64) error {
	if n > l.Cap() {
		return fmt.Errorf("too big")
	}
	index := l.top.pre
	var hasBottom bool
	for i := int64(0); i < n; i++ {
		if l.bottom == index {
			hasBottom = true
		}
		index = index.pre
		l.delNode(index.next)
		l.cap.Add(-1)
	}
	if hasBottom {
		l.bottom = index
	}
	index.next = l.top
	l.top.pre = index
	return nil
}

func (l *DLink[T]) Peek() T {
	return l.top.Value
}

func (l *DLink[T]) Push(value T) {
	if l.Len() == l.Cap() {
		if l.AutoLen {
			l.AddNode(5) // 自动扩充
		} else {
			if l.OnCover != nil {
				l.OnCover(l.top.Value)
			}
			l.top = l.top.next
			l.len.Add(-1)
		}
	}
	if l.bottom == nil {
		l.bottom = l.top
	} else {
		l.bottom = l.bottom.next
	}
	l.bottom.Value = value
	l.len.Add(1)
}

func (l *DLink[T]) Pull() (v T) {
	if l.Len() == 0 {
		return
	}
	c := l.top
	v = c.Value
	l.top = c.next
	l.len.Add(-1)
	return
}

func (l *DLink[T]) Range() (out []T) {
	index := l.top
	out = make([]T, 0, l.Len())
	for i := int64(0); i < l.Len(); i++ {
		out = append(out, index.Value)
		index = index.next
	}

	return out
}

func (l *DLink[T]) Len() int64 {
	return l.len.Load()
}

func (l *DLink[T]) Cap() int64 {
	return l.cap.Load()
}

func (l *DLink[T]) newNode() (node *Node[T]) {
	node = l.nodePool.Get().(*Node[T])
	var empty T
	node.pre = nil
	node.next = nil
	node.Value = empty
	return node
}

func (l *DLink[T]) delNode(node *Node[T]) {
	l.nodePool.Put(node)
}
