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
	completionCmd := cli.NewCLI("completion", "生成 Shell 补全脚本")
	completionCmd.CommandFunc = func(args []string) error {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "用法: myapp completion <shell>")
			fmt.Fprintln(os.Stderr, "支持的 shell: bash, zsh, fish")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "示例:")
			fmt.Fprintln(os.Stderr, "  # Bash")
			fmt.Fprintln(os.Stderr, "  source <(myapp completion bash)")
			fmt.Fprintln(os.Stderr, "  myapp completion bash > /etc/bash_completion.d/myapp")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "  # Zsh")
			fmt.Fprintln(os.Stderr, "  myapp completion zsh > \"${fpath[1]}/_myapp\"")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "  # Fish")
			fmt.Fprintln(os.Stderr, "  myapp completion fish > ~/.config/fish/completions/myapp.fish")
			return fmt.Errorf("缺少 shell 参数")
		}
		return cm.GenerateCompletion(args[0])
	}
	root.AddCommand(completionCmd)
	return completionCmd
}
