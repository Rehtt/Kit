package heap

import (
	"fmt"
	"testing"
)

func TestNewHeap(t *testing.T) {
	heap := NewHeap()
	heap.Push(1)
	heap.Push(2)
	heap.Push(3)
	fmt.Println(heap.Pop())
	heap.Push(4)

	fmt.Println("Len", heap.Len())

	fmt.Println(heap.Pop())
	fmt.Println(heap.Pop())
	fmt.Println(heap.Pop())
	fmt.Println(heap.Pop())
	fmt.Println(heap.Pop())

	fmt.Println("CountElem", heap.CountElem())
}
