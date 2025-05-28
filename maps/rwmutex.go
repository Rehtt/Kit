package maps

import (
	"container/heap"
	"sync"
	"time"
)

type Node[T any] struct {
	Key           string
	Value         T
	ExpirtimeUnix int64
}

// 通过RWMutex保护的线程安全的分片，包含一个map
type RWMutexMap[T any] struct {
	items map[string]*Node[T]
	rw    sync.RWMutex

	ttl          time.Duration
	expirtimeArr ExpirationHeap[T]
	nodePool     sync.Pool
}

func NewRWMutexMap[T any](ttl time.Duration) *RWMutexMap[T] {
	return &RWMutexMap[T]{
		items:    map[string]*Node[T]{},
		ttl:      ttl,
		nodePool: sync.Pool{New: func() any { return &Node[T]{} }},
	}
}

func (s *RWMutexMap[T]) Get(key string) (val T, ok bool) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	return s.getValidValue(key)
}

func (s *RWMutexMap[T]) Set(key string, value T, ttl ...time.Duration) {
	s.rw.Lock()
	defer s.rw.Unlock()
	n := s.nodePool.Get().(*Node[T])
	n.Value = value
	n.Key = key
	s.items[key] = n

	s.resetExpirtime(n, ttl...)
}

func (s *RWMutexMap[T]) SetByFunc(key string, newValueFunc func(oldValue T) (newValue T), ttl ...time.Duration) (newValue T) {
	s.rw.Lock()
	defer s.rw.Unlock()
	n := s.items[key]
	if n == nil {
		n = s.nodePool.Get().(*Node[T])
		n.Key = key
	}
	newValue = newValueFunc(n.Value)
	n.Value = newValue
	s.resetExpirtime(n, ttl...)
	return
}

func (s *RWMutexMap[T]) Delete(key string) {
	s.rw.Lock()
	defer s.rw.Unlock()
	delete(s.items, key)
}

func (s *RWMutexMap[T]) resetExpirtime(node *Node[T], ttl ...time.Duration) {
	// 过期时间
	if s.ttl == 0 && len(ttl) == 0 {
		return
	}
	if len(ttl) > 0 && ttl[0] > 0 {
		node.ExpirtimeUnix = time.Now().Add(ttl[0]).Unix()
		s.expirtimeArr.Push(node)
	} else if s.ttl > 0 {
		node.ExpirtimeUnix = time.Now().Add(s.ttl).Unix()
		s.expirtimeArr.Push(node)
	}
}

func (s *RWMutexMap[T]) getValidValue(key string) (_ T, _ bool) {
	n, ok := s.items[key]
	if !ok {
		return
	}
	if n.ExpirtimeUnix != 0 && time.Now().Unix() >= n.ExpirtimeUnix {
		return
	}
	return n.Value, true
}

// 清理过期的值
func (s *RWMutexMap[T]) Clear() {
	if s.expirtimeArr.Len() == 0 {
		return
	}
	now := time.Now().Unix()
	s.rw.Lock()
	defer s.rw.Unlock()
	// 一次heap.Init+多次heap.Pop 的时间成本比 多次heap.Push+多次heap.Pop 的时间成本要小
	heap.Init(&s.expirtimeArr)
	// heap.Init整理后的可以保证[0]是最小的
	for s.expirtimeArr.Len() > 0 && s.expirtimeArr[0].ExpirtimeUnix <= now {
		n := heap.Pop(&s.expirtimeArr).(*Node[T])
		delete(s.items, n.Key)
		s.nodePool.Put(n)
	}
}
