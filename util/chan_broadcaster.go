package util

import (
	"sync"
	"time"
)

type Broadcaster[T any] struct {
	mu            sync.RWMutex
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
			lastIndex := len(b.subscriberArr) - 1
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
	b.mu.RLock()
	defer b.mu.RUnlock()

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

// BroadcastWaitAll 阻塞式将 msg 同步分发给所有订阅者
// 默认50ms兜底，防止卡死整个广播调度
// *当timeout为0时不会丢弃消息，但如果某个订阅者一直不读，会卡死调用者和整个广播调度*
// 适用于 任务分发或强一致性消息队列
// BroadcastSync 阻塞式将 msg 分发给所有订阅者 (同步发送)
// 默认带有 50ms 超时兜底，防止单个慢速订阅者卡死整个广播系统。
//
//	timeout: 可选。
//	  - 不传: 默认 50ms 超时。
//	  - 传 0: 无限等待 (慎用! 可能导致死锁)。
//	  - 传 >0: 指定超时时间。
//
// 适用场景: 必须送达的关键通知 (如: 拍卖成交、倒计时结束、任务分发或强一致性消息队列)
// 区别:
//   - Broadcast: 非阻塞，缓冲区满则丢弃 (Fire-and-Forget)。
//   - BroadcastSync: 阻塞等待，直到发送成功或超时 (Reliable-ish)。
func (b *Broadcaster[T]) BroadcastSync(msg T, timeout ...time.Duration) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// 默认50ms兜底，防止卡死整个广播调度
	waitTime := 50 * time.Millisecond
	if len(timeout) > 0 {
		waitTime = timeout[0]
	}
	var sentCount int

	// 无限等待模式
	// *当timeout为0时，如果某个订阅者一直不读，会卡死调用者和整个广播调度*
	if waitTime <= 0 {
		for _, ch := range b.subscriberArr {
			ch <- msg
			sentCount++
		}
		return sentCount
	}

	// 超时模式
	timer := time.NewTimer(waitTime)
	defer timer.Stop()
	for _, ch := range b.subscriberArr {
		// 排空timer确保是重置状态
		// 如果是第一次循环，timer 刚创建 running，Stop() 返回 true，不会阻塞
		// 如果是后续循环，Stop() 负责清理可能的残留
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(waitTime)

		select {
		case ch <- msg:
			sentCount++
		case <-timer.C:
			// 超时跳过
		}
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
	defer b.mu.Unlock()

	for i, ch := range b.subscriberArr {
		close(ch)
		b.subscriberArr[i] = nil
	}
	b.subscriberArr = b.subscriberArr[:0]
}

func (b *Broadcaster[T]) Close() {
	b.UnsubscribesAll()
}

func (b *Broadcaster[T]) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscriberArr)
}
