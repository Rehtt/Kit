package util

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// 单元测试部分

func TestNewBroadcaster(t *testing.T) {
	// 测试默认缓冲区大小
	b1 := NewBroadcaster[int]()
	if b1.chanBufSize != 1 {
		t.Errorf("默认缓冲区大小应为 1，实际为 %d", b1.chanBufSize)
	}

	// 测试自定义缓冲区大小
	b2 := NewBroadcaster[string](10)
	if b2.chanBufSize != 10 {
		t.Errorf("自定义缓冲区大小应为 10，实际为 %d", b2.chanBufSize)
	}
}

func TestSubscribeAndBroadcast(t *testing.T) {
	b := NewBroadcaster[int]()

	// 创建三个订阅者
	ch1 := b.Subscribe()
	ch2 := b.Subscribe()
	ch3 := b.Subscribe()

	// 确认订阅者数量
	if b.Len() != 3 {
		t.Errorf("应有 3 个订阅者，实际有 %d 个", b.Len())
	}

	// 广播消息
	testMsg := 42
	b.Broadcast(testMsg)

	// 验证所有订阅者都收到了消息
	for i, ch := range []<-chan int{ch1, ch2, ch3} {
		select {
		case msg := <-ch:
			if msg != testMsg {
				t.Errorf("订阅者 %d 收到的消息应为 %d，实际为 %d", i+1, testMsg, msg)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("订阅者 %d 未在预期时间内收到消息", i+1)
		}
	}
}

func TestUnsubscribe(t *testing.T) {
	b := NewBroadcaster[int]()

	// 创建三个订阅者
	ch1 := b.Subscribe()
	ch2 := b.Subscribe()
	ch3 := b.Subscribe()

	// 取消订阅第二个订阅者
	b.Unsubscribe(ch2)

	// 确认订阅者数量
	if b.Len() != 2 {
		t.Errorf("取消订阅后应有 2 个订阅者，实际有 %d 个", b.Len())
	}

	// 广播消息
	testMsg := 42
	b.Broadcast(testMsg)

	// 验证 ch1 和 ch3 收到消息，ch2 已关闭
	for i, ch := range []<-chan int{ch1, ch3} {
		select {
		case msg := <-ch:
			if msg != testMsg {
				t.Errorf("订阅者 %d 收到的消息应为 %d，实际为 %d", i+1, testMsg, msg)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("订阅者 %d 未在预期时间内收到消息", i+1)
		}
	}

	// 验证 ch2 已关闭
	select {
	case _, ok := <-ch2:
		if ok {
			t.Error("已取消订阅的 channel 应该已关闭")
		}
	default:
		t.Error("已取消订阅的 channel 应该已关闭且可读取")
	}
}

func TestUnsubscribeAll(t *testing.T) {
	b := NewBroadcaster[int]()

	// 创建多个订阅者
	channels := make([]<-chan int, 5)
	for i := 0; i < 5; i++ {
		channels[i] = b.Subscribe()
	}

	// 取消所有订阅
	b.UnsubscribesAll()

	// 确认订阅者数量为 0
	if b.Len() != 0 {
		t.Errorf("取消所有订阅后应有 0 个订阅者，实际有 %d 个", b.Len())
	}

	// 验证所有 channel 都已关闭
	for i, ch := range channels {
		select {
		case _, ok := <-ch:
			if ok {
				t.Errorf("订阅者 %d 的 channel 应该已关闭", i+1)
			}
		default:
			t.Errorf("订阅者 %d 的 channel 应该已关闭且可读取", i+1)
		}
	}
}

func TestSubscribeHandle(t *testing.T) {
	b := NewBroadcaster[int]()

	// 用于同步的 WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)

	// 收到的消息
	var receivedMsgs []int

	// 启动处理函数
	go func() {
		b.SubscribeHandle(func(msg int) error {
			receivedMsgs = append(receivedMsgs, msg)
			// 收到 3 后退出
			if msg == 3 {
				return errors.New("exit")
			}
			return nil
		})
		wg.Done()
	}()

	// 广播消息
	for i := 1; i <= 5; i++ {
		time.Sleep(10 * time.Millisecond) // 给处理函数一些时间
		b.Broadcast(i)
	}

	// 等待处理函数完成
	wg.Wait()

	// 验证只收到了 1, 2, 3
	expected := []int{1, 2, 3}
	if len(receivedMsgs) != len(expected) {
		t.Errorf("应收到 %d 条消息，实际收到 %d 条", len(expected), len(receivedMsgs))
	}

	for i, msg := range receivedMsgs {
		if i < len(expected) && msg != expected[i] {
			t.Errorf("第 %d 条消息应为 %d，实际为 %d", i+1, expected[i], msg)
		}
	}
}

func TestBroadcastAsync(t *testing.T) {
	b := NewBroadcaster[int]()

	// 创建订阅者
	ch := b.Subscribe()

	// 异步广播
	b.BroadcastAsync(42)

	// 验证消息是否收到
	select {
	case msg := <-ch:
		if msg != 42 {
			t.Errorf("收到的消息应为 42，实际为 %d", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("未在预期时间内收到异步广播的消息")
	}
}

func TestConcurrentOperations(t *testing.T) {
	b := NewBroadcaster[int](10) // 使用较大的缓冲区

	// 并发订阅和取消订阅
	var wg sync.WaitGroup
	wg.Add(2)

	// 并发订阅
	go func() {
		defer wg.Done()
		channels := make([]<-chan int, 0, 100)
		for i := 0; i < 100; i++ {
			channels = append(channels, b.Subscribe())
			// 随机广播一些消息
			if i%10 == 0 {
				b.BroadcastAsync(i)
			}
		}
	}()

	// 并发取消订阅
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			ch := b.Subscribe()
			time.Sleep(time.Millisecond)
			b.Unsubscribe(ch)
		}
	}()

	// 同时进行广播
	for i := 0; i < 20; i++ {
		b.BroadcastAsync(i * 100)
		time.Sleep(time.Millisecond)
	}

	wg.Wait()

	// 测试通过的标准是没有 panic
	t.Log("并发测试完成，订阅者数量:", b.Len())
}

// 封装公共基准逻辑，便于在 b.Run 中复用。
func benchBroadcast(b *testing.B, subs, bufSize int, async bool) {
	broadcaster := NewBroadcaster[int](bufSize)

	// 预先创建订阅者，并用 goroutine 清空消息，避免基准本身受到读取阻塞影响
	for i := 0; i < subs; i++ {
		ch := broadcaster.Subscribe()
		go func() {
			for range ch {
			}
		}()
	}

	b.ResetTimer() // 排除上面准备阶段的耗时

	for i := 0; i < b.N; i++ {
		if async {
			broadcaster.BroadcastAsync(i)
		} else {
			broadcaster.Broadcast(i)
		}
	}

	b.StopTimer()
	broadcaster.UnsubscribesAll()
}

// BenchmarkBroadcast 按订阅者数量、缓冲区大小、同步/异步维度做分组
func BenchmarkBroadcast(b *testing.B) {
	cases := []struct {
		name    string
		subs    int
		bufSize int
		async   bool
	}{
		// ------ 同步广播 ------
		{"Sync/Sub1/Buf1", 1, 1, false},
		{"Sync/Sub10/Buf1", 10, 1, false},
		{"Sync/Sub100/Buf1", 100, 1, false},
		{"Sync/Sub1000/Buf1", 1000, 1, false},
		{"Sync/Sub10/Buf0", 10, 0, false},   // 无缓冲
		{"Sync/Sub10/Buf64", 10, 64, false}, // 大缓冲
		// ------ 异步广播 ------
		{"Async/Sub10/Buf1", 10, 1, true},
		{"Async/Sub100/Buf1", 100, 1, true},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			benchBroadcast(b, c.subs, c.bufSize, c.async)
		})
	}
}

// BenchmarkSubscribeUnsubscribe 关注订阅/取消本身的开销
func BenchmarkSubscribeUnsubscribe(b *testing.B) {
	broadcaster := NewBroadcaster[int]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := broadcaster.Subscribe()
		broadcaster.Unsubscribe(ch)
	}
	b.StopTimer()
}

// BenchmarkConcurrentSubscribeUnsubscribe 测试并发订阅/取消的性能
func BenchmarkConcurrentSubscribeUnsubscribe(b *testing.B) {
	broadcaster := NewBroadcaster[int]()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch := broadcaster.Subscribe()
			broadcaster.Unsubscribe(ch)
		}
	})
	b.StopTimer()
}

// BenchmarkConcurrentBroadcast 测试并发广播的性能
func BenchmarkConcurrentBroadcast(b *testing.B) {
	cases := []struct {
		name    string
		subs    int
		bufSize int
	}{
		{"Sub10/Buf1", 10, 1},
		{"Sub100/Buf1", 100, 1},
		{"Sub10/Buf64", 10, 64},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			broadcaster := NewBroadcaster[int](c.bufSize)

			// 创建订阅者
			for i := 0; i < c.subs; i++ {
				ch := broadcaster.Subscribe()
				go func() {
					for range ch {
					}
				}()
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				counter := 0
				for pb.Next() {
					broadcaster.Broadcast(counter)
					counter++
				}
			})

			b.StopTimer()
			broadcaster.UnsubscribesAll()
		})
	}
}
