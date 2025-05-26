package util

import "sync"

type try struct {
	err any
}

var tryPool = sync.Pool{
	New: func() any { return new(try) },
}

func Try(fn func()) (t *try) {
	t = tryPool.Get().(*try)
	t.err = nil
	if fn == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			t.err = err
		}
	}()
	fn()
	return &try{}
}

func (t *try) Catch(fn func(err any)) {
	if fn != nil {
		fn(t.err)
	}
}

func (t *try) Finally(fn func()) {
	if fn != nil {
		fn()
	}
	tryPool.Put(t)
}
