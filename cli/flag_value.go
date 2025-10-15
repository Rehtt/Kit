package cli

import (
	"flag"
	"strings"
	"unicode/utf8"
)

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

type ShortLongValue struct {
	Value     flag.Value
	ShortName string
	LongName  string
}

func (sl *ShortLongValue) String() string {
	return sl.Value.String()
}

func (sl *ShortLongValue) Set(s string) error {
	return sl.Value.Set(s)
}

func (sl *ShortLongValue) Get() any {
	return sl.Value.(flag.Getter).Get()
}

func (sl *ShortLongValue) GetNames() string {
	if sl.ShortName != "" && sl.LongName != "" {
		return "-" + sl.ShortName + "/--" + sl.LongName
	} else if sl.ShortName != "" {
		return "-" + sl.ShortName
	} else if sl.LongName != "" {
		return "--" + sl.LongName
	}
	return ""
}
