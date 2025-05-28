// link_test.go
package link

import (
	"reflect"
	"testing"
)

// ---------- 单元测试 ----------

func TestDLink_BasicOps(t *testing.T) {
	d := NewDLink[int]()
	if d.Len() != 0 || d.Cap() != 0 {
		t.Fatalf("new list should be empty")
	}

	// 指定初始容量
	if err := d.Size(3); err != nil {
		t.Fatalf("Size err=%v", err)
	}
	if d.Cap() != 3 {
		t.Fatalf("cap want 3, got %d", d.Cap())
	}

	// Push / Peek / Pull
	for i := 1; i <= 3; i++ {
		d.Push(i)
		if top := d.Peek(); top != 1 {
			t.Fatalf("peek want 1, got %v", top)
		}
	}
	if d.Len() != 3 {
		t.Fatalf("len want 3, got %d", d.Len())
	}
	if v := d.Range(); !reflect.DeepEqual(v, []int{1, 2, 3}) {
		t.Fatalf("range want [1 2 3], got %v", v)
	}

	// 自动扩容
	d.Push(4)
	if d.Cap() != 8 { // 自动 +5
		t.Fatalf("auto grow cap want 8, got %d", d.Cap())
	}

	// Pull 顺序
	for i := 1; i <= 4; i++ {
		if v := d.Pull(); v != i {
			t.Fatalf("pull want %d, got %v", i, v)
		}
	}
	if d.Len() != 0 {
		t.Fatalf("len should be 0 after pull all")
	}
}

func TestDLink_CoverMode(t *testing.T) {
	d := NewDLink[string]()
	_ = d.Size(2)
	d.AutoLen = false

	var covered any
	d.OnCover = func(v string) { covered = v }

	d.Push("a")
	d.Push("b")
	d.Push("c") // 覆盖 "a"

	if covered != "a" {
		t.Fatalf("cover callback want a, got %v", covered)
	}
	if v := d.Range(); !reflect.DeepEqual(v, []string{"b", "c"}) {
		t.Fatalf("after cover want [b c], got %v", v)
	}
}

func TestDLink_DelNodeShrink(t *testing.T) {
	d := NewDLink[int]()
	_ = d.Size(5)
	d.Push(1)
	d.Push(2)

	if err := d.DelNode(3); err != nil {
		t.Fatalf("DelNode err=%v", err)
	}
	if d.Cap() != 2 {
		t.Fatalf("cap want 2, got %d", d.Cap())
	}
	if v := d.Range(); !reflect.DeepEqual(v, []int{1, 2}) {
		t.Fatalf("data lost after shrink, got %v", v)
	}
}

func TestDLink_SizeEdgeCases(t *testing.T) {
	d := NewDLink[int]()

	cases := []struct {
		size int64
		want int64
	}{
		{0, 0},
		{1, 1},
		{5, 5},
		{2, 2}, // 缩容
	}

	for _, c := range cases {
		if err := d.Size(c.size); err != nil {
			t.Fatalf("Size(%d) err=%v", c.size, err)
		}
		if d.Cap() != c.want {
			t.Fatalf("cap want %d, got %d", c.want, d.Cap())
		}
	}
}

// ---------- 基准测试 ----------

// 基准参数
const benchN = 1_000_000

func BenchmarkPush(b *testing.B) {
	d := NewDLink[int]()
	_ = d.Size(benchN)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Push(i)
		// 重置 len 避免 AutoLen 干扰
		if d.len.Load() == benchN {
			d.len.Store(0)
			d.top = d.bottom.next
		}
	}
}

func BenchmarkPushPull(b *testing.B) {
	d := NewDLink[int]()
	_ = d.Size(benchN)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Push(i)
		_ = d.Pull()
	}
}

func BenchmarkRange(b *testing.B) {
	d := NewDLink[int]()
	_ = d.Size(benchN)
	for i := 0; i < benchN; i++ {
		d.Push(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.Range()
	}
}
