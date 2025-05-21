package util

import "sync"

type Broadcaster[T any] struct {
	mu          sync.Mutex
	subscribers map[<-chan T]chan T
}

// NewBroadcaster 创建一个新的 Broadcaster
func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		subscribers: make(map[<-chan T]chan T),
	}
}

// Subscribe 返回一个新的接收 channel，订阅后可以从该 channel 读取广播消息
func (b *Broadcaster[T]) Subscribe() <-chan T {
	ch := make(chan T, 1) // 带缓冲，避免阻塞发布者
	b.mu.Lock()
	b.subscribers[ch] = ch
	b.mu.Unlock()
	return ch
}

// SubscribeHandle 类似于 Subscribe，但是不返回 channel，而是通过函数进行处理
func (b *Broadcaster[T]) SubscribeHandle(f func(T)) {
	for msg := range b.Subscribe() {
		f(msg)
	}
}

// Unsubscribe 取消订阅，关闭对应 channel
func (b *Broadcaster[T]) Unsubscribe(ch <-chan T) {
	b.mu.Lock()
	if c, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(c)
	}
	b.mu.Unlock()
}

// Broadcast 将 msg 同步分发给所有活跃的订阅者
func (b *Broadcaster[T]) Broadcast(msg T) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subscribers {
		select {
		case ch <- msg:
			// 发送成功
		default:
			// 如果某个订阅者没及时读取，就丢弃这条消息以免阻塞
		}
	}
}
