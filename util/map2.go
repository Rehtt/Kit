package util

import (
	"hash/fnv"
	"sync"
)

var SHARD_COUNT = 32

// 分成SHARD_COUNT个分片的map
type ConcurrentMap []*ConcurrentMapShared

// 通过RWMutex保护的线程安全的分片，包含一个map
type ConcurrentMapShared struct {
	items map[string]any
	sync.RWMutex
}

// 创建并发map
func NewConcurrentMap() ConcurrentMap {
	m := make(ConcurrentMap, SHARD_COUNT)
	for i := 0; i < SHARD_COUNT; i++ {
		m[i] = &ConcurrentMapShared{items: make(map[string]any)}
	}
	return m
}

// 根据key计算分片索引
func (m ConcurrentMap) GetShard(key string) *ConcurrentMapShared {
	f := fnv.New32a()
	f.Write([]byte(key))

	return m[uint(f.Sum32())%uint(SHARD_COUNT)]
}

func (m ConcurrentMap) Set(key string, value any) {
	// 根据key计算出对应的分片
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	shard.items[key] = value
}

func (m ConcurrentMap) SetByFunc(key string, newValueFunc func(oldValue any) (newValue any)) (newValue any) {
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	oldValue := shard.items[key]
	newValue = newValueFunc(oldValue)
	shard.items[key] = newValue
	return
}

func (m ConcurrentMap) Get(key string) (any, bool) {
	// 根据key计算出对应的分片
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()
	val, ok := shard.items[key]
	return val, ok
}
