package maps

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

/* ----------------- 单元测试 ----------------- */

// TestSetGet ‒ 正常写入后可读取
func TestSetGet(t *testing.T) {
	m := NewRWMutexMap[int](0)

	m.Set("a", 10)
	v, ok := m.Get("a")
	if !ok || v != 10 {
		t.Fatalf("want (10,true), got (%v,%v)", v, ok)
	}
}

// TestTTLExpire ‒ TTL 到期后 Clear 能清掉键
func TestTTLExpire(t *testing.T) {
	m := NewRWMutexMap[int](0)

	m.Set("a", 1, 1500*time.Millisecond) // 1.5 s
	time.Sleep(1600 * time.Millisecond)  // 等到过期
	m.Clear()

	if _, ok := m.Get("a"); ok {
		t.Fatal("key a should have expired and been deleted")
	}
}

// TestDelete ‒ Delete 后无法再读取
func TestDelete(t *testing.T) {
	m := NewRWMutexMap[int](0)

	m.Set("a", 1)
	m.Delete("a")
	if _, ok := m.Get("a"); ok {
		t.Fatal("delete failed: key a still exists")
	}
}

// TestSetByFunc ‒ 根据旧值计算新值
func TestSetByFunc(t *testing.T) {
	m := NewRWMutexMap[int](0)

	m.Set("cnt", 1)
	newV := m.SetByFunc("cnt", func(old int) int { return old + 2 })
	if newV != 3 {
		t.Fatalf("want new value 3, got %d", newV)
	}
	v, _ := m.Get("cnt")
	if v != 3 {
		t.Fatalf("stored value want 3, got %d", v)
	}
}

// TestConcurrentSafety ‒ 并发读写不会竞态 (go test -race)
func TestConcurrentSafety(t *testing.T) {
	m := NewRWMutexMap[int](3 * time.Second)
	var wg sync.WaitGroup

	// 并发写
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			m.SetByFunc("num", func(old int) int { return old + 1 })
		}(i)
	}

	// 并发读
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = m.Get("num")
		}()
	}
	wg.Wait()
}

/* ----------------- 基准测试 ----------------- */

// BenchmarkSetClear ‒ 高频写入并定期 Clear
func BenchmarkSetClear(b *testing.B) {
	m := NewRWMutexMap[int](time.Second) // 统一 TTL 1s
	for i := 0; i < b.N; i++ {
		key := "k" + strconv.Itoa(i)
		m.Set(key, i)
		if i%1024 == 0 { // 每 1024 次写尝试清理一次
			m.Clear()
		}
	}
}
