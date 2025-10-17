// Copyright (c) 2025 Rehtt
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// @Author: Rehtt dsreshiram@gmail.com
// @Date: 2025/10/17

// completion_exec 是一个可执行的二进制工具，用于解析 Go 项目代码
// 它能识别使用 github.com/Rehtt/Kit/cli 注册的命令和 flag
// 并生成对应的 shell 补全脚本（支持 bash/zsh/fish）
package main

import (
	"fmt"
	"os"

	"github.com/Rehtt/Kit/cli"
)

var (
	dir       string
	shell     string
	cmdName   string
	output    string
	recursive bool
	verbose   bool
)

func main() {
	app := cli.NewCLI("completion_gen", "CLI 补全脚本生成器 - 从 Go 代码自动生成 shell 补全脚本")
	app.Usage = "[options] <path>"

	// 注册 flags
	app.StringVarShortLong(&shell, "s", "shell", "bash", "生成的 shell 类型",
		cli.NewFlagItemSelectString("bash", "zsh", "fish"))
	app.StringVarShortLong(&cmdName, "n", "name", "", "命令名称（默认从 main 包推断）")
	app.StringVarShortLong(&output, "o", "output", "", "输出文件路径（默认输出到 stdout）",
		cli.NewFlagItemFile())
	app.BoolVarShortLong(&recursive, "r", "recursive", true, "是否递归扫描子目录")
	app.BoolVarShortLong(&verbose, "v", "verbose", false, "显示详细的解析过程")

	// 设置主命令处理函数
	app.CommandFunc = func(args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("错误: 请指定项目路径")
		}
		dir = args[0]
		return run()
	}

	// 解析并运行
	if err := app.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

func run() error {
	// 验证参数
	if shell != "bash" && shell != "zsh" && shell != "fish" {
		return fmt.Errorf("错误: 不支持的 shell 类型 '%s'，仅支持 bash/zsh/fish", shell)
	}

	// 解析项目
	parser := NewParser(dir, recursive, verbose)
	cliInfo, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("解析错误: %w", err)
	}

	// 推断命令名称
	commandName := cmdName
	if commandName == "" {
		commandName = cliInfo.InferCommandName()
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "命令名称: %s\n", commandName)
		fmt.Fprintf(os.Stderr, "找到 %d 个顶级命令\n", len(cliInfo.Commands))
		fmt.Fprintf(os.Stderr, "找到 %d 个 flag\n", len(cliInfo.RootFlags))
	}

	// 生成补全脚本
	generator := NewScriptGenerator(cliInfo, commandName)
	var scriptContent string

	switch shell {
	case "bash":
		scriptContent = generator.GenerateBash()
	case "zsh":
		scriptContent = generator.GenerateZsh()
	case "fish":
		scriptContent = generator.GenerateFish()
	}

	// 输出
	if output != "" {
		if err := os.WriteFile(output, []byte(scriptContent), 0o644); err != nil {
			return fmt.Errorf("写入文件错误: %w", err)
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "补全脚本已写入: %s\n", output)
		}
	} else {
		fmt.Print(scriptContent)
	}

	return nil
}
