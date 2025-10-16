package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
		SubCommands       map[string]*CLI
		CompletionManager *CompletionManager
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
	cli.CompletionManager = NewCompletionManager(cli)
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
			subCommands = append(subCommands, v.Use)
		}
		sort.Strings(subCommands)

		fmt.Fprintln(w, "\nSubcommands:")
		for _, use := range subCommands {
			sub := c.SubCommands[use]
			fmt.Fprintf(w, "  %s\t%s\n", sub.Use, sub.Instruction)
		}
	}
}

func (c *CLI) Parse(arguments []string) error {
	c.FlagSet.Usage = c.Help

	// 处理补全命令
	if len(arguments) > 0 && arguments[0] == "__complete" {
		return c.handleCompletion(arguments[1:])
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

// handleCompletion 处理补全请求
func (c *CLI) handleCompletion(args []string) error {
	// 检查是否指定了格式
	format := "simple"
	if len(args) > 0 && strings.HasPrefix(args[0], "--format=") {
		format = strings.TrimPrefix(args[0], "--format=")
		args = args[1:]
	}

	var toComplete string
	if len(args) > 0 {
		lastArg := args[len(args)-1]

		// 清理空白字符
		trimmedLastArg := strings.TrimSpace(lastArg)

		// 情况1: 最后一个参数是完整的参数名（如 "--type"）
		if strings.HasPrefix(lastArg, "-") && c.isCompleteFlagInContext(lastArg, args) {
			// 最后一个参数是完整的参数名，用户想要补全该参数的值
			toComplete = ""
			// args 保持不变，包含该参数
		} else if trimmedLastArg == "" && len(args) >= 2 {
			// 情况2: 最后一个参数是空白字符，检查倒数第二个参数是否是完整参数名
			secondLastArg := args[len(args)-2]
			if strings.HasPrefix(secondLastArg, "-") && c.isCompleteFlagInContext(secondLastArg, args[:len(args)-1]) {
				// 倒数第二个参数是完整的参数名，用户想要补全该参数的值
				toComplete = ""
				args = args[:len(args)-1] // 移除空白参数，保留参数名
			} else {
				// 正常情况：最后一个参数是要补全的内容
				toComplete = trimmedLastArg
				args = args[:len(args)-1]
			}
		} else {
			// 情况3: 正常情况，最后一个参数是要补全的内容
			toComplete = trimmedLastArg
			args = args[:len(args)-1]
		}
	}

	switch format {
	case "zsh", "fish":
		// 使用带描述的格式
		items := c.CompletionManager.CompleteWithDesc(args, toComplete)
		for _, item := range items {
			if item.Description != "" {
				if format == "zsh" {
					fmt.Printf("%s:%s\n", item.Value, item.Description)
				} else { // fish
					fmt.Printf("%s\t%s\n", item.Value, item.Description)
				}
			} else {
				fmt.Println(item.Value)
			}
		}
	default:
		// 简单格式（bash 或默认）
		completions := c.CompletionManager.Complete(args, toComplete)
		for _, completion := range completions {
			fmt.Println(completion)
		}
	}
	return nil
}

// isCompleteFlag 检查给定的字符串是否是一个完整的参数名
func (c *CLI) isCompleteFlag(arg string) bool {
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
	if c.FlagSet.shortLongMap != nil {
		if _, exists := c.FlagSet.shortLongMap[flagName]; exists {
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

// isCompleteFlagInContext 在给定上下文中检查参数是否完整
func (c *CLI) isCompleteFlagInContext(arg string, args []string) bool {
	// 如果有子命令，需要在子命令的上下文中检查
	if len(args) > 0 && c.SubCommands != nil {
		if subCmd, exists := c.SubCommands[args[0]]; exists {
			// 在子命令上下文中检查
			return subCmd.isCompleteFlag(arg)
		}
	}

	// 在当前命令上下文中检查
	return c.isCompleteFlag(arg)
}

// RegisterFileCompletion 为指定参数注册文件补全
func (c *CLI) RegisterFileCompletion(flagName string, extensions ...string) {
	c.CompletionManager.RegisterFileCompletion(flagName, extensions...)
}

// RegisterDirectoryCompletion 为指定参数注册目录补全
func (c *CLI) RegisterDirectoryCompletion(flagName string) {
	c.CompletionManager.RegisterDirectoryCompletion(flagName)
}

// RegisterCustomCompletion 为指定参数注册自定义补全
// 支持两种函数类型：
//   - func(string) []string - 简单补全，只返回值
//   - func(string) []CompletionItem - 带描述补全，返回值和描述
func (c *CLI) RegisterCustomCompletion(flagName string, fn CompletionFunc) {
	c.CompletionManager.RegisterCustomCompletion(flagName, fn)
}

// RegisterCustomCompletionPrefixMatches 为指定参数注册自定义匹配补全
// 支持 []string 和 []CompletionItem
func (c *CLI) RegisterCustomCompletionPrefixMatches(flagName string, completionItems any) {
	var cis []CompletionItem
	switch completionItems := completionItems.(type) {
	case []string:
		cis = make([]CompletionItem, 0, len(completionItems))
		for _, v := range completionItems {
			cis = append(cis, CompletionItem{Value: v})
		}
	case []CompletionItem:
		cis = completionItems
	}
	c.RegisterCustomCompletion(flagName, func(toComplete string) []CompletionItem {
		var matches []CompletionItem
		for _, t := range cis {
			if strings.HasPrefix(t.Value, toComplete) {
				matches = append(matches, t)
			}
		}
		return matches
	})
}

// GenerateCompletion 生成指定 shell 的补全脚本
func (c *CLI) GenerateCompletion(shell string, cname ...string) error {
	var cmdName string
	if len(cname) > 0 {
		cmdName = cname[0]
	} else {
		path, _ := os.Executable()
		_, cmdName = filepath.Split(path)
	}

	switch shell {
	case "bash":
		return c.CompletionManager.GenerateBashCompletion(os.Stdout, cmdName)
	case "zsh":
		return c.CompletionManager.GenerateZshCompletion(os.Stdout, cmdName)
	case "fish":
		return c.CompletionManager.GenerateFishCompletion(os.Stdout, cmdName)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

func AddCommand(cli ...*CLI) error { return CommandLine.AddCommand(cli...) }

func Parse() error { return CommandLine.Parse(os.Args[1:]) }

func Run() error { return CommandLine.Run(os.Args[1:]) }

func Parsed() bool { return CommandLine.Parsed() }

func Args() []string { return CommandLine.Args() }

func NArg() int { return CommandLine.NArg() }

// RegisterFileCompletion 为全局 CommandLine 注册文件补全
func RegisterFileCompletion(flagName string, extensions ...string) {
	CommandLine.RegisterFileCompletion(flagName, extensions...)
}

// RegisterDirectoryCompletion 为全局 CommandLine 注册目录补全
func RegisterDirectoryCompletion(flagName string) {
	CommandLine.RegisterDirectoryCompletion(flagName)
}

// RegisterCustomCompletion 为全局 CommandLine 注册自定义补全
// 支持两种函数类型：
//   - func(string) []string - 简单补全，只返回值
//   - func(string) []CompletionItem - 带描述补全，返回值和描述
func RegisterCustomCompletion(flagName string, fn CompletionFunc) {
	CommandLine.RegisterCustomCompletion(flagName, fn)
}

// GenerateCompletion 为全局 CommandLine 生成补全脚本
func GenerateCompletion(shell, cmdName string) error {
	return CommandLine.GenerateCompletion(shell, cmdName)
}
