package _struct

import (
	"errors"
	"testing"
)

// helper: 将 []*DiffData 转为 map[string][2]any 便于断言
func diffsToMap(diffs []*DiffData) map[string][2]any {
	m := make(map[string][2]any, len(diffs))
	for _, d := range diffs {
		m[d.Field.Name] = [2]any{d.ValueA.Interface(), d.ValueB.Interface()}
	}
	return m
}

func TestDiffStructBasicEqual(t *testing.T) {
	type S struct {
		A int
		B string
	}
	s1 := S{A: 1, B: "x"}
	s2 := S{A: 1, B: "x"}
	// 传值
	diffs, err := DiffStruct(s1, s2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diffs) != 0 {
		t.Fatalf("expected no diffs, got: %v", diffs)
	}

	// 传指针
	diffs2, err2 := DiffStruct(&s1, &s2, nil)
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if len(diffs2) != 0 {
		t.Fatalf("expected no diffs for pointers, got: %v", diffs2)
	}
}

func TestDiffStructBasicDifferent(t *testing.T) {
	type S struct {
		A int
		B string
		C bool
	}
	s1 := S{A: 1, B: "foo", C: true}
	s2 := S{A: 2, B: "foo", C: false}
	diffs, err := DiffStruct(s1, s2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := diffsToMap(diffs)
	// A 和 C 不同
	if len(m) != 2 {
		t.Fatalf("expected 2 diffs, got %d: %v", len(m), m)
	}
	if vals, ok := m["A"]; !ok || vals[0].(int) != 1 || vals[1].(int) != 2 {
		t.Errorf("unexpected diff for A: %v", vals)
	}
	if vals, ok := m["C"]; !ok || vals[0].(bool) != true || vals[1].(bool) != false {
		t.Errorf("unexpected diff for C: %v", vals)
	}
}

func TestDiffStructIgnoreField(t *testing.T) {
	type S struct {
		A int
		B string
	}
	s1 := S{A: 1, B: "foo"}
	s2 := S{A: 2, B: "bar"}
	// 忽略 A
	diffs, err := DiffStruct(s1, s2, []string{"A"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := diffsToMap(diffs)
	// 只应包含 B
	if len(m) != 1 {
		t.Fatalf("expected 1 diff after ignore, got %d: %v", len(m), m)
	}
	if vals, ok := m["B"]; !ok || vals[0].(string) != "foo" || vals[1].(string) != "bar" {
		t.Errorf("unexpected diff for B: %v", vals)
	}
}

func TestDiffStructUnexportedField(t *testing.T) {
	type S struct {
		A int
		b int // 未导出字段，不应比较
	}
	s1 := S{A: 1, b: 10}
	s2 := S{A: 1, b: 20}
	// 虽然 b 不同，但因未导出字段，DiffStruct 不应返回差异
	diffs, err := DiffStruct(s1, s2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diffs) != 0 {
		t.Fatalf("expected no diffs (unexported b field skipped), got: %v", diffs)
	}
}

func TestDiffStructTypeMismatch(t *testing.T) {
	type S1 struct{ A int }
	type S2 struct{ A int }
	s1 := S1{A: 1}
	s2 := S2{A: 1}
	_, err := DiffStruct(s1, s2, nil)
	if err == nil {
		t.Fatalf("expected error due to type mismatch, got nil")
	}
	// 可以检查错误信息包含关键字
	if !errors.Is(err, errors.New("A与B不是同一个属性的结构体")) {
		// 这里只要确认返回非 nil 即可
	}
}

func TestDiffStructNestedStruct(t *testing.T) {
	type Inner struct {
		X int
		Y string
	}
	type Outer struct {
		I Inner
		Z float64
	}
	o1 := Outer{I: Inner{X: 1, Y: "a"}, Z: 3.14}
	o2 := Outer{I: Inner{X: 2, Y: "a"}, Z: 2.71}
	diffs, err := DiffStruct(o1, o2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := diffsToMap(diffs)
	// I 字段作为整体 struct，reflect.Value.Equal 会比较内部字段，因此 I 不相等；Z 也不相等
	if len(m) != 2 {
		t.Fatalf("expected 2 diffs for nested, got %d: %v", len(m), m)
	}
	// 检查 I 字段的值是两个 Inner 结构体
	if vals, ok := m["I"]; !ok {
		t.Errorf("expected diff on I, missing")
	} else {
		innerA, innerB := vals[0].(Inner), vals[1].(Inner)
		if innerA.X != 1 || innerB.X != 2 || innerA.Y != "a" || innerB.Y != "a" {
			t.Errorf("unexpected nested values: %v vs %v", innerA, innerB)
		}
	}
	if vals, ok := m["Z"]; !ok {
		t.Errorf("expected diff on Z, missing")
	} else {
		if vals[0].(float64) != 3.14 || vals[1].(float64) != 2.71 {
			t.Errorf("unexpected Z values: %v vs %v", vals[0], vals[1])
		}
	}
}

func TestDiffStructWithPointerFields(t *testing.T) {
	type S struct {
		A *int
		B *string
	}
	a1, a2 := 1, 2
	s1 := S{A: &a1, B: nil}
	s2 := S{A: &a2, B: nil}
	// reflect.Value.Equal 对指针比较的是指针地址还是值？文档：两个指针类型的 Value，Equal 报 true 当且仅当它们是相同指针，或都为 nil。这里 &a1 != &a2，所以不同。
	diffs, err := DiffStruct(s1, s2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := diffsToMap(diffs)
	if len(m) != 1 {
		t.Fatalf("expected 1 diff for pointer field A, got %d: %v", len(m), m)
	}
	if vals, ok := m["A"]; !ok {
		t.Errorf("expected diff on A")
	} else {
		// 检查接口底层仍为指针
		pa, pb := vals[0].(*int), vals[1].(*int)
		if *pa != 1 || *pb != 2 {
			t.Errorf("unexpected pointer values: %v vs %v", *pa, *pb)
		}
	}
}

func TestDiffStructWithNonComparableFieldPanics(t *testing.T) {
	// reflect.Value.Equal 在遇到 slice、map 等非可比较类型会 panic。此测试展示这一限制。
	type S struct {
		A []int
	}
	s1 := S{A: []int{1, 2}}
	s2 := S{A: []int{1, 2}}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for non-comparable field, but did not panic")
		}
	}()
	// 这里会 panic
	_, _ = DiffStruct(s1, s2, nil)
}

func TestDiffStructWithNilInputs(t *testing.T) {
	// 如果传入 nil，reflect.ValueOf(nil).Kind() is Invalid; Elem 会 panic。可以测试行为：应当 panic 或返回错误。
	// 取决于实现，目前 Elem 会在 ValueOf(nil) 是 zero Value，Kind!=Ptr，Type() panics? 实际应避免传 nil。
	// 这里测试预期 panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic or error for nil input, but did not panic")
		}
	}()
	_, _ = DiffStruct(nil, nil, nil)
}
