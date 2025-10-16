package completion

import "github.com/Rehtt/Kit/cli"

// New 创建补全管理器
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
