// util/merge_sort_test.go
package util

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

// 帮助函数：slice ↔ list
func listFromSlice[T any](vs []T) *Node[T] {
	if len(vs) == 0 {
		return nil
	}
	head := &Node[T]{Val: vs[0]}
	curr := head
	for _, v := range vs[1:] {
		curr.Next = &Node[T]{Val: v}
		curr = curr.Next
	}
	return head
}

func sliceFromList[T any](head *Node[T]) []T {
	var out []T
	for p := head; p != nil; p = p.Next {
		out = append(out, p.Val)
	}
	return out
}

// int 比较函数
func lessInt(a, b int) bool { return a < b }

// 测试用结构体，用于检查“稳定性”
type Person struct {
	ID   int // 原序号
	Age  int // 参与比较的键
	Name string
}

func lessByAge(a, b Person) bool { return a.Age < b.Age }

func TestMergeSort(t *testing.T) {
	t.Run("nil list", func(t *testing.T) {
		if got := MergeSort[int](nil, lessInt); got != nil {
			t.Fatalf("want nil, got %#v", got)
		}
	})

	t.Run("single element", func(t *testing.T) {
		exp := []int{42}
		if got := sliceFromList(MergeSort(listFromSlice(exp), lessInt)); !reflect.DeepEqual(got, exp) {
			t.Fatalf("want %v, got %v", exp, got)
		}
	})

	t.Run("already sorted", func(t *testing.T) {
		exp := []int{1, 2, 3, 4, 5}
		if got := sliceFromList(MergeSort(listFromSlice(exp), lessInt)); !reflect.DeepEqual(got, exp) {
			t.Fatalf("want %v, got %v", exp, got)
		}
	})

	t.Run("reverse order", func(t *testing.T) {
		in := []int{9, 7, 5, 3, 1}
		exp := []int{1, 3, 5, 7, 9}
		if got := sliceFromList(MergeSort(listFromSlice(in), lessInt)); !reflect.DeepEqual(got, exp) {
			t.Fatalf("want %v, got %v", exp, got)
		}
	})

	t.Run("random 1e4", func(t *testing.T) {
		const n = 1e4
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		in := make([]int, n)
		for i := range in {
			in[i] = r.Intn(100000)
		}
		head := MergeSort(listFromSlice(in), lessInt)
		got := sliceFromList(head)
		// 验证非降序
		for i := 1; i < len(got); i++ {
			if got[i-1] > got[i] {
				t.Fatalf("not sorted at %d: %d > %d", i, got[i-1], got[i])
			}
		}
	})

	t.Run("stability", func(t *testing.T) {
		in := []Person{
			{ID: 1, Age: 30, Name: "A"},
			{ID: 2, Age: 20, Name: "B"},
			{ID: 3, Age: 30, Name: "C"}, // 与 ID=1 同键
			{ID: 4, Age: 20, Name: "D"}, // 与 ID=2 同键
		}
		expOrderByID := []int{2, 4, 1, 3} // Age 升序且同 Age 保持原序号顺序

		gotList := MergeSort(listFromSlice(in), lessByAge)
		gotIDs := make([]int, 0, len(in))
		for p := gotList; p != nil; p = p.Next {
			gotIDs = append(gotIDs, p.Val.ID)
		}
		if !reflect.DeepEqual(gotIDs, expOrderByID) {
			t.Fatalf("want ID order %v, got %v", expOrderByID, gotIDs)
		}
	})
}
