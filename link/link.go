package link

import (
	"fmt"
	"sync"
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
	top     *Node
	bottom  *Node
	cap     int
	len     int
	AutoLen bool
}

// 双向循环链表
func NewDLink() *DLink {
	return &DLink{
		AutoLen: true,
	}
}
func (l *DLink) Size(size int) error {
	if size > l.cap {
		l.AddNode(size - l.cap)
	} else if size < l.cap {
		err := l.DelNode(l.cap - size)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddNode 扩容
func (l *DLink) AddNode(n int) {
	if l.cap == 0 {
		l.top = newNode()
		l.top.next = l.top
		l.top.pre = l.top
		l.cap += 1
		n -= 1
	}
	var index = l.top.pre
	for i := 0; i < n; i++ {
		index.next = newNode()
		index.next.pre = index
		index = index.next
		l.cap += 1
	}
	index.next = l.top
	l.top.pre = index
}

// DelNode 缩容
func (l *DLink) DelNode(n int) error {
	if n > l.cap {
		return fmt.Errorf("too big")
	}
	var index = l.top.pre
	var hasBottom bool
	for i := 0; i < n; i++ {
		if l.bottom == index {
			hasBottom = true
		}
		index = index.pre
		delNode(index.next)
		l.cap -= 1
	}
	if hasBottom {
		l.bottom = index
	}
	index.next = l.top
	l.top.pre = index
	return nil
}

func (l *DLink) Push(value interface{}) {
	if l.len == l.cap {
		if l.AutoLen {
			l.AddNode(5) // 自动扩充
		} else {
			l.top = l.top.next
			l.len -= 1
		}
	}
	if l.bottom == nil {
		l.bottom = l.top
	} else {
		l.bottom = l.bottom.next
	}
	l.bottom.Value = value
	l.len += 1
}
func (l *DLink) Pull() interface{} {
	l.top = l.top.next
	if l.top.pre == l.bottom {
		l.bottom = l.top
	}
	defer func() {
		l.top.pre.Value = nil
	}()
	if l.len != 0 {
		l.len -= 1
	}
	return l.top.pre.Value
}

func (l *DLink) Range() (out []interface{}) {
	index := l.top
	for i := 0; i < l.len; i++ {
		out = append(out, index.Value)
		index = index.next
	}

	return out
}

func (l *DLink) Len() int {
	return l.len
}
func (l *DLink) Cap() int {
	return l.cap
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
