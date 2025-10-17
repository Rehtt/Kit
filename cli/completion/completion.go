package completion

import "github.com/Rehtt/Kit/cli"

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
