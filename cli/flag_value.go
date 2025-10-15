package cli

import (
	"strings"
	"unicode/utf8"
)

// PasswordValue 用于在帮助信息中隐藏密码
type PasswordValue struct {
	value *string
	show  int
}

func (p *PasswordValue) String() string {
	if p.value == nil || *p.value == "" {
		return ""
	}
	l := p.show
	if l == 0 {
		l = utf8.RuneCountInString(*p.value)
	}
	return strings.Repeat("*", l)
}

func (p *PasswordValue) Set(s string) error {
	*p.value = s
	return nil
}

func (p *PasswordValue) Get() any { return *p.value }

// StringsValue 用于接收多个字符串参数
type StringsValue []string

func (s *StringsValue) String() string {
	if s == nil || len(*s) == 0 {
		return ""
	}
	return strings.Join(*s, ",")
}

func (s *StringsValue) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (s *StringsValue) Get() any {
	return []string(*s)
}
