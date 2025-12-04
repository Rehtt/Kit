package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/Rehtt/Kit/util"
)

var CommandLine *CLI

type (
	CommandFunc func(args []string) error
	CLI         struct {
		// Use 是命令的名称
		Use string
		// Instruction 是命令被调用时显示的说明信息
		Instruction string
		// Usage 是在没有参数时显示的用法信息
		Usage string
		// CommandFunc 是该命令被调用时执行的函数
		CommandFunc CommandFunc
		*FlagSet
		SubCommands *SubCommands
		// Hidden 表示该命令在帮助中不显示
		Hidden bool
		// Raw 表示该命令不被解析
		Raw bool
	}
)

func init() {
	if len(os.Args) > 0 {
		CommandLine = NewCLI(os.Args[0], "")
	} else {
		CommandLine = NewCLI("", "")
	}
}

func NewCLI(use, instruction string) *CLI {
	cli := &CLI{
		Use:         use,
		Instruction: instruction,
		FlagSet:     &FlagSet{FlagSet: flag.NewFlagSet(use, flag.ContinueOnError)},
		SubCommands: &SubCommands{},
	}
	return cli
}

func (c *CLI) AddCommand(cli ...*CLI) error {
	return c.SubCommands.Add(cli...)
}

func (c *CLI) Help() {
	w := tabwriter.NewWriter(c.Output(), 0, 0, 2, ' ', 0)
	defer w.Flush()
	if c.Instruction != "" {
		fmt.Fprintf(w, "%s\n\n", c.Instruction)
	}
	if c.Usage == "" {
		c.Usage = "[flags]"
		if c.SubCommands.Len() > 0 {
			c.Usage += " [command]"
		}
	}
	fmt.Fprintln(w, "Usage: "+c.Use+" "+c.Usage)

	fmt.Fprintln(w, "\nFlags:")
	c.PrintDefaults()
	if c.SubCommands.Len() > 0 {
		fmt.Fprintln(w, "\nAvailable Commands:")
		subs := c.SubCommands.CloneList()
		switch c.SubCommands.GetSort() {
		case CommandSortAlphaAsc:
			slices.SortFunc(subs, func(a, b *CLI) int { return strings.Compare(a.Use, b.Use) })
		case CommandSortAlphaDesc:
			slices.SortFunc(subs, func(a, b *CLI) int { return strings.Compare(b.Use, a.Use) })
		}
		for _, v := range subs {
			if !v.Hidden {
				fmt.Fprintf(w, "  %s\t%s\n", v.Use, v.Instruction)
			}
		}
		fmt.Fprintf(w, "\nUse '%s [command] -h, --help' for more information about a command.\n", c.Use)
	}
}

func (c *CLI) OutputErrHelp(err error) {
	fmt.Fprintln(c.Output(), err)
	c.Help()
}

func (c *CLI) Parse(arguments []string) error {
	c.FlagSet.Usage = c.Help

	if c.Raw {
		if c.CommandFunc == nil {
			return errors.New("raw command must have a command function")
		}
		return c.CommandFunc(arguments)
	}

	if err := c.FlagSet.Parse(arguments); err != nil {
		if e := util.UnwrapError[cliFlagError](err); e != err {
			c.OutputErrHelp(e)
		}
		return err
	}
	if c.SubCommands.Len() > 0 && c.NArg() > 0 {
		cmdName := c.Arg(0)
		sub := c.SubCommands.Get(cmdName)
		if sub == nil {
			err := fmt.Errorf("unknown subcommand %q: %w", cmdName, flag.ErrHelp)
			c.OutputErrHelp(err)
			return err
		}
		return sub.Parse(c.Args()[1:])
	}
	if c.CommandFunc == nil {
		err := fmt.Errorf("no command: %w", flag.ErrHelp)
		c.OutputErrHelp(err)
		return err
	}
	if err := c.CommandFunc(c.Args()); err != nil && err != flag.ErrHelp {
		return err
	}
	return nil
}

func (c *CLI) Run(arguments []string) error {
	return c.Parse(arguments)
}

// IsCompleteFlag 检查给定的字符串是否是一个完整的参数名
func (c *CLI) IsCompleteFlag(arg string) bool {
	if !strings.HasPrefix(arg, "-") {
		return false
	}

	// 检查是否是已定义的参数
	flagName := strings.TrimPrefix(strings.TrimPrefix(arg, "--"), "-")

	// 对于长参数前缀（如 --t），不认为是完整参数
	if strings.HasPrefix(arg, "--") && len(flagName) == 1 {
		return false // --t 这样的应该被当作参数名前缀，不是完整参数
	}

	// 对于短参数，只有单个字符才认为是完整的
	if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") {
		if len(flagName) != 1 {
			return false // 多字符的短参数前缀，不是完整参数
		}
	}

	// 检查短长名映射
	if c.FlagSet.ShortLongMap != nil {
		if _, exists := c.FlagSet.ShortLongMap[flagName]; exists {
			return true
		}
	}

	// 检查是否是标准参数
	found := false
	c.FlagSet.VisitAll(func(f *flag.Flag) {
		if f.Name == flagName {
			found = true
		}
	})

	return found
}

// IsCompleteFlagInContext 在给定上下文中检查参数是否完整
func (c *CLI) IsCompleteFlagInContext(arg string, args []string) bool {
	// 如果有子命令，需要在子命令的上下文中检查
	if len(args) > 0 && c.SubCommands != nil {
		if sub := c.SubCommands.Get(args[0]); sub != nil {
			// 在子命令上下文中检查
			return sub.IsCompleteFlag(arg)
		}
	}

	// 在当前命令上下文中检查
	return c.IsCompleteFlag(arg)
}

func AddCommand(cli ...*CLI) error { return CommandLine.AddCommand(cli...) }

func Parse() error { return CommandLine.Parse(os.Args[1:]) }

func Run() error { return CommandLine.Run(os.Args[1:]) }

func Parsed() bool { return CommandLine.Parsed() }

func Args() []string { return CommandLine.Args() }

func NArg() int { return CommandLine.NArg() }
