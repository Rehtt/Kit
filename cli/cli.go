package cli

import (
	"encoding"
	"flag"
	"fmt"
	"os"
	"time"
)

var CommandLine *CLI

type (
	CommandFunc func(args []string) error
	CLI         struct {
		Use         string
		Instruction string
		CommandFunc CommandFunc
		*flag.FlagSet
		SubCommands map[string]*CLI
	}
)

func init() {
	if len(flag.Args()) > 0 {
		CommandLine = NewCLI(os.Args[0], "", flag.ContinueOnError)
	} else {
		CommandLine = NewCLI("", "", flag.ExitOnError)
	}
}

func NewCLI(use, instruction string, errorHandling flag.ErrorHandling) *CLI {
	return &CLI{
		Use:         use,
		Instruction: instruction,
		FlagSet:     flag.NewFlagSet(use, errorHandling),
	}
}

func (c *CLI) AddCommand(cli ...*CLI) error {
	if c.SubCommands == nil {
		c.SubCommands = make(map[string]*CLI, len(cli))
	}
	for _, v := range cli {
		if _, ok := c.SubCommands[v.Use]; ok {
			return fmt.Errorf("duplicate command: %s", v.Use)
		}
		c.SubCommands[v.Use] = v
	}
	return nil
}

func (c *CLI) Help() {
	if c.Instruction != "" {
		fmt.Fprintf(c.Output(), "%s\n\n", c.Instruction)
	}
	fmt.Fprintf(c.Output(), "Usage of %s:\n", c.Name()) // <-- 修复点 1
	c.PrintDefaults()
	if len(c.SubCommands) > 0 {
		fmt.Fprintf(c.Output(), "\nSubcommands:\n")
		for _, v := range c.SubCommands {
			fmt.Fprintf(c.Output(), "  %s\t%s\n", v.Use, v.Instruction)
		}
	}
}

func (c *CLI) Parse(arguments []string) error {
	c.FlagSet.Usage = c.Help
	if err := c.FlagSet.Parse(arguments); err != nil {
		return err
	}
	if len(c.SubCommands) > 0 && c.NArg() > 0 {
		cmdName := c.Arg(0)
		if sub, ok := c.SubCommands[cmdName]; ok {
			return sub.Parse(c.Args()[1:])
		}
		c.Help()
		return nil
	}
	if c.CommandFunc == nil {
		c.Help()
		return nil
	}
	return c.CommandFunc(c.Args())
}

// Parse 别名
func (c *CLI) Run(arguments []string) error {
	return c.Parse(arguments)
}

func AddCommand(cli ...*CLI) error { return CommandLine.AddCommand(cli...) }

// Parse 解析命令行参数（使用默认 CommandLine 实例）
func Parse() error { return CommandLine.Parse(os.Args[1:]) }

// Parsed 判断命令行参数是否已被解析
func Parsed() bool { return CommandLine.Parsed() }

// Args 返回非 flag 参数（使用默认 CommandLine 实例）
func Args() []string { return CommandLine.Args() }

// NArg 返回非 flag 参数的数量（使用默认 CommandLine 实例）
func NArg() int { return CommandLine.NArg() }

// BoolVar 定义一个 bool 类型 flag（使用默认 CommandLine 实例）
func BoolVar(p *bool, name string, value bool, usage string) {
	CommandLine.BoolVar(p, name, value, usage)
}

// Bool 定义并返回一个 bool 类型 flag 指针（使用默认 CommandLine 实例）
func Bool(name string, value bool, usage string) *bool { return CommandLine.Bool(name, value, usage) }

// StringVar 定义一个 string 类型 flag（使用默认 CommandLine 实例）
func StringVar(p *string, name string, value string, usage string) {
	CommandLine.StringVar(p, name, value, usage)
}

// String 定义并返回一个 string 类型 flag 指针（使用默认 CommandLine 实例）
func String(name string, value string, usage string) *string {
	return CommandLine.String(name, value, usage)
}

// IntVar 定义一个 int 类型 flag（使用默认 CommandLine 实例）
func IntVar(p *int, name string, value int, usage string) { CommandLine.IntVar(p, name, value, usage) }

// Int 定义并返回一个 int 类型 flag 指针（使用默认 CommandLine 实例）
func Int(name string, value int, usage string) *int { return CommandLine.Int(name, value, usage) }

// Int64Var 定义一个 int64 类型 flag（使用默认 CommandLine 实例）
func Int64Var(p *int64, name string, value int64, usage string) {
	CommandLine.Int64Var(p, name, value, usage)
}

// Int64 定义并返回一个 int64 类型 flag 指针（使用默认 CommandLine 实例）
func Int64(name string, value int64, usage string) *int64 {
	return CommandLine.Int64(name, value, usage)
}

// UintVar 定义一个 uint 类型 flag（使用默认 CommandLine 实例）
func UintVar(p *uint, name string, value uint, usage string) {
	CommandLine.UintVar(p, name, value, usage)
}

// Uint 定义并返回一个 uint 类型 flag 指针（使用默认 CommandLine 实例）
func Uint(name string, value uint, usage string) *uint { return CommandLine.Uint(name, value, usage) }

// Uint64Var 定义一个 uint64 类型 flag（使用默认 CommandLine 实例）
func Uint64Var(p *uint64, name string, value uint64, usage string) {
	CommandLine.Uint64Var(p, name, value, usage)
}

// Uint64 定义并返回一个 uint64 类型 flag 指针（使用默认 CommandLine 实例）
func Uint64(name string, value uint64, usage string) *uint64 {
	return CommandLine.Uint64(name, value, usage)
}

// Float64Var 定义一个 float64 类型 flag（使用默认 CommandLine 实例）
func Float64Var(p *float64, name string, value float64, usage string) {
	CommandLine.Float64Var(p, name, value, usage)
}

// Float64 定义并返回一个 float64 类型 flag 指针（使用默认 CommandLine 实例）
func Float64(name string, value float64, usage string) *float64 {
	return CommandLine.Float64(name, value, usage)
}

// DurationVar 定义一个 time.Duration 类型 flag（使用默认 CommandLine 实例）
func DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	CommandLine.DurationVar(p, name, value, usage)
}

// Duration 定义并返回一个 time.Duration 类型 flag 指针（使用默认 CommandLine 实例）
func Duration(name string, value time.Duration, usage string) *time.Duration {
	return CommandLine.Duration(name, value, usage)
}

// TextVar 定义一个实现 encoding.TextUnmarshaler 的 flag（使用默认 CommandLine 实例）
func TextVar(p encoding.TextUnmarshaler, name string, value encoding.TextMarshaler, usage string) {
	CommandLine.TextVar(p, name, value, usage)
}

// Func 定义一个自定义处理函数的 flag（使用默认 CommandLine 实例）
func Func(name, usage string, fn func(string) error) { CommandLine.Func(name, usage, fn) }

// BoolFunc 定义一个自定义处理函数的 bool flag（使用默认 CommandLine 实例）
func BoolFunc(name, usage string, fn func(string) error) { CommandLine.BoolFunc(name, usage, fn) }

// Var 定义一个自定义 flag.Value 的 flag（使用默认 CommandLine 实例）
func Var(p flag.Value, name string, usage string) { CommandLine.Var(p, name, usage) }

// VisitAll 按字典序遍历所有 flag，调用 fn（包括未设置的 flag）
func VisitAll(fn func(*flag.Flag)) {
	CommandLine.VisitAll(fn)
}

// Visit 按字典序遍历已设置的 flag，调用 fn
func Visit(fn func(*flag.Flag)) {
	CommandLine.Visit(fn)
}

// Lookup 返回指定名称的 flag 指针，若不存在则返回 nil
func Lookup(name string) *flag.Flag {
	return CommandLine.Lookup(name)
}

// Set 设置指定名称 flag 的值
func Set(name, value string) error {
	return CommandLine.Set(name, value)
}
