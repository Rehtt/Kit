//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/Rehtt/Kit/cli"
	"github.com/Rehtt/Kit/cli/completion"
)

func main() {
	root := cli.NewCLI("myapp", "演示补全功能的示例应用")
	root.Usage = "[flags] [command]"

	var config, output, env string
	var verbose bool

	root.StringVarShortLong(&config, "c", "config", "", "配置文件路径")
	root.StringVarShortLong(&output, "o", "output", "", "输出目录")
	root.StringVarShortLong(&env, "e", "env", "dev", "环境")
	root.BoolVarShortLong(&verbose, "v", "verbose", false, "详细输出")

	// 创建补全管理器
	cm := completion.New(root)
	cm.RegisterFileCompletion(root, "config", ".json", ".yaml", ".toml")
	cm.RegisterDirectoryCompletion(root, "output")
	cm.RegisterCustomCompletionPrefixMatches(root, "env", []completion.CompletionItem{
		{Value: "dev", Description: "开发环境"},
		{Value: "test", Description: "测试环境"},
		{Value: "staging", Description: "预发布环境"},
		{Value: "prod", Description: "生产环境"},
	})

	hello := cli.NewCLI("hello", "打印问候语")
	hello.Usage = "[flags] [name]"
	var name string
	hello.StringVar(&name, "name", "world", "要问候的名字")
	hello.CommandFunc = func(args []string) error {
		if verbose {
			fmt.Printf("配置: %s, 输出: %s, 环境: %s\n", config, output, env)
		}
		target := name
		if len(args) > 0 {
			target = args[0]
		}
		fmt.Printf("Hello, %s!\n", target)
		return nil
	}

	build := cli.NewCLI("build", "构建项目")
	build.Usage = "[flags]"
	var target string
	build.StringVar(&target, "target", "all", "构建目标")

	cm.RegisterCustomCompletionPrefixMatches(build, "target", []completion.CompletionItem{
		{Value: "all", Description: "构建所有组件"},
		{Value: "frontend", Description: "只构建前端"},
		{Value: "backend", Description: "只构建后端"},
		{Value: "docs", Description: "只构建文档"},
	})

	build.CommandFunc = func(args []string) error {
		if verbose {
			fmt.Printf("目标: %s, 输出: %s\n", target, output)
		}
		fmt.Printf("正在构建 %s...\n", target)
		return nil
	}

	completion := cli.NewCLI("completion", "生成补全脚本")
	completion.Usage = "[bash|zsh|fish]"
	completion.CommandFunc = func(args []string) error {
		if len(args) == 0 {
			fmt.Println("用法: myapp completion [bash|zsh|fish]")
			return nil
		}
		switch args[0] {
		case "bash", "zsh", "fish":
			return cm.GenerateCompletion(args[0])
		default:
			return fmt.Errorf("不支持的 shell: %s", args[0])
		}
	}

	if err := root.AddCommand(hello, build, completion); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if err := root.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
