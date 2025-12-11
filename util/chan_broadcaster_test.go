package util

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ===========================
// 基础功能测试
// ===========================

func TestNewBroadcaster(t *testing.T) {
	b1 := NewBroadcaster[int]()
	if b1.chanBufSize != 1 {
		t.Errorf("默认缓冲区大小应为 1，实际为 %d", b1.chanBufSize)
	}

	b2 := NewBroadcaster[string](10)
	if b2.chanBufSize != 10 {
		t.Errorf("自定义缓冲区大小应为 10，实际为 %d", b2.chanBufSize)
	}
}

func TestSubscribeAndBroadcast(t *testing.T) {
	b := NewBroadcaster[int]()
	defer b.Close()

	ch1 := b.Subscribe()
	ch2 := b.Subscribe()

	if b.Len() != 2 {
		t.Errorf("期望 2 个订阅者，实际 %d", b.Len())
	}

	testMsg := 42
	sent := b.Broadcast(testMsg)
	if sent != 2 {
		t.Errorf("期望发送给 2 个订阅者，实际发送成功 %d", sent)
	}

	for i, ch := range []<-chan int{ch1, ch2} {
		select {
		case msg := <-ch:
			if msg != testMsg {
				t.Errorf("订阅者 %d 消息不匹配: 期望 %d, 实际 %d", i, testMsg, msg)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("订阅者 %d 超时未收到消息", i)
		}
	}
}

// ===========================
// 核心逻辑测试：丢包与阻塞
// ===========================

func TestBroadcast_DropMessage(t *testing.T) {
	// 缓冲区为 1
	b := NewBroadcaster[int](1)
	defer b.Close()

	ch := b.Subscribe()

	// 1. 发送第一条，填满缓冲区
	b.Broadcast(1)

	// 2. 发送第二条，因为消费者没读，缓冲区已满，应该丢弃
	// 注意：Broadcast 返回的是发送成功的数量
	sent2 := b.Broadcast(2)
	if sent2 != 0 {
		t.Errorf("第二条消息应该因为缓冲区满被丢弃，但显示发送成功了 %d 个", sent2)
	}

	// 3. 消费第一条
	msg := <-ch
	if msg != 1 {
		t.Errorf("读到的应该是第一条消息 1，实际是 %d", msg)
	}

	// 4. 验证没有第二条消息
	select {
	case val := <-ch:
		t.Errorf("不应该收到第二条消息，但收到了: %d", val)
	default:
		// 正常
	}
}

func TestBroadcastSync_Timeout(t *testing.T) {
	// 缓冲区为 0，必须同步读写
	b := NewBroadcaster[int](0)
	defer b.Close()

	_ = b.Subscribe() // 阻塞的订阅者

	start := time.Now()
	// 设置 50ms 超时
	sent := b.BroadcastSync(100, 50*time.Millisecond)
	elapsed := time.Since(start)

	if sent != 0 {
		t.Errorf("订阅者阻塞，应该发送失败(0)，实际返回 %d", sent)
	}

	// 允许少量误差
	if elapsed < 40*time.Millisecond {
		t.Errorf("BroadcastSync 应该阻塞约 50ms，实际阻塞 %v", elapsed)
	}
}

func TestBroadcastSync_InfiniteWait(t *testing.T) {
	b := NewBroadcaster[int](0)
	defer b.Close()

	ch := b.Subscribe()

	// 启动一个延迟读取的 goroutine
	go func() {
		time.Sleep(100 * time.Millisecond)
		<-ch
	}()

	start := time.Now()
	// timeout=0 表示无限等待
	sent := b.BroadcastSync(200, 0)
	elapsed := time.Since(start)

	if sent != 1 {
		t.Errorf("消息应该最终发送成功，实际返回 %d", sent)
	}

	if elapsed < 100*time.Millisecond {
		t.Errorf("应该阻塞直到消费者读取(>100ms)，实际只用了 %v", elapsed)
	}
}

func TestBroadcastSync_Mixed(t *testing.T) {
	b := NewBroadcaster[int](0)
	defer b.Close()

	// 1. 快速消费者
	chFast := b.Subscribe()
	go func() {
		for range chFast {
		}
	}()

	// 2. 阻塞消费者 (不读)
	_ = b.Subscribe()
	// 3. 阻塞消费者 (不读)
	_ = b.Subscribe()

	// 广播，超时设为 20ms
	start := time.Now()
	// 因为是串行分发，遇到两个阻塞者，总耗时应在 40ms 左右
	sent := b.BroadcastSync(999, 20*time.Millisecond)
	elapsed := time.Since(start)

	if sent != 1 {
		t.Errorf("期望 1 个发送成功，实际 %d", sent)
	}

	if elapsed < 35*time.Millisecond {
		t.Errorf("期望耗时至少约 40ms，实际 %v", elapsed)
	}
}

// ===========================
// 订阅管理与并发测试
// ===========================

func TestUnsubscribe(t *testing.T) {
	b := NewBroadcaster[int]()
	ch1 := b.Subscribe()
	ch2 := b.Subscribe()

	b.Unsubscribe(ch1)

	if b.Len() != 1 {
		t.Errorf("Unsubscribe 后长度应为 1，实际 %d", b.Len())
	}

	// 验证 ch1 已关闭
	select {
	case _, ok := <-ch1:
		if ok {
			t.Error("ch1 应该已关闭")
		}
	default:
		t.Error("ch1 应该已关闭且可读(零值)")
	}

	// 验证 ch2 正常工作
	b.Broadcast(1)
	select {
	case v := <-ch2:
		if v != 1 {
			t.Error("ch2 接收错误")
		}
	default:
		t.Error("ch2 应收到消息")
	}
}

func TestUnsubscribe_Idempotent(t *testing.T) {
	b := NewBroadcaster[int]()
	ch := b.Subscribe()

	b.Unsubscribe(ch)
	b.Unsubscribe(ch) // 重复取消不应 panic

	if b.Len() != 0 {
		t.Error("订阅者数量应为 0")
	}
}

func TestUnsubscribeAll(t *testing.T) {
	b := NewBroadcaster[int]()
	ch1 := b.Subscribe()
	ch2 := b.Subscribe()

	b.UnsubscribesAll()

	if b.Len() != 0 {
		t.Errorf("清空后长度应为 0，实际 %d", b.Len())
	}

	for _, ch := range []<-chan int{ch1, ch2} {
		if _, ok := <-ch; ok {
			t.Error("Channel 应该已关闭")
		}
	}
}

// 修复后的 TestSubscribeHandle
func TestSubscribeHandle(t *testing.T) {
	// 【关键修复】使用 Buffer 10，防止发送太快导致 msg=3 被丢弃
	// 如果用默认 Buffer 1，下面的 Broadcast 连续调用会导致丢包，从而导致死锁
	b := NewBroadcaster[int](10)
	defer b.Close()

	var count int32
	done := make(chan struct{})

	go func() {
		// 只有当收到 3 时才退出，如果 3 被丢包，这里就会永久阻塞
		_ = b.SubscribeHandle(func(msg int) error {
			atomic.AddInt32(&count, 1)
			if msg == 3 {
				return errors.New("stop")
			}
			return nil
		})
		close(done)
	}()

	// 稍微等待订阅建立
	time.Sleep(10 * time.Millisecond)

	b.Broadcast(1)
	b.Broadcast(2)
	b.Broadcast(3) // 关键消息
	b.Broadcast(4) // 应该被忽略，因为 handle 已经退出

	select {
	case <-done:
		// 成功
	case <-time.After(1 * time.Second):
		t.Fatal("测试超时：SubscribeHandle 未能收到消息 3 并退出 (可能是发生了丢包)")
	}

	val := atomic.LoadInt32(&count)
	if val != 3 {
		t.Errorf("应该处理 3 条消息，实际处理了 %d 条", val)
	}
}

func TestBroadcastAsync(t *testing.T) {
	b := NewBroadcaster[int]()
	defer b.Close()
	ch := b.Subscribe()

	b.BroadcastAsync(100)

	select {
	case v := <-ch:
		if v != 100 {
			t.Errorf("值错误: %d", v)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("异步广播超时")
	}
}

func TestConcurrentSafety(t *testing.T) {
	b := NewBroadcaster[int](10)
	defer b.Close()

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			b.Broadcast(i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			ch := b.Subscribe()
			time.Sleep(time.Microsecond)
			b.Unsubscribe(ch)
		}
	}()

	go func() {
		defer wg.Done()
		for k := 0; k < 5; k++ {
			go func() {
				ch := b.Subscribe()
				for range ch {
				}
			}()
		}
	}()

	wg.Wait()
}

// ===========================
// 性能基准测试
// ===========================

func benchBroadcast(b *testing.B, subs, bufSize int) {
	broadcaster := NewBroadcaster[int](bufSize)

	for i := 0; i < subs; i++ {
		ch := broadcaster.Subscribe()
		go func() {
			for range ch {
			}
		}()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		broadcaster.Broadcast(i)
	}
	b.StopTimer()
	broadcaster.UnsubscribesAll()
}

func BenchmarkBroadcast_Sync(b *testing.B) {
	cases := []struct {
		name    string
		subs    int
		bufSize int
	}{
		{"Sub1/Buf1", 1, 1},
		{"Sub10/Buf1", 10, 1},
		{"Sub100/Buf1", 100, 1},
		{"Sub100/Buf64", 100, 64},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			benchBroadcast(b, c.subs, c.bufSize)
		})
	}
}

// BenchmarkBroadcastAsync_Parallel 使用 RunParallel 测试异步广播。
// 这是测试 Async 场景的正确方式，避免一次性创建百万个 goroutine 导致崩溃。
func BenchmarkBroadcastAsync_Parallel(b *testing.B) {
	broadcaster := NewBroadcaster[int](100)
	// 只有1个消费者，单纯测试发送端开销
	ch := broadcaster.Subscribe()
	go func() {
		for range ch {
		}
	}()

	b.ResetTimer()

	// 使用 RunParallel 限制并发数等于 GOMAXPROCS
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			broadcaster.BroadcastAsync(i)
			i++
		}
	})

	b.StopTimer()
	broadcaster.UnsubscribesAll()
}
