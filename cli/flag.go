package cli

import "flag"

type FlagSet struct {
	*flag.FlagSet
}

// PasswordStringVar 定义一个密码字符串类型 flag（使用默认 CommandLine 实例）
// 在帮助信息中，密码值会显示为 ********
// showNum 指定密码显示的字符个数
func (f *FlagSet) PasswordStringVar(p *string, name string, value string, usage string, showNum ...int) {
	*p = value
	var num int
	if len(showNum) > 0 {
		num = showNum[0]
	}
	f.Var(&PasswordValue{value: p, show: num}, name, usage)
}

// PasswordString 定义并返回一个密码字符串类型 flag 指针（使用默认 CommandLine 实例）
// 在帮助信息中，密码值会显示为 ********
// showNum 指定密码显示的字符个数
func (f *FlagSet) PasswordString(name string, value string, usage string, showNum ...int) *string {
	p := new(string)
	f.PasswordStringVar(p, name, value, usage, showNum...)
	return p
}

// StringsVar 定义一个字符串切片类型 flag，可以接收多个值
// 例如: -s value1 -s value2 -s value3
func (f *FlagSet) StringsVar(p *[]string, name string, value []string, usage string) {
	*p = value
	stringsValue := (*StringsValue)(p)
	f.Var(stringsValue, name, usage)
}

// Strings 定义并返回一个字符串切片类型 flag 指针，可以接收多个值
// 例如: -s value1 -s value2 -s value3
func (f *FlagSet) Strings(name string, value []string, usage string) *[]string {
	p := new([]string)
	f.StringsVar(p, name, value, usage)
	return p
}
