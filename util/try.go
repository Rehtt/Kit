package util

type try struct {
	err interface{}
}

func Try(fn func()) (t *try) {
	t = new(try)
	defer func() {
		if err := recover(); err != nil {
			t.err = err
		}
	}()
	fn()
	return &try{}
}
func (t *try) Catch(fn func(err interface{})) {
	fn(t.err)
}
