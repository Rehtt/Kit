package maps

import (
	"container/heap"
	"math/rand"
	"testing"
	"time"
)

/* ---------- 辅助 ---------- */

// n 生成一个带过期时间的节点，方便测试
func n[T any](v T, ts int64) *Node[T] {
	return &Node[T]{Value: v, ExpirtimeUnix: ts}
}

/* ---------- 单元测试 ---------- */

// TestHeapOrder 验证 Pop 顺序
func TestHeapOrder(t *testing.T) {
	var h ExpirationHeap[int]
	heap.Init(&h)

	in := []int64{9, 3, 7, 1, 5}
	for _, ts := range in {
		heap.Push(&h, n(0, ts))
	}
	if h.Len() != len(in) {
		t.Fatalf("len want %d, got %d", len(in), h.Len())
	}

	want := []int64{1, 3, 5, 7, 9}
	for i, w := range want {
		x := heap.Pop(&h).(*Node[int])
		if x.ExpirtimeUnix != w {
			t.Errorf("pop %d: want %d, got %d", i, w, x.ExpirtimeUnix)
		}
	}
	if h.Len() != 0 {
		t.Fatalf("after pops len want 0, got %d", h.Len())
	}
}

// TestPopEmpty 空堆 Pop 应返回 nil 且不 panic
func TestPopEmpty(t *testing.T) {
	var h ExpirationHeap[string]
	heap.Init(&h)

	if v := heap.Pop(&h); v != nil {
		t.Fatalf("empty pop want nil, got %#v", v)
	}
}

// TestLenUpdate Push/Pop 后 Len 是否准确
func TestLenUpdate(t *testing.T) {
	var h ExpirationHeap[int]
	heap.Init(&h)

	for i := 0; i < 50; i++ {
		heap.Push(&h, n(i, int64(i)))
	}
	if h.Len() != 50 {
		t.Fatalf("after push want 50, got %d", h.Len())
	}
	for i := 0; i < 20; i++ {
		heap.Pop(&h)
	}
	if h.Len() != 30 {
		t.Fatalf("after pop want 30, got %d", h.Len())
	}
}

/* ---------- 基准测试 ---------- */

// BenchmarkPushPop 交替 Push/Pop
func BenchmarkPushPop(b *testing.B) {
	var h ExpirationHeap[int]
	heap.Init(&h)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < b.N; i++ {
		heap.Push(&h, n(0, r.Int63()))
		heap.Pop(&h)
	}
}

// BenchmarkBulk 批量 Push 然后批量 Pop
func BenchmarkBulk(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < b.N; i++ {
		var h ExpirationHeap[int]
		heap.Init(&h)

		for j := 0; j < 1024; j++ {
			heap.Push(&h, n(0, r.Int63()))
		}
		for h.Len() > 0 {
			heap.Pop(&h)
		}
	}
}
