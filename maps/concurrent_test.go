package maps

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

/* ---------- 单元测试 ---------- */

// 1. 基本 Set / Get
func TestConcurrentMapBasic(t *testing.T) {
	m := NewConcurrentMap[int]()

	m.Set("foo", 42)
	v, ok := m.Get("foo")
	if !ok || v != 42 {
		t.Fatalf("want (42,true), got (%v,%v)", v, ok)
	}

	if _, ok := m.Get("none"); ok {
		t.Fatalf("expect false on absent key")
	}
}

// 2. TTL 到期后应视为不存在
func TestConcurrentMapTTL(t *testing.T) {
	m := NewConcurrentMap[int](EnableExpired(400 * time.Millisecond))

	m.Set("bar", 1)
	time.Sleep(450 * time.Millisecond) // 让它过期

	// RWMutexMap.getValidValue 会直接返回 ok=false
	if _, ok := m.Get("bar"); ok {
		t.Fatal("expired key should return ok=false")
	}

	// 手动触发 Clear 再确认已删除（可选，验证 Clear/Pool）
	for _, shard := range m.maps {
		shard.Clear()
	}
	if _, ok := m.Get("bar"); ok {
		t.Fatal("Clear should remove expired key completely")
	}
}

// 3. Delete
func TestConcurrentMapDelete(t *testing.T) {
	m := NewConcurrentMap[string]()

	m.Set("k", "v")
	m.Delete("k")
	if _, ok := m.Get("k"); ok {
		t.Fatal("delete failed, key still exists")
	}
}

// 4. SetByFunc
func TestConcurrentMapSetByFunc(t *testing.T) {
	m := NewConcurrentMap[int]()

	m.Set("cnt", 1)
	out := m.SetByFunc("cnt", func(old int) int { return old + 2 })
	if out != 3 {
		t.Fatalf("SetByFunc returned %d, want 3", out)
	}
	v, _ := m.Get("cnt")
	if v != 3 {
		t.Fatalf("stored value %d, want 3", v)
	}
}

// 5. 并发安全性（配合 go test -race）
func TestConcurrentMapRace(t *testing.T) {
	m := NewConcurrentMap[int]()
	var wg sync.WaitGroup

	workers := 64
	ops := 200

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				key := fmt.Sprintf("w%d-%d", id, j)
				m.Set(key, j)
				m.Get(key)
				m.SetByFunc(key, func(old int) int { return old + 1 })
				m.Delete(key)
			}
		}(i)
	}
	wg.Wait()
}

/* ---------- 基准测试 ---------- */

// 高频写入
func BenchmarkConcurrentMapSet(b *testing.B) {
	m := NewConcurrentMap[int]()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		m.Set(key, i)
	}
}

// 高频写入 + Clear（TTL 500 ms）
func BenchmarkConcurrentMapSetClear(b *testing.B) {
	m := NewConcurrentMap[int](EnableExpired(500 * time.Millisecond))
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		m.Set(key, i)
		if i%1024 == 0 { // 模拟后台清理
			for _, shard := range m.maps {
				shard.Clear()
			}
		}
	}
}
