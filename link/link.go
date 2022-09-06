package link

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var nodePool = sync.Pool{
	New: func() interface{} {
		return new(Node)
	},
}

type Node struct {
	pre   *Node
	next  *Node
	Value interface{}
}
type DLink struct {
	top    *Node
	bottom *Node
	cap    *int64
	len    *int64
	// 自动扩容
	AutoLen bool
	// 返回被循环链表覆盖的值
	OnCover func(value interface{})
}

// 双向循环链表
func NewDLink() *DLink {
	return &DLink{
		AutoLen: true,
		len:     new(int64),
		cap:     new(int64),
	}
}
func (l *DLink) Size(size int64) error {
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
func (l *DLink) AddNode(n int64) {
	if l.Cap() == 0 {
		l.top = newNode()
		l.top.next = l.top
		l.top.pre = l.top
		atomic.AddInt64(l.cap, 1)
		n -= 1
	}
	var index = l.top.pre
	for i := int64(0); i < n; i++ {
		index.next = newNode()
		index.next.pre = index
		index = index.next
		atomic.AddInt64(l.cap, 1)
	}
	index.next = l.top
	l.top.pre = index
}

// DelNode 缩容
func (l *DLink) DelNode(n int64) error {
	if n > l.Cap() {
		return fmt.Errorf("too big")
	}
	var index = l.top.pre
	var hasBottom bool
	for i := int64(0); i < n; i++ {
		if l.bottom == index {
			hasBottom = true
		}
		index = index.pre
		delNode(index.next)
		atomic.AddInt64(l.cap, -1)
	}
	if hasBottom {
		l.bottom = index
	}
	index.next = l.top
	l.top.pre = index
	return nil
}
func (l *DLink) Peek() interface{} {
	return l.top.Value
}

func (l *DLink) Push(value interface{}) {
	if l.Len() == l.Cap() {
		if l.AutoLen {
			l.AddNode(5) // 自动扩充
		} else {
			if l.OnCover != nil {
				l.OnCover(l.top.Value)
			}
			l.top = l.top.next
			atomic.AddInt64(l.len, -1)

		}
	}
	if l.bottom == nil {
		l.bottom = l.top
	} else {
		l.bottom = l.bottom.next
	}
	l.bottom.Value = value
	atomic.AddInt64(l.len, 1)
}
func (l *DLink) Pull() (v interface{}) {
	if l.Len() == 0 {
		return nil
	}
	v = l.top.Value
	l.top.Value = nil
	l.top = l.top.next
	atomic.AddInt64(l.len, -1)
	return
}

func (l *DLink) Range() (out []interface{}) {
	index := l.top
	for i := int64(0); i < l.Len(); i++ {
		out = append(out, index.Value)
		index = index.next
	}

	return out
}

func (l *DLink) Len() int64 {
	return atomic.LoadInt64(l.len)
}
func (l *DLink) Cap() int64 {
	return atomic.LoadInt64(l.cap)
}

func newNode() (node *Node) {
	node = nodePool.Get().(*Node)
	node.pre = nil
	node.next = nil
	node.Value = nil
	return node
}
func delNode(l *Node) {
	nodePool.Put(l)
}
