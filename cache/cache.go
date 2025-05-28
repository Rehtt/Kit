package cache

import (
	"container/heap"
	"sync"
	"time"
)

type Item struct {
	key        string
	value      any
	expiration int64
	index      int // add index for heap.Interface
}

// Deprecated: use github.com/Rehtt/Kit/maps
type Cache struct {
	items             map[string]*Item
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	mu                sync.RWMutex
	expHeap           ExpirationHeap // use ExpirationHeap instead of heap.MinHeap
}

type ExpirationHeap []*Item

var itemPool = sync.Pool{
	New: func() any {
		return &Item{}
	},
}

func (eh ExpirationHeap) Len() int {
	return len(eh)
}

func (eh ExpirationHeap) Less(i, j int) bool {
	return eh[i].expiration < eh[j].expiration
}

func (eh ExpirationHeap) Swap(i, j int) {
	eh[i], eh[j] = eh[j], eh[i]
	eh[i].index = i
	eh[j].index = j
}

func (eh *ExpirationHeap) Push(x any) {
	n := len(*eh)
	item := x.(*Item)
	item.index = n
	*eh = append(*eh, item)
}

func (eh *ExpirationHeap) Pop() any {
	old := *eh
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*eh = old[0 : n-1]
	return item
}

// Deprecated: use github.com/Rehtt/Kit/maps
func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	items := make(map[string]*Item)
	expHeap := make(ExpirationHeap, 0)
	cache := &Cache{
		items:             items,
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		expHeap:           expHeap,
	}
	heap.Init(&cache.expHeap) // initialize the heap
	go cache.cleanupExpired()
	return cache
}

func (c *Cache) Set(key string, value any, expiration time.Duration) {
	item := itemPool.Get().(*Item)
	item.value = value
	item.key = key
	now := time.Now().UnixNano()
	if expiration == 0 {
		item.expiration = now + int64(c.defaultExpiration)
	} else {
		item.expiration = now + int64(expiration)
	}
	c.mu.Lock()
	_, found := c.items[key]
	if found {
		heap.Remove(&c.expHeap, item.index)
	}
	c.items[key] = item
	heap.Push(&c.expHeap, item)
	c.mu.Unlock()
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().UnixNano() < item.expiration {
		return nil, false
	}
	return item.value, true
}

func (c *Cache) cleanupExpired() {
	for {
		time.Sleep(c.cleanupInterval)

		c.mu.Lock()
		for c.expHeap.Len() > 0 {
			item := heap.Pop(&c.expHeap).(*Item)
			if item.expiration > time.Now().UnixNano() {
				heap.Push(&c.expHeap, item) // 将未过期的item放回heap
				break                       // 如果最早过期的item都未过期，则退出循环
			}
			delete(c.items, item.key)
			itemPool.Put(item)
		}
		c.mu.Unlock()
	}
}
