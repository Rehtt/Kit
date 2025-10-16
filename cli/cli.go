package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

var CommandLine *CLI

type (
	CommandFunc func(args []string) error
	CLI         struct {
		Use         string
		Instruction string
		Usage       string
		CommandFunc CommandFunc
		*FlagSet
		SubCommands map[string]*CLI
		Hidden      bool
		Raw         bool
	}
)

func init() {
	if len(flag.Args()) > 0 {
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
	}
	return cli
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
	w := tabwriter.NewWriter(c.Output(), 0, 0, 2, ' ', 0)
	defer w.Flush()
	if c.Instruction != "" {
		fmt.Fprintf(w, "%s\n\n", c.Instruction)
	}
	fmt.Fprintln(w, "Usage: "+c.Use+" "+c.Usage)
	c.PrintDefaults()
	if len(c.SubCommands) > 0 {
		var subCommands []string
		for _, v := range c.SubCommands {
			if !v.Hidden {
				subCommands = append(subCommands, v.Use)
			}
		}
		sort.Strings(subCommands)

		if len(subCommands) > 0 {
			fmt.Fprintln(w, "\nSubcommands:")
			for _, use := range subCommands {
				sub := c.SubCommands[use]
				fmt.Fprintf(w, "  %s\t%s\n", sub.Use, sub.Instruction)
			}
		}
	}
}

func (c *CLI) Parse(arguments []string) error {
	c.FlagSet.Usage = c.Help

	if c.Raw {
		if c.CommandFunc == nil {
			return errors.New("Raw command must have a command function")
		}
		return c.CommandFunc(arguments)
	}

	if err := c.FlagSet.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if len(c.SubCommands) > 0 && c.NArg() > 0 {
		cmdName := c.Arg(0)
		if sub, ok := c.SubCommands[cmdName]; ok {
			return sub.Parse(c.Args()[1:])
		}
		c.Help()
		return fmt.Errorf("unknown subcommand %q: %w", cmdName, flag.ErrHelp)
	}
	if c.CommandFunc == nil {
		c.Help()
		return fmt.Errorf("no command: %w", flag.ErrHelp)
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
		if subCmd, exists := c.SubCommands[args[0]]; exists {
			// 在子命令上下文中检查
			return subCmd.IsCompleteFlag(arg)
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
