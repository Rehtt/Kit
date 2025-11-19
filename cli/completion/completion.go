package completion

import (
	"fmt"
	"os"

	"github.com/Rehtt/Kit/cli"
)

// New 创建并初始化补全管理器
// 自动从 FlagSet.Item 读取 FlagItem 并生成补全
// 注册隐藏的 __complete 命令处理补全请求
func New(root *cli.CLI) *CompletionManager {
	com := NewCompletionManager(root)

	c := cli.NewCLI("__complete", "")
	c.Hidden = true
	c.Raw = true
	c.CommandFunc = func(args []string) error {
		return com.HandleCompletion(args)
	}
	root.AddCommand(c)

	return com
}

func NewCompletionCommand(root *cli.CLI) *cli.CLI {
	cm := New(root)
	appName := root.Use
	completionCmd := cli.NewCLI("completion", "生成 Shell 补全脚本")
	completionCmd.CommandFunc = func(args []string) error {
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "用法: %s completion <shell>\n", appName)
			fmt.Fprintln(os.Stderr, "支持的 shell: bash, zsh, fish")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "示例:")
			fmt.Fprintln(os.Stderr, "  # Bash")
			fmt.Fprintf(os.Stderr, "  source <(%s completion bash)\n", appName)
			fmt.Fprintf(os.Stderr, "  %s completion bash > /etc/bash_completion.d/%s\n", appName, appName)
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "  # Zsh")
			fmt.Fprintf(os.Stderr, "  %s completion zsh > \"${fpath[1]}/_%s\"\n", appName, appName)
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "  # Fish")
			fmt.Fprintf(os.Stderr, "  %s completion fish > ~/.config/fish/completions/%s.fish\n", appName, appName)
			return fmt.Errorf("缺少 shell 参数")
		}
		return cm.GenerateCompletion(args[0])
	}
	root.AddCommand(completionCmd)
	return completionCmd
}
