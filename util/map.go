package util

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

type MapItem struct {
	m  map[string]any
	mu sync.RWMutex
}

type Maps map[uint8]*MapItem

func NewMap() Maps {
	return make(Maps)
}

func (m Maps) Get(key string, f ...func(any, bool)) (any, bool) {
	k := m.hash(key)
	mm, ok := m[k[0]]
	if !ok {
		for _, ff := range f {
			ff(nil, false)
		}
		return nil, false
	}
	mm.mu.RLock()
	data, ok := mm.m[k]
	for _, ff := range f {
		ff(data, ok)
	}

	mm.mu.RUnlock()
	return data, ok
}

func (m Maps) Set(key string, value any, f ...func()) {
	k := m.hash(key)
	mm := m[k[0]]
	if mm == nil {
		mm = &MapItem{
			m: map[string]any{},
		}
		m[k[0]] = mm
	}
	mm.mu.Lock()
	m.runFunc(f)
	mm.m[k] = value
	mm.mu.Unlock()
}

func (m Maps) Delete(key string, f ...func()) {
	k := m.hash(key)
	mm := m[k[0]]
	if mm == nil {
		return
	}
	mm.mu.Lock()
	m.runFunc(f)
	delete(mm.m, k)
	mm.mu.Unlock()
}

func (m Maps) runFunc(funcs []func()) {
	for _, f := range funcs {
		f()
	}
}

func (m Maps) hash(key string) string {
	s := sha256.New()
	s.Write([]byte(key))
	return hex.EncodeToString(s.Sum(nil))
}
