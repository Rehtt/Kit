package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
)

type FlagSet struct {
	*flag.FlagSet
	ShortLongMap map[string]*ShortLongValue // 跟踪短长名关系
	Item         map[*ShortLongValue]FlagItem
}

// Parse 重写 Parse 方法以支持组合的短 flag
func (f *FlagSet) Parse(arguments []string) error {
	// 展开组合的短 flag
	expandedArgs := f.expandCombinedFlags(arguments)
	return f.FlagSet.Parse(expandedArgs)
}

// expandCombinedFlags 将组合的短 flag 展开
// 例如: -abc 展开为 -a -b -c
// 例如: -abf value 展开为 -a -b -f value
func (f *FlagSet) expandCombinedFlags(arguments []string) []string {
	var result []string
	
	for i := 0; i < len(arguments); i++ {
		arg := arguments[i]
		
		// 检查是否是短 flag 组合（单破折号，长度>2，不是负数）
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) > 2 {
			flagName := arg[1:]
			
			// 检查是否是负数（如 -123）
			if isNumeric(flagName) {
				result = append(result, arg)
				continue
			}
			
			// 尝试展开组合的 flag
			expanded := f.tryExpandCombinedFlag(flagName, arguments, i)
			if expanded != nil {
				result = append(result, expanded.flags...)
				// 如果最后一个 flag 消耗了后面的参数，跳过它
				i += expanded.consumedNext
				continue
			}
		}
		
		result = append(result, arg)
	}
	
	return result
}

type expandResult struct {
	flags        []string
	consumedNext int // 消耗了后面多少个参数
}

// tryExpandCombinedFlag 尝试展开组合的短 flag
func (f *FlagSet) tryExpandCombinedFlag(flagName string, arguments []string, currentIndex int) *expandResult {
	flags := []string{}
	
	for i, char := range flagName {
		shortFlag := string(char)
		
		// 检查这个短 flag 是否存在
		flagExists := f.Lookup(shortFlag) != nil
		if !flagExists {
			// 如果 flag 不存在，不展开
			return nil
		}
		
		isLastChar := i == len(flagName)-1
		
		// 检查 flag 类型
		isBoolFlag := f.isBoolFlag(shortFlag)
		
		if !isLastChar && !isBoolFlag {
			// 如果不是最后一个字符，且不是布尔类型，不能组合
			return nil
		}
		
		flags = append(flags, "-"+shortFlag)
		
		// 如果是最后一个字符且不是布尔类型，检查是否需要值
		if isLastChar && !isBoolFlag {
			// 检查下一个参数是否是值
			if currentIndex+1 < len(arguments) {
				nextArg := arguments[currentIndex+1]
				// 如果下一个参数不是 flag，将其作为值
				if !strings.HasPrefix(nextArg, "-") {
					flags = append(flags, nextArg)
					return &expandResult{flags: flags, consumedNext: 1}
				}
			}
		}
	}
	
	return &expandResult{flags: flags, consumedNext: 0}
}

// isBoolFlag 检查给定的 flag 是否是布尔类型
func (f *FlagSet) isBoolFlag(name string) bool {
	flag := f.Lookup(name)
	if flag == nil {
		return false
	}
	
	// 通过检查 DefValue 来判断是否是布尔类型
	return flag.DefValue == "false" || flag.DefValue == "true"
}

// isNumeric 检查字符串是否是数字
func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
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
	if f.ShortLongMap == nil {
		f.ShortLongMap = make(map[string]*ShortLongValue)
	}
}

// ensureItemMap 确保 item 已初始化
func (f *FlagSet) ensureItemMap() {
	if f.Item == nil {
		f.Item = make(map[*ShortLongValue]FlagItem)
	}
}

// addShortLongMapping 添加短长名映射关系
func (f *FlagSet) addShortLongMapping(short, long string) *ShortLongValue {
	f.ensureShortLongMap()

	slValue := &ShortLongValue{ShortName: short, LongName: long}
	if short != "" {
		f.ShortLongMap[short] = slValue
	}
	if long != "" {
		f.ShortLongMap[long] = slValue
	}
	return slValue
}

// registerFlagItem 为单个 flag 注册 ShortLongValue 和 FlagItem
func (f *FlagSet) registerFlagItem(name string, item ...FlagItem) {
	if len(item) > 0 {
		if item[0].Type == FlagItemSelect && len(item[0].Nodes) == 0 {
			return
		}
		f.ensureShortLongMap()
		f.ensureItemMap()

		slValue := &ShortLongValue{LongName: name}
		f.ShortLongMap[name] = slValue
		f.Item[slValue] = item[0]
	}
}

// registerShortLongFlag 注册带短长名的 flag 的通用逻辑
func (f *FlagSet) registerShortLongFlag(short, long string, shortRegister, longRegister func(string), item ...FlagItem) {
	slValue := f.addShortLongMapping(short, long)
	if len(item) > 0 {
		if !(item[0].Type == FlagItemSelect && len(item[0].Nodes) == 0) {
			f.ensureItemMap()
			f.Item[slValue] = item[0]
		}
	}

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
func (f *FlagSet) StringVarShortLong(p *string, short, long string, value string, usage string, item ...FlagItem) {
	f.registerShortLongFlag(short, long,
		func(name string) { f.FlagSet.StringVar(p, name, value, usage) },
		func(name string) { f.FlagSet.StringVar(p, name, value, usage) },
		item...,
	)
}

// StringShortLong 定义并返回一个带短名和长名的 string 类型 flag 指针
func (f *FlagSet) StringShortLong(short, long string, value string, usage string, item ...FlagItem) *string {
	p := new(string)
	f.StringVarShortLong(p, short, long, value, usage, item...)
	return p
}

// IntVarShortLong 定义一个带短名和长名的 int 类型 flag
func (f *FlagSet) IntVarShortLong(p *int, short, long string, value int, usage string, item ...FlagItem) {
	f.registerShortLongFlag(short, long,
		func(name string) { f.FlagSet.IntVar(p, name, value, usage) },
		func(name string) { f.FlagSet.IntVar(p, name, value, usage) },
		item...,
	)
}

// IntShortLong 定义并返回一个带短名和长名的 int 类型 flag 指针
func (f *FlagSet) IntShortLong(short, long string, value int, usage string, item ...FlagItem) *int {
	p := new(int)
	f.IntVarShortLong(p, short, long, value, usage, item...)
	return p
}

// BoolVarShortLong 定义一个带短名和长名的 bool 类型 flag
func (f *FlagSet) BoolVarShortLong(p *bool, short, long string, value bool, usage string, item ...FlagItem) {
	f.registerShortLongFlag(short, long,
		func(name string) { f.FlagSet.BoolVar(p, name, value, usage) },
		func(name string) { f.FlagSet.BoolVar(p, name, value, usage) },
		item...,
	)
}

// BoolShortLong 定义并返回一个带短名和长名的 bool 类型 flag 指针
func (f *FlagSet) BoolShortLong(short, long string, value bool, usage string, item ...FlagItem) *bool {
	p := new(bool)
	f.BoolVarShortLong(p, short, long, value, usage, item...)
	return p
}

// Int64VarShortLong 定义一个带短名和长名的 int64 类型 flag
func (f *FlagSet) Int64VarShortLong(p *int64, short, long string, value int64, usage string, item ...FlagItem) {
	f.registerShortLongFlag(short, long,
		func(name string) { f.FlagSet.Int64Var(p, name, value, usage) },
		func(name string) { f.FlagSet.Int64Var(p, name, value, usage) },
		item...,
	)
}

// Int64ShortLong 定义并返回一个带短名和长名的 int64 类型 flag 指针
func (f *FlagSet) Int64ShortLong(short, long string, value int64, usage string, item ...FlagItem) *int64 {
	p := new(int64)
	f.Int64VarShortLong(p, short, long, value, usage, item...)
	return p
}

// UintVarShortLong 定义一个带短名和长名的 uint 类型 flag
func (f *FlagSet) UintVarShortLong(p *uint, short, long string, value uint, usage string, item ...FlagItem) {
	f.registerShortLongFlag(short, long,
		func(name string) { f.FlagSet.UintVar(p, name, value, usage) },
		func(name string) { f.FlagSet.UintVar(p, name, value, usage) },
		item...,
	)
}

// UintShortLong 定义并返回一个带短名和长名的 uint 类型 flag 指针
func (f *FlagSet) UintShortLong(short, long string, value uint, usage string, item ...FlagItem) *uint {
	p := new(uint)
	f.UintVarShortLong(p, short, long, value, usage, item...)
	return p
}

// Uint64VarShortLong 定义一个带短名和长名的 uint64 类型 flag
func (f *FlagSet) Uint64VarShortLong(p *uint64, short, long string, value uint64, usage string, item ...FlagItem) {
	f.registerShortLongFlag(short, long,
		func(name string) { f.FlagSet.Uint64Var(p, name, value, usage) },
		func(name string) { f.FlagSet.Uint64Var(p, name, value, usage) },
		item...,
	)
}

// Uint64ShortLong 定义并返回一个带短名和长名的 uint64 类型 flag 指针
func (f *FlagSet) Uint64ShortLong(short, long string, value uint64, usage string, item ...FlagItem) *uint64 {
	p := new(uint64)
	f.Uint64VarShortLong(p, short, long, value, usage, item...)
	return p
}

// Float64VarShortLong 定义一个带短名和长名的 float64 类型 flag
func (f *FlagSet) Float64VarShortLong(p *float64, short, long string, value float64, usage string, item ...FlagItem) {
	f.registerShortLongFlag(short, long,
		func(name string) { f.FlagSet.Float64Var(p, name, value, usage) },
		func(name string) { f.FlagSet.Float64Var(p, name, value, usage) },
		item...,
	)
}

// Float64ShortLong 定义并返回一个带短名和长名的 float64 类型 flag 指针
func (f *FlagSet) Float64ShortLong(short, long string, value float64, usage string, item ...FlagItem) *float64 {
	p := new(float64)
	f.Float64VarShortLong(p, short, long, value, usage, item...)
	return p
}

// DurationVarShortLong 定义一个带短名和长名的 time.Duration 类型 flag
func (f *FlagSet) DurationVarShortLong(p *time.Duration, short, long string, value time.Duration, usage string, item ...FlagItem) {
	f.registerShortLongFlag(short, long,
		func(name string) { f.FlagSet.DurationVar(p, name, value, usage) },
		func(name string) { f.FlagSet.DurationVar(p, name, value, usage) },
		item...,
	)
}

// DurationShortLong 定义并返回一个带短名和长名的 time.Duration 类型 flag 指针
func (f *FlagSet) DurationShortLong(short, long string, value time.Duration, usage string, item ...FlagItem) *time.Duration {
	p := new(time.Duration)
	f.DurationVarShortLong(p, short, long, value, usage, item...)
	return p
}

// StringsVarShortLong 定义一个带短名和长名的字符串切片类型 flag
func (f *FlagSet) StringsVarShortLong(p *[]string, short, long string, value []string, usage string, item ...FlagItem) {
	f.registerShortLongFlag(short, long,
		func(name string) {
			*p = value
			stringsValue := (*StringsValue)(p)
			f.Var(stringsValue, name, usage)
		},
		func(name string) {
			*p = value
			stringsValue := (*StringsValue)(p)
			f.Var(stringsValue, name, usage)
		},
		item...,
	)
}

// StringsShortLong 定义并返回一个带短名和长名的字符串切片类型 flag 指针
func (f *FlagSet) StringsShortLong(short, long string, value []string, usage string, item ...FlagItem) *[]string {
	p := new([]string)
	f.StringsVarShortLong(p, short, long, value, usage, item...)
	return p
}

// StringVar 定义一个 string 类型 flag (支持 FlagItem)
func (f *FlagSet) StringVar(p *string, name string, value string, usage string, item ...FlagItem) {
	f.registerFlagItem(name, item...)
	f.FlagSet.StringVar(p, name, value, usage)
}

// String 定义并返回一个 string 类型 flag 指针 (支持 FlagItem)
func (f *FlagSet) String(name string, value string, usage string, item ...FlagItem) *string {
	p := new(string)
	f.StringVar(p, name, value, usage, item...)
	return p
}

// IntVar 定义一个 int 类型 flag (支持 FlagItem)
func (f *FlagSet) IntVar(p *int, name string, value int, usage string, item ...FlagItem) {
	f.registerFlagItem(name, item...)
	f.FlagSet.IntVar(p, name, value, usage)
}

// Int 定义并返回一个 int 类型 flag 指针 (支持 FlagItem)
func (f *FlagSet) Int(name string, value int, usage string, item ...FlagItem) *int {
	p := new(int)
	f.IntVar(p, name, value, usage, item...)
	return p
}

// BoolVar 定义一个 bool 类型 flag (支持 FlagItem)
func (f *FlagSet) BoolVar(p *bool, name string, value bool, usage string, item ...FlagItem) {
	f.registerFlagItem(name, item...)
	f.FlagSet.BoolVar(p, name, value, usage)
}

// Bool 定义并返回一个 bool 类型 flag 指针 (支持 FlagItem)
func (f *FlagSet) Bool(name string, value bool, usage string, item ...FlagItem) *bool {
	p := new(bool)
	f.BoolVar(p, name, value, usage, item...)
	return p
}

// Int64Var 定义一个 int64 类型 flag (支持 FlagItem)
func (f *FlagSet) Int64Var(p *int64, name string, value int64, usage string, item ...FlagItem) {
	f.registerFlagItem(name, item...)
	f.FlagSet.Int64Var(p, name, value, usage)
}

// Int64 定义并返回一个 int64 类型 flag 指针 (支持 FlagItem)
func (f *FlagSet) Int64(name string, value int64, usage string, item ...FlagItem) *int64 {
	p := new(int64)
	f.Int64Var(p, name, value, usage, item...)
	return p
}

// UintVar 定义一个 uint 类型 flag (支持 FlagItem)
func (f *FlagSet) UintVar(p *uint, name string, value uint, usage string, item ...FlagItem) {
	f.registerFlagItem(name, item...)
	f.FlagSet.UintVar(p, name, value, usage)
}

// Uint 定义并返回一个 uint 类型 flag 指针 (支持 FlagItem)
func (f *FlagSet) Uint(name string, value uint, usage string, item ...FlagItem) *uint {
	p := new(uint)
	f.UintVar(p, name, value, usage, item...)
	return p
}

// Uint64Var 定义一个 uint64 类型 flag (支持 FlagItem)
func (f *FlagSet) Uint64Var(p *uint64, name string, value uint64, usage string, item ...FlagItem) {
	f.registerFlagItem(name, item...)
	f.FlagSet.Uint64Var(p, name, value, usage)
}

// Uint64 定义并返回一个 uint64 类型 flag 指针 (支持 FlagItem)
func (f *FlagSet) Uint64(name string, value uint64, usage string, item ...FlagItem) *uint64 {
	p := new(uint64)
	f.Uint64Var(p, name, value, usage, item...)
	return p
}

// Float64Var 定义一个 float64 类型 flag (支持 FlagItem)
func (f *FlagSet) Float64Var(p *float64, name string, value float64, usage string, item ...FlagItem) {
	f.registerFlagItem(name, item...)
	f.FlagSet.Float64Var(p, name, value, usage)
}

// Float64 定义并返回一个 float64 类型 flag 指针 (支持 FlagItem)
func (f *FlagSet) Float64(name string, value float64, usage string, item ...FlagItem) *float64 {
	p := new(float64)
	f.Float64Var(p, name, value, usage, item...)
	return p
}

// DurationVar 定义一个 time.Duration 类型 flag (支持 FlagItem)
func (f *FlagSet) DurationVar(p *time.Duration, name string, value time.Duration, usage string, item ...FlagItem) {
	f.registerFlagItem(name, item...)
	f.FlagSet.DurationVar(p, name, value, usage)
}

// Duration 定义并返回一个 time.Duration 类型 flag 指针 (支持 FlagItem)
func (f *FlagSet) Duration(name string, value time.Duration, usage string, item ...FlagItem) *time.Duration {
	p := new(time.Duration)
	f.DurationVar(p, name, value, usage, item...)
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
// showNum: 0 表示*数量与 value 一致
func (f *FlagSet) PasswordStringVarShortLong(p *string, short, long string, value string, usage string, showNum int, item ...FlagItem) {
	f.registerShortLongFlag(short, long,
		func(name string) { f.PasswordStringVar(p, name, value, usage, showNum) },
		func(name string) { f.PasswordStringVar(p, name, value, usage, showNum) },
		item...,
	)
}

// PasswordStringShortLong 定义并返回一个带短名和长名的密码字符串类型 flag 指针
// showNum: 0 表示*数量与 value 一致
func (f *FlagSet) PasswordStringShortLong(short, long string, value string, usage string, showNum int, item ...FlagItem) *string {
	p := new(string)
	f.PasswordStringVarShortLong(p, short, long, value, usage, showNum, item...)
	return p
}

// StringsVar 定义一个字符串切片类型 flag (支持 FlagItem)
func (f *FlagSet) StringsVar(p *[]string, name string, value []string, usage string, item ...FlagItem) {
	f.registerFlagItem(name, item...)
	*p = value
	stringsValue := (*StringsValue)(p)
	f.Var(stringsValue, name, usage)
}

// Strings 定义并返回一个字符串切片类型 flag 指针 (支持 FlagItem)
func (f *FlagSet) Strings(name string, value []string, usage string, item ...FlagItem) *[]string {
	p := new([]string)
	f.StringsVar(p, name, value, usage, item...)
	return p
}

// PrintDefaults 自定义帮助信息显示，将短长名合并显示
func (f *FlagSet) PrintDefaults() {
	if f.ShortLongMap == nil {
		f.ShortLongMap = make(map[string]*ShortLongValue)
	}
	if f.Item == nil {
		f.Item = make(map[*ShortLongValue]FlagItem)
	}

	processed := make(map[string]bool)

	w := tabwriter.NewWriter(f.Output(), 0, 0, 2, ' ', 0)
	defer w.Flush()
	f.VisitAll(func(flag *flag.Flag) {
		if processed[flag.Name] {
			return
		}

		if slValue, exists := f.ShortLongMap[flag.Name]; exists {
			if slValue.ShortName != "" && slValue.LongName != "" {
				names := "-" + slValue.ShortName + ",\t--" + slValue.LongName
				f.printFlag(w, names, flag)
				processed[slValue.ShortName] = true
				processed[slValue.LongName] = true
			} else if slValue.ShortName != "" {
				f.printFlag(w, "-"+slValue.ShortName+"\t", flag)
				processed[slValue.ShortName] = true
			} else if slValue.LongName != "" {
				f.printFlag(w, "\t--"+slValue.LongName, flag)
				processed[slValue.LongName] = true
			}
		} else {
			f.printFlag(w, "-"+flag.Name+"\t", flag)
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
		// 安全地获取 FlagItem 信息
		if slValue, exists := f.ShortLongMap[flag.Name]; exists {
			if item, hasItem := f.Item[slValue]; hasItem {
				s += " " + item.String()
			} else {
				s += " value"
			}
		} else {
			s += " value"
		}
	}

	if flag.Usage != "" {
		s += fmt.Sprintf("    \t%s", flag.Usage)
		if flag.DefValue != "" && flag.DefValue != "false" {
			s += fmt.Sprintf(" (default %q)", flag.DefValue)
		}
	}
	s += "\n"

	fmt.Fprint(w, s)
}
