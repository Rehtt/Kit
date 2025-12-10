package util

import (
	"log"
)

type try struct {
	err any
}

// Deprecated: Try uses a non-idiomatic chainable pattern that incurs unnecessary
// heap allocations and obscures control flow.
//
// Please use SafeRun instead, which implements the standard Go defer-recover pattern.
//
// Example replacement:
//
//	util.SafeRun(func() {
//	    // Do logic
//	}, func(err any) {
//	    // Handle panic
//	})
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

// Deprecated: See Try.
func (t *try) Catch(fn func(err any)) *try {
	if t.err != nil && fn != nil {
		fn(t.err)
	}
	return t
}

// Deprecated: See Try.
func (t *try) Finally(fn func()) {
	if fn != nil {
		fn()
	}
}

// SafeRun executes fn safely, catching any panics.
// If a panic occurs, catchHandler is called with the panic value.
// If catchHandler is nil, the panic is logged by default.
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
