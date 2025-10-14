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
	f.Var(&passwordValue{value: p, show: num}, name, usage)
}

// PasswordString 定义并返回一个密码字符串类型 flag 指针（使用默认 CommandLine 实例）
// 在帮助信息中，密码值会显示为 ********
// showNum 指定密码显示的字符个数
func (f *FlagSet) PasswordString(name string, value string, usage string, showNum ...int) *string {
	p := new(string)
	f.PasswordStringVar(p, name, value, usage, showNum...)
	return p
}
