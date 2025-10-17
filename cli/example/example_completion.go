//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/Rehtt/Kit/cli"
	"github.com/Rehtt/Kit/cli/completion"
)

func main() {
	// 创建主命令
	root := cli.NewCLI("myapp", "演示自动补全功能的示例应用")

	// ========== 自动补全示例 ==========
	// 这些 flag 会自动生成补全，无需手动注册

	// 文件补全
	root.FlagSet.StringShortLong("c", "config", "", "配置文件",
		cli.NewFlagItemFile())

	// 目录补全
	root.FlagSet.StringShortLong("d", "dir", ".", "工作目录",
		cli.NewFlagItemDir())

	// 选择项补全（简单）
	root.FlagSet.StringShortLong("f", "format", "json", "输出格式",
		cli.NewFlagItemSelectString("json", "yaml", "xml", "toml"))

	// 选择项补全（带描述）
	root.FlagSet.StringShortLong("l", "level", "info", "日志级别",
		cli.NewFlagItemSelect(
			cli.FlagItemNode{Value: "debug", Description: "调试模式 - 显示所有日志"},
			cli.FlagItemNode{Value: "info", Description: "信息模式 - 显示一般信息"},
			cli.FlagItemNode{Value: "warn", Description: "警告模式 - 只显示警告"},
			cli.FlagItemNode{Value: "error", Description: "错误模式 - 只显示错误"},
		))

	// 普通 flag（无补全）
	root.FlagSet.BoolShortLong("v", "verbose", false, "详细输出")

	root.CommandFunc = func(args []string) error {
		fmt.Println("主命令执行")
		return nil
	}

	// ========== 子命令示例 ==========
	build := cli.NewCLI("build", "构建项目")
	build.StringShortLong("t", "target", "", "构建目标",
		cli.NewFlagItemSelectString("all", "frontend", "backend", "api", "worker"))
	build.FlagSet.StringShortLong("o", "output", "", "输出文件",
		cli.NewFlagItemFile())
	build.CommandFunc = func(args []string) error {
		fmt.Println("构建命令执行")
		return nil
	}
	root.AddCommand(build)

	deploy := cli.NewCLI("deploy", "部署应用")
	deploy.FlagSet.StringShortLong("e", "env", "", "部署环境",
		cli.NewFlagItemSelect(
			cli.FlagItemNode{Value: "dev", Description: "开发环境"},
			cli.FlagItemNode{Value: "staging", Description: "预发布环境"},
			cli.FlagItemNode{Value: "prod", Description: "生产环境"},
		))
	deploy.FlagSet.StringShortLong("r", "region", "", "部署区域",
		cli.NewFlagItemSelectString("us-east", "us-west", "eu-central", "ap-southeast"))
	deploy.CommandFunc = func(args []string) error {
		fmt.Println("部署命令执行")
		return nil
	}
	root.AddCommand(deploy)

	// ========== 创建补全管理器 ==========
	// 重要：这会自动从上面定义的 FlagItem 生成补全
	cm := completion.New(root)

	// ========== 可选：手动注册/覆盖补全 ==========
	// 如果需要更复杂的补全逻辑，可以手动注册
	cm.RegisterCustomCompletion(build, "target", func(toComplete string) []completion.CompletionItem {
		// 这会覆盖自动生成的补全
		return []completion.CompletionItem{
			{Value: "all", Description: "构建所有目标"},
			{Value: "frontend", Description: "只构建前端"},
			{Value: "backend", Description: "只构建后端"},
			{Value: "api", Description: "只构建 API 服务"},
			{Value: "worker", Description: "只构建后台任务"},
			{Value: "mobile", Description: "构建移动端应用"},
		}
	})

	// ========== 添加补全脚本生成命令 ==========
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

	// 运行应用
	if err := root.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}
