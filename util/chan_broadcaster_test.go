package util

import (
	"testing"
)

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
