package util

type Node[T any] struct {
	Val  T
	Next *Node[T]
}

// CompareFn 定义比较函数，返回 true 表示 a < b
type CompareFn[T any] func(a, b T) bool

// MergeSort 对链表做稳定排序，返回排好序的新表头
func MergeSort[T any](head *Node[T], less CompareFn[T]) *Node[T] {
	if head == nil || head.Next == nil {
		return head
	}

	length := 0
	for p := head; p != nil; p = p.Next {
		length++
	}

	dummy := &Node[T]{Next: head}

	// 2. bottom-up 归并
	for step := 1; step < length; step <<= 1 {
		prev, curr := dummy, dummy.Next
		for curr != nil {
			// 2.1 取 a 段
			a := curr
			aLen := step
			for aLen > 0 && curr != nil {
				curr = curr.Next
				aLen--
			}
			if aLen > 0 { // a 不足 step ⇒ 整段挂回即可
				prev.Next = a
				break
			}

			// 2.2 取 b 段
			b := curr
			bLen := step
			for bLen > 0 && curr != nil {
				curr = curr.Next
				bLen--
			}

			// 2.3 归并 a、b
			mergedHead, mergedTail := merge(a, step, b, step-bLen, less)

			// 2.4 接回主链
			prev.Next = mergedHead
			mergedTail.Next = curr
			prev = mergedTail
		}
	}
	return dummy.Next
}

// merge 把两条已知长度的有序链表归并，返回头尾指针
func merge[T any](a *Node[T], aLen int, b *Node[T], bLen int,
	less CompareFn[T],
) (head, tail *Node[T]) {
	dummy := &Node[T]{}
	p := dummy
	for aLen > 0 && bLen > 0 {
		// 稳定：若 a==b 先取 a
		if !less(b.Val, a.Val) { // a<=b
			p.Next, a = a, a.Next
			aLen--
		} else { // b<a
			p.Next, b = b, b.Next
			bLen--
		}
		p = p.Next
	}
	for aLen > 0 {
		p.Next, a = a, a.Next
		aLen--
		p = p.Next
	}
	for bLen > 0 {
		p.Next, b = b, b.Next
		bLen--
		p = p.Next
	}
	p.Next = nil // 断尾，防残链
	return dummy.Next, p
}
