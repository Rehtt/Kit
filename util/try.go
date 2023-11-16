package util

type try struct {
	err any
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
func (t *try) Catch(fn func(err any)) {
	fn(t.err)
}
