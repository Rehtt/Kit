// channel_test.go
package channel_test

import (
	"sync"
	"testing"
	"time"

	"github.com/Rehtt/Kit/channel"
)

// ---------- 功能测试 ----------

// TestOrder 保证写入顺序 == 读出顺序（FIFO）
func TestOrder(t *testing.T) {
	const n = 1000
	c := channel.New[int]()
	defer c.Close()

	// 写
	go func() {
		for i := 0; i < n; i++ {
			c.In <- i
		}
		c.Close()
	}()

	// 读并校验
	for i := 0; i < n; i++ {
		v, ok := <-c.Out
		if !ok {
			t.Fatalf("out unexpectedly closed at %d", i)
		}
		if v != i {
			t.Fatalf("expect %d, got %v", i, v)
		}
	}

	// 读完后，Out 应该被关闭
	if _, ok := <-c.Out; ok {
		t.Fatalf("out should be closed after all data drained")
	}
}

// TestLenCap 验证 Len/Cap 辅助方法
func TestLenCap(t *testing.T) {
	c := channel.New[int]()
	defer c.Close()

	if c.Len() != 0 {
		t.Fatalf("new channel should have Len==0")
	}

	for i := 0; i < 10; i++ {
		c.In <- i
	}

	// 允许 goroutine 有时间把数据搬到 dlink
	time.Sleep(1 * time.Millisecond)

	if got := c.Len(); got != 10 {
		t.Fatalf("Len want 10, got %d", got)
	}
	if cap := c.Cap(); cap < int64(10) {
		t.Fatalf("Cap should be at least 10, got %d", cap)
	}
}

// TestClose 重复 Close 不应 panic，且 Out 会被关闭
func TestClose(t *testing.T) {
	c := channel.New[int]()
	c.Close()
	c.Close() // 再次关闭应当安全

	if _, ok := <-c.Out; ok {
		t.Fatalf("out should be closed after Close()")
	}
}

// ---------- 基准测试 ----------

// BenchmarkSerial 单写单读——纯吞吐
func BenchmarkSerial(b *testing.B) {
	c := channel.New[int]()
	var wg sync.WaitGroup
	wg.Add(1)

	// drain goroutine
	go func() {
		for range c.Out {
		}
		wg.Done()
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.In <- i
	}
	b.StopTimer()

	c.Close()
	wg.Wait()
}

// BenchmarkParallel 并发写——测试内部锁/队列争用
func BenchmarkParallel(b *testing.B) {
	c := channel.New[int]()
	var wg sync.WaitGroup
	wg.Add(1)

	// drain goroutine
	go func() {
		for range c.Out {
		}
		wg.Done()
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			c.In <- i
		}
	})
	b.StopTimer()

	c.Close()
	wg.Wait()
}
