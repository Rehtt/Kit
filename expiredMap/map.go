package expiredMap

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Rehtt/Kit/util"
)

var ErrMapClose = errors.New("ExpiredMap is close")

// Deprecated: use github.com/Rehtt/Kit/maps
type ExpiredMap struct {
	lock     sync.RWMutex
	value    map[any]any
	expire   map[any]*time.Time
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

func (e *ExpiredMap) Set(key, value any, ttl ...time.Duration) error {
	if e.CheckClose() {
		return ErrMapClose
	}
	e.lock.Lock()
	e.set(key, value, ttl...)
	e.lock.Unlock()
	return nil
}

func (e *ExpiredMap) set(key, value any, ttl ...time.Duration) {
	e.value[key] = value
	if len(ttl) != 0 {
		e.expire[key] = util.TimeToPrt(time.Now().Add(ttl[0]))
	}
}

func (e *ExpiredMap) Get(key any) (any, bool, error) {
	if e.CheckClose() {
		return nil, false, ErrMapClose
	}
	e.lock.RLock()
	v, ok := e.get(key)
	ttl := e.ttl(key)
	e.lock.RUnlock()
	if ok && ttl < 1 {
		e.Delete(key)
		return nil, false, nil
	}
	return v, ok, nil
}

func (e *ExpiredMap) get(key any) (any, bool) {
	v, ok := e.value[key]
	return v, ok
}

func (e *ExpiredMap) Delete(key any) error {
	if e.CheckClose() {
		return ErrMapClose
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	e.delete(key)
	return nil
}

func (e *ExpiredMap) delete(key any) {
	delete(e.value, key)
	delete(e.expire, key)
}

func (e *ExpiredMap) TTL(key any) (time.Duration, error) {
	if e.CheckClose() {
		return 0, ErrMapClose
	}
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.ttl(key), nil
}

func (e *ExpiredMap) ttl(key any) time.Duration {
	t, ok := e.expire[key]
	if ok {
		return t.Sub(time.Now())
	}
	return -1
}

func (e *ExpiredMap) SetAutoClearInterval(i time.Duration) error {
	if e.CheckClose() {
		return ErrMapClose
	}
	e.interval = i
	return nil
}

func (e *ExpiredMap) Range(f func(key, value any, ttl *time.Duration)) error {
	if e.CheckClose() {
		return ErrMapClose
	}
	e.lock.RLock()
	value := util.DeepCopy(e.value)
	expire := util.DeepCopy(e.expire)
	e.lock.RUnlock()
	for k, v := range value.(map[any]any) {
		now := time.Now()
		exp, ok := expire.(map[any]*time.Time)[k]
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
	return nil
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

func (e *ExpiredMap) Close() error {
	if e.CheckClose() {
		return ErrMapClose
	}
	e.cancel()
	return nil
}

func (e *ExpiredMap) CheckClose() bool {
	select {
	case <-e.ctx.Done():
		return true
	default:
		return false
	}
}
