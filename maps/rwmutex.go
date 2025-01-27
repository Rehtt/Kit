package maps

import (
	"sync"
	"time"
)

// 通过RWMutex保护的线程安全的分片，包含一个map
type RWMutexMap[T any] struct {
	items map[string]T
	rw    sync.RWMutex

	ttl       time.Duration
	expirtime map[string]time.Time
}

func (s *RWMutexMap[T]) Get(key string) (val T, ok bool) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	return s.getValidValue(key)
}

func (s *RWMutexMap[T]) Set(key string, value T, ttl ...time.Duration) {
	s.rw.Lock()
	defer s.rw.Unlock()
	s.items[key] = value

	s.resetExpirtime(key, ttl...)
}

func (s *RWMutexMap[T]) SetByFunc(key string, newValueFunc func(oldValue T) (newValue T), ttl ...time.Duration) (newValue T) {
	s.rw.Lock()
	defer s.rw.Unlock()
	oldValue := s.items[key]
	newValue = newValueFunc(oldValue)
	s.items[key] = newValue

	s.resetExpirtime(key, ttl...)
	return
}

func (s *RWMutexMap[T]) Delete(key string) {
	s.rw.Lock()
	defer s.rw.Unlock()
	delete(s.items, key)

	if s.expirtime != nil {
		delete(s.expirtime, key)
	}
}

func (s *RWMutexMap[T]) resetExpirtime(key string, ttl ...time.Duration) {
	// 过期时间
	if (s.ttl > 0 || len(ttl) > 0) && s.expirtime == nil {
		s.expirtime = make(map[string]time.Time)
	}
	if len(ttl) > 0 {
		s.expirtime[key] = time.Now().Add(ttl[0])
	} else if s.ttl > 0 {
		s.expirtime[key] = time.Now().Add(s.ttl)
	}
}

func (s *RWMutexMap[T]) getValidValue(key string) (_ T, _ bool) {
	val, ok := s.items[key]
	if !ok {
		return
	}
	if s.expirtime != nil {
		if t, tok := s.expirtime[key]; tok && time.Since(t) > 0 {
			return
		}
	}
	return val, true
}

// 清理过期的值
func (s *RWMutexMap[T]) Clear() {
	if s.expirtime == nil {
		return
	}

	s.rw.Lock()
	defer s.rw.Unlock()
	for key, t := range s.expirtime {
		if time.Since(t) > 0 {
			delete(s.items, key)
			delete(s.expirtime, key)
		}
	}
}
