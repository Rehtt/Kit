package util

import (
	"sync"
)

type Broadcaster[T any] struct {
	mu            sync.Mutex
	subscriberArr []chan T
	chanBufSize   int
}

// NewBroadcaster 创建一个新的 Broadcaster
// chanBufSize 设置 channel 的缓冲大小,默认为 1
func NewBroadcaster[T any](chanBufSize ...int) *Broadcaster[T] {
	csize := 1
	if len(chanBufSize) > 0 {
		csize = chanBufSize[0]
	}

	out := &Broadcaster[T]{
		chanBufSize: csize,
	}
	return out
}

// Subscribe 返回一个新的接收 channel，订阅后可以从该 channel 读取广播消息
func (b *Broadcaster[T]) Subscribe() <-chan T {
	ch := make(chan T, b.chanBufSize) // 带缓冲，避免阻塞发布者
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscriberArr = append(b.subscriberArr, ch)
	return ch
}

// SubscribeHandle 类似于 Subscribe，但是不返回 channel，而是通过函数进行处理
func (b *Broadcaster[T]) SubscribeHandle(f func(T) error) (err error) {
	ch := b.Subscribe()
	defer b.Unsubscribe(ch)
	for msg := range ch {
		if err = f(msg); err != nil {
			return err
		}
	}
	return nil
}

// Unsubscribe 取消订阅，关闭对应 channel
func (b *Broadcaster[T]) Unsubscribe(ch <-chan T) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, subch := range b.subscriberArr {
		if subch == ch {
			lastIndex := b.Len() - 1
			if i != lastIndex {
				b.subscriberArr[i] = b.subscriberArr[lastIndex]
			}
			b.subscriberArr[lastIndex] = nil
			b.subscriberArr = b.subscriberArr[:lastIndex]

			close(subch)
			break
		}
	}
}

// Broadcast 非阻塞式将 msg 同步分发给所有活跃的订阅者
// *如果某个订阅者没及时读取，就丢弃这条消息以免阻塞调用者，因此不能保证消息100%送达所有订阅者*
// 适用于 允许丢包 的实时通知场景
// 不适用 于任务分发或强一致性消息队列
func (b *Broadcaster[T]) Broadcast(msg T) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	var sentCount int
	for _, ch := range b.subscriberArr {
		select {
		case ch <- msg:
			// 发送成功
			sentCount++
		default:
			// 如果某个订阅者没及时读取，就丢弃这条消息以免阻塞
		}
	}
	return sentCount
}

// BroadcastWaitAll 阻塞式将 msg 100%同步分发给所有订阅者
// *如果某个订阅者一直不读，会卡死调用者*
// *如果订阅者过多且发送频繁可能会导致内存溢出*
// 适用于 任务分发或强一致性消息队列
func (b *Broadcaster[T]) BroadcastWaitAll(msg T) int {
	b.mu.Lock()
	// 复制一份订阅者列表
	subs := make([]chan T, len(b.subscriberArr))
	copy(subs, b.subscriberArr)
	b.mu.Unlock()

	var sentCount int

	for _, ch := range subs {
		SafeRun(func() {
			ch <- msg
			sentCount++
		}, nil)
	}
	return sentCount
}

// BroadcastAsync 异步分发
func (b *Broadcaster[T]) BroadcastAsync(msg T) {
	go b.Broadcast(msg)
}

// UnsubscribesAll 取消所有订阅，关闭所有 channel
func (b *Broadcaster[T]) UnsubscribesAll() {
	b.mu.Lock()
	for i, ch := range b.subscriberArr {
		close(ch)
		b.subscriberArr[i] = nil
	}
	b.subscriberArr = b.subscriberArr[:0]
	b.mu.Unlock()
}

func (b *Broadcaster[T]) Close() {
	b.UnsubscribesAll()
}

func (b *Broadcaster[T]) Len() int {
	return len(b.subscriberArr)
}
