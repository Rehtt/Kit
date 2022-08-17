package client

import "fmt"

const (
	write = iota
	read
)

type middle func(ctx *Context) error
type Middle struct {
	readMiddles  []middle
	writeMiddles []middle
}

func (m *Middle) AddReadMiddleware(f middle) {
	m.readMiddles = append(m.readMiddles, f)
}
func (m *Middle) AddWriteMiddleware(f middle) {
	m.writeMiddles = append(m.writeMiddles, f)
}
func (m *Middle) useMiddleware(ctx *Context, flag int) error {
	var ms *[]middle
	switch flag {
	case write:
		ms = &m.writeMiddles
	case read:
		ms = &m.writeMiddles
	default:
		return fmt.Errorf("状态位错误")
	}
	for _, f := range *ms {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}
