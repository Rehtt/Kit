package util

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --------- 基本并发处理 ------------------------------------------------------

func TestWorkerPool_ProcessAllJobs(t *testing.T) {
	const N = 100
	var processed int64 // 统计已处理数量
	var wg sync.WaitGroup
	wg.Add(N)

	// 每处理一个元素就 ++ 并 Done
	pool := NewWorkerPool[int](func(i int) {
		atomic.AddInt64(&processed, 1)
		wg.Done()
	}, 3) // 自定义 poolSize=3
	defer pool.Close()

	for i := 0; i < N; i++ {
		pool.Do(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if got := atomic.LoadInt64(&processed); got != N {
			t.Fatalf("want %d jobs processed, got %d", N, got)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for jobs to finish")
	}
}

// --------- 默认大小 -----------------------------------------------------------

func TestWorkerPool_DefaultPoolSize(t *testing.T) {
	pool := NewWorkerPool[int](func(int) {})
	defer pool.Close()

	if pool.poolSize != defaultPoolSize {
		t.Fatalf("default pool size: want %d, got %d", defaultPoolSize, pool.poolSize)
	}
}

// --------- 通道容量符合 GOMAXPROCS 逻辑 ---------------------------------------

func TestWorkerPool_ChannelCapacity(t *testing.T) {
	pool := NewWorkerPool[int](func(int) {})
	defer pool.Close()

	wantCap := 1
	if runtime.GOMAXPROCS(0) == 1 {
		wantCap = 0 // 无缓冲
	}
	if cap(pool.ch) != wantCap {
		t.Fatalf("channel capacity: want %d, got %d", wantCap, cap(pool.ch))
	}
}

// --------- Close 可正常关闭 ----------------------------------------------------

func TestWorkerPool_Close(t *testing.T) {
	pool := NewWorkerPool[int](func(int) {})
	// 先发送一个任务，确保 goroutine 已经启动
	pool.Do(1)
	pool.Close()

	// 再次调用 Close 应 panic；使用 recover 断言
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("second Close() should panic, but did not")
		}
	}()
	pool.Close()
}
