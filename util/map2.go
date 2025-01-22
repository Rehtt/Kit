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
	sync.RWMutex
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
	shard.Lock()
	defer shard.Unlock()
	shard.items[key] = value
}

func (m ConcurrentMap[T]) SetByFunc(key string, newValueFunc func(oldValue T) (newValue T)) (newValue T) {
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	oldValue := shard.items[key]
	newValue = newValueFunc(oldValue)
	shard.items[key] = newValue
	return
}

func (m ConcurrentMap[T]) Get(key string) (T, bool) {
	// 根据key计算出对应的分片
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()
	val, ok := shard.items[key]
	return val, ok
}

func (m ConcurrentMap[T]) Delete(key string) {
	// 根据key计算出对应的分片
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	delete(shard.items, key)
}
