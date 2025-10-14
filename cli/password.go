package cli

import "strings"

// passwordValue 用于在帮助信息中隐藏密码
type passwordValue string

func (p *passwordValue) String() string {
	return strings.Repeat("*", len(*p))
}

func (p *passwordValue) Set(s string) error {
	*p = passwordValue(s)
	return nil
}

// PasswordStringVar 定义一个密码字符串类型 flag（使用默认 CommandLine 实例）
// 在帮助信息中，密码值会显示为 ********
func PasswordStringVar(p *string, name string, value string, usage string) {
	*p = value
	pass := passwordValue(*p)
	CommandLine.Var(&pass, name, usage)
}

// PasswordString 定义并返回一个密码字符串类型 flag 指针（使用默认 CommandLine 实例）
// 在帮助信息中，密码值会显示为 ********
func PasswordString(name string, value string, usage string) *string {
	p := new(string)
	PasswordStringVar(p, name, value, usage)
	return p
}
