package maps

import (
	"hash/fnv"
	"sync"
	"time"
)

var (
	SHARD_COUNT         uint32 = 32               // 只能在NewConcurrentMap之前修改
	AUTO_CLEAR_INTERVAL        = 10 * time.Minute // 可修改
)

// 分成SHARD_COUNT个分片的map
type ConcurrentMap[T any] struct {
	maps   []*RWMutexMap[T]
	option *Option

	runautoclear sync.Once
}

type Option struct {
	ttl time.Duration
}

// 启用过期时间
func EnableExpired(ttl time.Duration) func(option *Option) {
	return func(option *Option) {
		option.ttl = ttl
	}
}

// 创建并发map
func NewConcurrentMap[T any](options ...func(option *Option)) *ConcurrentMap[T] {
	m := ConcurrentMap[T]{
		maps:   make([]*RWMutexMap[T], SHARD_COUNT),
		option: &Option{},
	}
	for _, f := range options {
		f(m.option)
	}
	for i := uint32(0); i < SHARD_COUNT; i++ {
		m.maps[i] = NewRWMutexMap[T](m.option.ttl)
	}

	if m.option.ttl > 0 {
		m.autoClear()
	}
	return &m
}

// 根据key计算分片索引
func (m *ConcurrentMap[T]) GetShard(key string) uint32 {
	f := fnv.New32a()
	f.Write([]byte(key))

	return f.Sum32() % SHARD_COUNT
}

func (m *ConcurrentMap[T]) Set(key string, value T, ttl ...time.Duration) {
	// 根据key计算出对应的分片
	index := m.GetShard(key)
	m.maps[index].Set(key, value, ttl...)

	if len(ttl) > 0 {
		m.autoClear()
	}
}

func (m *ConcurrentMap[T]) SetByFunc(key string, newValueFunc func(oldValue T) (newValue T), ttl ...time.Duration) (newValue T) {
	index := m.GetShard(key)
	newValue = m.maps[index].SetByFunc(key, newValueFunc, ttl...)

	if len(ttl) > 0 {
		m.autoClear()
	}

	return newValue
}

func (m *ConcurrentMap[T]) Get(key string) (value T, ok bool) {
	// 根据key计算出对应的分片
	index := m.GetShard(key)

	return m.maps[index].Get(key)
}

func (m *ConcurrentMap[T]) Delete(key string) {
	// 根据key计算出对应的分片
	index := m.GetShard(key)
	m.maps[index].Delete(key)
}

// 定期清理
func (m *ConcurrentMap[T]) autoClear() {
	m.runautoclear.Do(func() {
		go func() {
			for {
				time.Sleep(AUTO_CLEAR_INTERVAL)
				for _, v := range m.maps {
					v.Clear()
				}
			}
		}()
	})
}
