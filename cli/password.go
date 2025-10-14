package cli

import (
	"strings"
	"unicode/utf8"
)

// passwordValue 用于在帮助信息中隐藏密码
type passwordValue struct {
	value *string
	show  int
}

func (p *passwordValue) String() string {
	if p.value == nil || *p.value == "" {
		return ""
	}
	l := p.show
	if l == 0 {
		l = utf8.RuneCountInString(*p.value)
	}
	return strings.Repeat("*", l)
}

func (p *passwordValue) Set(s string) error {
	*p.value = s
	return nil
}
