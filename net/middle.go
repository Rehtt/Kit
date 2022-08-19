package gonet

import "fmt"

const (
	read middleModel = iota
	write
)

type middleModel uint8

type MiddleInterface interface {
	// BeforeReading HandleFunc前调用
	BeforeReading(ctx *Context) error

	// BeforeSend 发送前调用
	BeforeSend(ctx *Context) error
}

type middle struct {
	middles []MiddleInterface
}

func (m *middle) Add(middle MiddleInterface) {
	m.middles = append(m.middles, middle)
}

func (m *middle) use(ctx *Context, flag middleModel) (err error) {
	for i := range m.middles {
		switch flag {
		case read:
			err = m.middles[i].BeforeReading(ctx)
		case write:
			err = m.middles[i].BeforeSend(ctx)
		default:
			return fmt.Errorf("unknown flag: %d\n", flag)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
