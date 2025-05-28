package maps

type ExpirationHeap[T any] []*Node[T]

func (e ExpirationHeap[T]) Len() int {
	return len(e)
}

func (e ExpirationHeap[T]) Less(i, j int) bool {
	return e[i].ExpirtimeUnix < e[j].ExpirtimeUnix
}

func (e ExpirationHeap[T]) Swap(i, j int) {
	if i > e.Len()-1 || j > e.Len()-1 {
		return
	}
	e[i], e[j] = e[j], e[i]
}

func (e *ExpirationHeap[T]) Push(x any) {
	*e = append(*e, x.(*Node[T]))
}

func (e *ExpirationHeap[T]) Pop() any {
	if e.Len() == 0 {
		return nil
	}
	list := *e
	data := list[e.Len()-1]
	list[e.Len()-1] = nil // 底层数组释放引用，更利于GC
	*e = list[:e.Len()-1]
	return data
}
