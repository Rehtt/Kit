package util

import (
	"slices"
	"sync"
)

type Broadcaster[T any] struct {
	mu            sync.Mutex
	subscribers   map[<-chan T]int
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
		subscribers: make(map[<-chan T]int),
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
	b.subscribers[ch] = len(b.subscriberArr) - 1
	return ch
}

// SubscribeHandle 类似于 Subscribe，但是不返回 channel，而是通过函数进行处理
func (b *Broadcaster[T]) SubscribeHandle(f func(T) (exit bool)) {
	ch := b.Subscribe()
	defer b.Unsubscribe(ch)
	for msg := range ch {
		if f(msg) {
			return
		}
	}
}

// Unsubscribe 取消订阅，关闭对应 channel
func (b *Broadcaster[T]) Unsubscribe(ch <-chan T) {
	b.mu.Lock()
	if index, ok := b.subscribers[ch]; ok {
		close(b.subscriberArr[index])
		delete(b.subscribers, ch)
		b.subscriberArr = slices.Delete(b.subscriberArr, index, index+1)
	}
	b.mu.Unlock()
}

// Broadcast 将 msg 同步分发给所有活跃的订阅者
func (b *Broadcaster[T]) Broadcast(msg T) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subscriberArr {
		select {
		case ch <- msg:
			// 发送成功
		default:
			// 如果某个订阅者没及时读取，就丢弃这条消息以免阻塞
		}
	}
}

// BroadcastAsync 异步分发
func (b *Broadcaster[T]) BroadcastAsync(msg T) {
	go b.Broadcast(msg)
}

// UnsubscribesAll 取消所有订阅，关闭所有 channel
func (b *Broadcaster[T]) UnsubscribesAll() {
	b.mu.Lock()
	for _, ch := range b.subscriberArr {
		close(ch)
	}
	b.subscribers = make(map[<-chan T]int)
	b.subscriberArr = b.subscriberArr[:0]
	b.mu.Unlock()
}

func (b *Broadcaster[T]) Close() {
	b.UnsubscribesAll()
}

func (b *Broadcaster[T]) Len() int {
	return len(b.subscriberArr)
}
