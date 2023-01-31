package expiredMap

import (
	"context"
	"github.com/Rehtt/Kit/util"
	"sync"
	"time"
)

type ExpiredMap struct {
	lock     sync.RWMutex
	value    map[interface{}]interface{}
	expire   map[interface{}]*time.Time
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
}

func New() *ExpiredMap {
	e := new(ExpiredMap)
	e.interval = 30 * time.Second // 间隔30分钟主动删除
	e.ctx, e.cancel = context.WithCancel(context.Background())
	go e.run()
	return e
}
func (e *ExpiredMap) Set(key, value interface{}, ttl ...time.Duration) {
	e.lock.Lock()
	e.set(key, value, ttl...)
	e.lock.Unlock()
}
func (e *ExpiredMap) set(key, value interface{}, ttl ...time.Duration) {
	e.value[key] = value
	if len(ttl) != 0 {
		e.expire[key] = util.TimeToPrt(time.Now().Add(ttl[0]))
	}
}

func (e *ExpiredMap) Get(key interface{}) (interface{}, bool) {
	e.lock.RLock()
	v, ok := e.get(key)
	ttl := e.ttl(key)
	e.lock.RUnlock()
	if ok && ttl < 1 {
		e.Delete(key)
		return nil, false
	}
	return v, ok
}
func (e *ExpiredMap) get(key interface{}) (interface{}, bool) {
	v, ok := e.value[key]
	return v, ok
}
func (e *ExpiredMap) Delete(key interface{}) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.delete(key)
}
func (e *ExpiredMap) delete(key interface{}) {
	delete(e.value, key)
	delete(e.expire, key)
}

func (e *ExpiredMap) TTL(key interface{}) time.Duration {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.ttl(key)
}
func (e *ExpiredMap) ttl(key interface{}) time.Duration {
	t, ok := e.expire[key]
	if ok {
		return t.Sub(time.Now())
	}
	return -1
}
func (e *ExpiredMap) SetAutoClearInterval(i time.Duration) {
	e.interval = i
}

func (e *ExpiredMap) Range(f func(key, value interface{}, ttl *time.Duration)) {
	e.lock.RLock()
	value := util.DeepCopy(e.value)
	expire := util.DeepCopy(e.expire)
	e.lock.RUnlock()
	for k, v := range value.(map[interface{}]interface{}) {
		now := time.Now()
		exp, ok := expire.(map[interface{}]*time.Time)[k]
		var t time.Duration
		if ok {
			if !now.Before(*exp) {
				e.Delete(k)
				continue
			}
			t = exp.Sub(now)
		}
		f(k, v, &t)
	}
}
func (e *ExpiredMap) run() {
	t := time.NewTimer(e.interval)
	for {
		select {
		case <-t.C:
			now := time.Now()
			e.lock.Lock()
			for key, exp := range e.expire {
				if !now.Before(*exp) {
					e.delete(key)
				}
			}
			e.lock.Unlock()
			t.Reset(e.interval)
		case <-e.ctx.Done():
			return
		}
	}
}
