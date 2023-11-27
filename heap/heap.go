package heap

import "sync/atomic"

type node struct {
	next *node
	data any
}

func (n *node) init() *node {
	n.next = nil
	n.data = nil
	return n
}

func (h *Heap) newNode() *node {
	atomic.AddInt32(&h.len, 1)
	if h.fuS == nil {
		return &node{}
	}
	atomic.AddInt32(&h.fuLen, -1)
	this := h.fuS
	h.fuS = this.next
	return this.init()
}

func (h *Heap) closeD(a *node) {
	atomic.AddInt32(&h.fuLen, 1)
	atomic.AddInt32(&h.len, -1)
	a.init()
	if h.fuS == nil {
		h.fuS = a
	} else if h.fuE != nil {
		h.fuE.next = a
	}
	h.fuE = a
}

type Heap struct {
	thisPrt *node
	endPrt  *node
	len     int32

	fuS   *node
	fuE   *node
	fuLen int32
}

func (h *Heap) Pop() any {
	if h.thisPrt == nil {
		return nil
	}
	this := h.thisPrt
	h.thisPrt = h.thisPrt.next
	defer h.closeD(this)
	return this.data
}

func (h *Heap) Push(data any) {
	endPrt := h.newNode()
	endPrt.data = data

	if h.thisPrt == nil {
		h.thisPrt = endPrt
	} else if h.endPrt != nil {
		h.endPrt.next = endPrt
	}
	h.endPrt = endPrt
}

func (h *Heap) Len() int32 {
	return h.len
}

func (h *Heap) CountElem() int32 {
	return h.len + h.fuLen
}

func NewHeap() *Heap {
	return &Heap{}
}
