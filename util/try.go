package util

import "log"

type try struct {
	err any
}

func Try(fn func()) *try {
	t := &try{}

	if fn == nil {
		return t
	}

	defer func() {
		if err := recover(); err != nil {
			t.err = err
		}
	}()

	fn()
	return t
}

func (t *try) Catch(fn func(err any)) *try {
	if t.err != nil && fn != nil {
		fn(t.err)
	}
	return t
}

func (t *try) Finally(fn func()) {
	if fn != nil {
		fn()
	}
}

func SafeRun(fn func(), catchHandler func(err any)) {
	defer func() {
		if r := recover(); r != nil {
			if catchHandler != nil {
				catchHandler(r)
			} else {
				log.Printf("Recovered from panic: %v", r)
			}
		}
	}()
	fn()
}
