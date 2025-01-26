package util

import (
	"hash/fnv"
	"sync"
)

var SHARD_COUNT uint32 = 32

// 分成SHARD_COUNT个分片的map
type ConcurrentMap[T any] []*ConcurrentMapShared[T]

// 通过RWMutex保护的线程安全的分片，包含一个map
type ConcurrentMapShared[T any] struct {
	items map[string]T
	rw    sync.RWMutex
}

// 创建并发map
func NewConcurrentMap[T any]() ConcurrentMap[T] {
	m := make(ConcurrentMap[T], SHARD_COUNT)
	for i := uint32(0); i < SHARD_COUNT; i++ {
		m[i] = &ConcurrentMapShared[T]{items: make(map[string]T)}
	}
	return m
}

// 根据key计算分片索引
func (m ConcurrentMap[T]) GetShard(key string) *ConcurrentMapShared[T] {
	f := fnv.New32a()
	f.Write([]byte(key))

	return m[f.Sum32()%SHARD_COUNT]
}

func (m ConcurrentMap[T]) Set(key string, value T) {
	// 根据key计算出对应的分片
	shard := m.GetShard(key)
	shard.Set(key, value)
}

func (m ConcurrentMap[T]) SetByFunc(key string, newValueFunc func(oldValue T) (newValue T)) (newValue T) {
	shard := m.GetShard(key)
	return shard.SetByFunc(key, newValueFunc)
}

func (m ConcurrentMap[T]) Get(key string) (T, bool) {
	// 根据key计算出对应的分片
	shard := m.GetShard(key)
	return shard.Get(key)
}

func (m ConcurrentMap[T]) Delete(key string) {
	// 根据key计算出对应的分片
	shard := m.GetShard(key)
	shard.Delete(key)
}

func (s *ConcurrentMapShared[T]) Get(key string) (T, bool) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	val, ok := s.items[key]
	return val, ok
}

func (s *ConcurrentMapShared[T]) Set(key string, value T) {
	s.rw.Lock()
	defer s.rw.Unlock()
	s.items[key] = value
}

func (s *ConcurrentMapShared[T]) SetByFunc(key string, newValueFunc func(oldValue T) (newValue T)) (newValue T) {
	s.rw.Lock()
	defer s.rw.Unlock()
	oldValue := s.items[key]
	newValue = newValueFunc(oldValue)
	s.items[key] = newValue
	return
}

func (s *ConcurrentMapShared[T]) Delete(key string) {
	s.rw.Lock()
	defer s.rw.Unlock()
	delete(s.items, key)
}
