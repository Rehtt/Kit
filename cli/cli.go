package cli

import (
	"flag"
	"fmt"
	"os"
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
		*flag.FlagSet
		SubCommands map[string]*CLI
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
	return &CLI{
		Use:         use,
		Instruction: instruction,
		FlagSet:     flag.NewFlagSet(use, flag.ContinueOnError),
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
	w := tabwriter.NewWriter(c.Output(), 0, 0, 2, ' ', 0)
	defer w.Flush()
	if c.Instruction != "" {
		fmt.Fprintf(w, "%s\n\n", c.Instruction)
	}
	fmt.Fprintln(w, "Usage: "+c.Use+" "+c.Usage)
	c.PrintDefaults()
	if len(c.SubCommands) > 0 {
		fmt.Fprintln(w, "\nSubcommands:")
		for _, v := range c.SubCommands {
			fmt.Fprintf(w, "  %s\t%s\n", v.Use, v.Instruction)
		}
	}
}

func (c *CLI) Parse(arguments []string) error {
	c.FlagSet.Usage = c.Help
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

// Parse 别名
func (c *CLI) Run(arguments []string) error {
	return c.Parse(arguments)
}

func AddCommand(cli ...*CLI) error { return CommandLine.AddCommand(cli...) }

// Parse 解析命令行参数（使用默认 CommandLine 实例）
func Parse() error { return CommandLine.Parse(os.Args[1:]) }

// Run 执行命令行参数（使用默认 CommandLine 实例）
func Run() error { return CommandLine.Run(os.Args[1:]) }

// Parsed 判断命令行参数是否已被解析
func Parsed() bool { return CommandLine.Parsed() }

// Args 返回非 flag 参数（使用默认 CommandLine 实例）
func Args() []string { return CommandLine.Args() }

// NArg 返回非 flag 参数的数量（使用默认 CommandLine 实例）
func NArg() int { return CommandLine.NArg() }
