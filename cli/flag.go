package cli

import (
	"flag"
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

type FlagSet struct {
	*flag.FlagSet
	shortLongMap map[string]*ShortLongValue // 跟踪短长名关系
}

// Alias 为已存在的 flag 添加别名
func (f *FlagSet) Alias(alias, original string) {
	originalFlag := f.Lookup(original)
	if originalFlag == nil {
		panic("flag " + original + " does not exist")
	}
	f.Var(originalFlag.Value, alias, originalFlag.Usage)
}

// ensureShortLongMap 确保 shortLongMap 已初始化
func (f *FlagSet) ensureShortLongMap() {
	if f.shortLongMap == nil {
		f.shortLongMap = make(map[string]*ShortLongValue)
	}
}

// addShortLongMapping 添加短长名映射关系
func (f *FlagSet) addShortLongMapping(short, long string) {
	f.ensureShortLongMap()

	slValue := &ShortLongValue{ShortName: short, LongName: long}
	if short != "" {
		f.shortLongMap[short] = slValue
	}
	if long != "" {
		f.shortLongMap[long] = slValue
	}
}

// registerShortLongFlag 注册带短长名的 flag 的通用逻辑
func (f *FlagSet) registerShortLongFlag(short, long, usage string, shortRegister, longRegister func(string)) {
	f.addShortLongMapping(short, long)

	if short != "" {
		shortRegister(short)
	}

	if long != "" {
		if short != "" {
			f.Alias(long, short)
		} else {
			longRegister(long)
		}
	}
}

// StringVarShortLong 定义一个带短名和长名的 string 类型 flag
func (f *FlagSet) StringVarShortLong(p *string, short, long string, value string, usage string) {
	f.registerShortLongFlag(short, long, usage,
		func(name string) { f.StringVar(p, name, value, usage) },
		func(name string) { f.StringVar(p, name, value, usage) },
	)
}

// StringShortLong 定义并返回一个带短名和长名的 string 类型 flag 指针
func (f *FlagSet) StringShortLong(short, long string, value string, usage string) *string {
	p := new(string)
	f.StringVarShortLong(p, short, long, value, usage)
	return p
}

// IntVarShortLong 定义一个带短名和长名的 int 类型 flag
func (f *FlagSet) IntVarShortLong(p *int, short, long string, value int, usage string) {
	f.registerShortLongFlag(short, long, usage,
		func(name string) { f.IntVar(p, name, value, usage) },
		func(name string) { f.IntVar(p, name, value, usage) },
	)
}

// IntShortLong 定义并返回一个带短名和长名的 int 类型 flag 指针
func (f *FlagSet) IntShortLong(short, long string, value int, usage string) *int {
	p := new(int)
	f.IntVarShortLong(p, short, long, value, usage)
	return p
}

// BoolVarShortLong 定义一个带短名和长名的 bool 类型 flag
func (f *FlagSet) BoolVarShortLong(p *bool, short, long string, value bool, usage string) {
	f.registerShortLongFlag(short, long, usage,
		func(name string) { f.BoolVar(p, name, value, usage) },
		func(name string) { f.BoolVar(p, name, value, usage) },
	)
}

// BoolShortLong 定义并返回一个带短名和长名的 bool 类型 flag 指针
func (f *FlagSet) BoolShortLong(short, long string, value bool, usage string) *bool {
	p := new(bool)
	f.BoolVarShortLong(p, short, long, value, usage)
	return p
}

// Int64VarShortLong 定义一个带短名和长名的 int64 类型 flag
func (f *FlagSet) Int64VarShortLong(p *int64, short, long string, value int64, usage string) {
	f.registerShortLongFlag(short, long, usage,
		func(name string) { f.Int64Var(p, name, value, usage) },
		func(name string) { f.Int64Var(p, name, value, usage) },
	)
}

// Int64ShortLong 定义并返回一个带短名和长名的 int64 类型 flag 指针
func (f *FlagSet) Int64ShortLong(short, long string, value int64, usage string) *int64 {
	p := new(int64)
	f.Int64VarShortLong(p, short, long, value, usage)
	return p
}

// UintVarShortLong 定义一个带短名和长名的 uint 类型 flag
func (f *FlagSet) UintVarShortLong(p *uint, short, long string, value uint, usage string) {
	f.registerShortLongFlag(short, long, usage,
		func(name string) { f.UintVar(p, name, value, usage) },
		func(name string) { f.UintVar(p, name, value, usage) },
	)
}

// UintShortLong 定义并返回一个带短名和长名的 uint 类型 flag 指针
func (f *FlagSet) UintShortLong(short, long string, value uint, usage string) *uint {
	p := new(uint)
	f.UintVarShortLong(p, short, long, value, usage)
	return p
}

// Uint64VarShortLong 定义一个带短名和长名的 uint64 类型 flag
func (f *FlagSet) Uint64VarShortLong(p *uint64, short, long string, value uint64, usage string) {
	f.registerShortLongFlag(short, long, usage,
		func(name string) { f.Uint64Var(p, name, value, usage) },
		func(name string) { f.Uint64Var(p, name, value, usage) },
	)
}

// Uint64ShortLong 定义并返回一个带短名和长名的 uint64 类型 flag 指针
func (f *FlagSet) Uint64ShortLong(short, long string, value uint64, usage string) *uint64 {
	p := new(uint64)
	f.Uint64VarShortLong(p, short, long, value, usage)
	return p
}

// Float64VarShortLong 定义一个带短名和长名的 float64 类型 flag
func (f *FlagSet) Float64VarShortLong(p *float64, short, long string, value float64, usage string) {
	f.registerShortLongFlag(short, long, usage,
		func(name string) { f.Float64Var(p, name, value, usage) },
		func(name string) { f.Float64Var(p, name, value, usage) },
	)
}

// Float64ShortLong 定义并返回一个带短名和长名的 float64 类型 flag 指针
func (f *FlagSet) Float64ShortLong(short, long string, value float64, usage string) *float64 {
	p := new(float64)
	f.Float64VarShortLong(p, short, long, value, usage)
	return p
}

// DurationVarShortLong 定义一个带短名和长名的 time.Duration 类型 flag
func (f *FlagSet) DurationVarShortLong(p *time.Duration, short, long string, value time.Duration, usage string) {
	f.registerShortLongFlag(short, long, usage,
		func(name string) { f.DurationVar(p, name, value, usage) },
		func(name string) { f.DurationVar(p, name, value, usage) },
	)
}

// DurationShortLong 定义并返回一个带短名和长名的 time.Duration 类型 flag 指针
func (f *FlagSet) DurationShortLong(short, long string, value time.Duration, usage string) *time.Duration {
	p := new(time.Duration)
	f.DurationVarShortLong(p, short, long, value, usage)
	return p
}

// StringsVarShortLong 定义一个带短名和长名的字符串切片类型 flag
func (f *FlagSet) StringsVarShortLong(p *[]string, short, long string, value []string, usage string) {
	f.registerShortLongFlag(short, long, usage,
		func(name string) { f.StringsVar(p, name, value, usage) },
		func(name string) { f.StringsVar(p, name, value, usage) },
	)
}

// StringsShortLong 定义并返回一个带短名和长名的字符串切片类型 flag 指针
func (f *FlagSet) StringsShortLong(short, long string, value []string, usage string) *[]string {
	p := new([]string)
	f.StringsVarShortLong(p, short, long, value, usage)
	return p
}

// PasswordStringVar 定义一个密码字符串类型 flag，在帮助信息中密码值会被隐藏
func (f *FlagSet) PasswordStringVar(p *string, name string, value string, usage string, showNum ...int) {
	*p = value
	var num int
	if len(showNum) > 0 {
		num = showNum[0]
	}
	f.Var(&PasswordValue{value: p, show: num}, name, usage)
}

// PasswordString 定义并返回一个密码字符串类型 flag 指针
func (f *FlagSet) PasswordString(name string, value string, usage string, showNum ...int) *string {
	p := new(string)
	f.PasswordStringVar(p, name, value, usage, showNum...)
	return p
}

// PasswordStringVarShortLong 定义一个带短名和长名的密码字符串类型 flag
func (f *FlagSet) PasswordStringVarShortLong(p *string, short, long string, value string, usage string, showNum ...int) {
	f.registerShortLongFlag(short, long, usage,
		func(name string) { f.PasswordStringVar(p, name, value, usage, showNum...) },
		func(name string) { f.PasswordStringVar(p, name, value, usage, showNum...) },
	)
}

// PasswordStringShortLong 定义并返回一个带短名和长名的密码字符串类型 flag 指针
func (f *FlagSet) PasswordStringShortLong(short, long string, value string, usage string, showNum ...int) *string {
	p := new(string)
	f.PasswordStringVarShortLong(p, short, long, value, usage, showNum...)
	return p
}

// StringsVar 定义一个字符串切片类型 flag
func (f *FlagSet) StringsVar(p *[]string, name string, value []string, usage string) {
	*p = value
	stringsValue := (*StringsValue)(p)
	f.Var(stringsValue, name, usage)
}

// Strings 定义并返回一个字符串切片类型 flag 指针
func (f *FlagSet) Strings(name string, value []string, usage string) *[]string {
	p := new([]string)
	f.StringsVar(p, name, value, usage)
	return p
}

// PrintDefaults 自定义帮助信息显示，将短长名合并显示
func (f *FlagSet) PrintDefaults() {
	if f.shortLongMap == nil {
		f.FlagSet.PrintDefaults()
		return
	}

	processed := make(map[string]bool)

	w := tabwriter.NewWriter(f.Output(), 0, 0, 2, ' ', 0)
	defer w.Flush()
	f.VisitAll(func(flag *flag.Flag) {
		if processed[flag.Name] {
			return
		}

		if slValue, exists := f.shortLongMap[flag.Name]; exists {
			if slValue.ShortName != "" && slValue.LongName != "" {
				names := "-" + slValue.ShortName + "/--" + slValue.LongName
				f.printFlag(w, names, flag)
				processed[slValue.ShortName] = true
				processed[slValue.LongName] = true
			} else if slValue.ShortName != "" {
				f.printFlag(w, "-"+slValue.ShortName, flag)
				processed[slValue.ShortName] = true
			} else if slValue.LongName != "" {
				f.printFlag(w, "--"+slValue.LongName, flag)
				processed[slValue.LongName] = true
			}
		} else {
			f.printFlag(w, "-"+flag.Name, flag)
			processed[flag.Name] = true
		}
	})
}

func (f *FlagSet) printFlag(w io.Writer, name string, flag *flag.Flag) {
	s := fmt.Sprintf("  %s", name)

	// 更准确地判断是否需要显示 value
	// 布尔类型的 flag 不需要显示 value，其他类型需要
	isBoolFlag := flag.DefValue == "false" || flag.DefValue == "true"
	if !isBoolFlag {
		s += " value"
	}

	s += "\n"
	if flag.Usage != "" {
		s += fmt.Sprintf("    \t%s", flag.Usage)
		if flag.DefValue != "" && flag.DefValue != "false" {
			s += fmt.Sprintf(" (default %q)", flag.DefValue)
		}
	}
	s += "\n"

	fmt.Fprint(w, s)
}
