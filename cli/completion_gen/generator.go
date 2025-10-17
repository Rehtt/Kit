// Copyright (c) 2025 Rehtt
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// ScriptGenerator 补全脚本生成器
type ScriptGenerator struct {
	cliInfo     *CLIInfo
	commandName string
}

// NewScriptGenerator 创建脚本生成器
func NewScriptGenerator(cliInfo *CLIInfo, commandName string) *ScriptGenerator {
	return &ScriptGenerator{
		cliInfo:     cliInfo,
		commandName: commandName,
	}
}

// templateData 模板数据
type templateData struct {
	CommandName        string
	Commands           []*CommandInfo // 原始命令树（用于zsh递归）
	RootFlags          []*FlagInfo
	AllCommands        []*FlatCommand // 扁平化的所有命令（用于bash/fish）
	FirstLevelCommands []*CommandInfo // 第一级子命令列表（不包含根命令本身）
}

// FlatCommand 扁平化的命令信息（包含完整路径）
type FlatCommand struct {
	Name        string         // 命令名称
	FullPath    string         // 完整路径，如 "docker container ls"
	PathParts   []string       // 路径部分 ["docker", "container", "ls"]
	Flags       []*FlagInfo    // 该命令的flags
	SubCommands []*CommandInfo // 直接子命令
	Instruction string         // 命令说明
	Depth       int            // 命令深度
	Hidden      bool           // 是否隐藏
}

// 模板辅助函数
var templateFuncs = template.FuncMap{
	"sub": func(a, b int) int {
		return a - b
	},
	"slice": func(s []string, start, end int) []string {
		if start < 0 || end > len(s) || start > end {
			return []string{}
		}
		return s[start:end]
	},
	"index": func(s []string, i int) string {
		if i < 0 || i >= len(s) {
			return ""
		}
		return s[i]
	},
	"flagsToString": func(flags []*FlagInfo) string {
		var items []string
		for _, flag := range flags {
			if flag.Short != "" {
				items = append(items, "-"+flag.Short)
			}
			if flag.Long != "" {
				items = append(items, "--"+flag.Long)
			}
		}
		return strings.Join(items, " ")
	},
	"commandsToString": func(commands []*CommandInfo) string {
		var items []string
		for _, cmd := range commands {
			if !cmd.Hidden {
				items = append(items, cmd.Name)
			}
		}
		return strings.Join(items, " ")
	},
	"zshFlagSpec": func(flag *FlagInfo) string {
		var spec string
		if flag.Short != "" && flag.Long != "" {
			if flag.Type == "bool" {
				spec = fmt.Sprintf("'(-%s --%s)'{-%s,--%s}'[%s]'", flag.Short, flag.Long, flag.Short, flag.Long, flag.Usage)
			} else {
				argName := flag.Long
				if argName == "" {
					argName = flag.Short
				}
				valueSpec := getZshFlagValueSpec(flag)
				if valueSpec != "" {
					spec = fmt.Sprintf("'(-%s --%s)'{-%s,--%s}'[%s]:%s:%s'", flag.Short, flag.Long, flag.Short, flag.Long, flag.Usage, argName, valueSpec)
				} else {
					spec = fmt.Sprintf("'(-%s --%s)'{-%s,--%s}'[%s]:%s:'", flag.Short, flag.Long, flag.Short, flag.Long, flag.Usage, argName)
				}
			}
		} else if flag.Long != "" {
			if flag.Type == "bool" {
				spec = fmt.Sprintf("'--%s[%s]'", flag.Long, flag.Usage)
			} else {
				argName := flag.Long
				if argName == "" {
					argName = "value"
				}
				valueSpec := getZshFlagValueSpec(flag)
				if valueSpec != "" {
					spec = fmt.Sprintf("'--%s[%s]:%s:%s'", flag.Long, flag.Usage, argName, valueSpec)
				} else {
					spec = fmt.Sprintf("'--%s[%s]:%s:'", flag.Long, flag.Usage, argName)
				}
			}
		} else if flag.Short != "" {
			if flag.Type == "bool" {
				spec = fmt.Sprintf("'-%s[%s]'", flag.Short, flag.Usage)
			} else {
				argName := flag.Short
				if argName == "" {
					argName = "value"
				}
				valueSpec := getZshFlagValueSpec(flag)
				if valueSpec != "" {
					spec = fmt.Sprintf("'-%s[%s]:%s:%s'", flag.Short, flag.Usage, argName, valueSpec)
				} else {
					spec = fmt.Sprintf("'-%s[%s]:%s:'", flag.Short, flag.Usage, argName)
				}
			}
		}
		return spec
	},
	"fishFlagSpec": func(commandName string, flag *FlagInfo, condition string) string {
		desc := flag.Usage
		if desc == "" {
			desc = fmt.Sprintf("%s flag", flag.Long)
		}

		baseCmd := fmt.Sprintf("complete -c %s", commandName)
		if condition != "" {
			baseCmd += fmt.Sprintf(" -n '%s'", condition)
		}

		var parts []string
		if flag.Short != "" && flag.Long != "" {
			parts = append(parts, fmt.Sprintf("-s %s -l %s", flag.Short, flag.Long))
		} else if flag.Long != "" {
			parts = append(parts, fmt.Sprintf("-l %s", flag.Long))
		} else if flag.Short != "" {
			parts = append(parts, fmt.Sprintf("-s %s", flag.Short))
		}

		parts = append(parts, fmt.Sprintf("-d '%s'", desc))

		// 添加值补全
		switch flag.ItemType {
		case "file":
			parts = append(parts, "-r -F")
		case "dir":
			parts = append(parts, "-r -a '(__fish_complete_directories)'")
		case "select":
			if len(flag.SelectNodes) > 0 {
				var values []string
				for _, node := range flag.SelectNodes {
					values = append(values, node.Value)
				}
				parts = append(parts, fmt.Sprintf("-r -a '%s'", strings.Join(values, " ")))
			}
		}

		return baseCmd + " " + strings.Join(parts, " ")
	},
}

func getZshFlagValueSpec(flag *FlagInfo) string {
	switch flag.ItemType {
	case "file":
		return "_files"
	case "dir":
		return "_directories"
	case "select":
		if len(flag.SelectNodes) > 0 {
			// 检查是否有描述信息
			hasDesc := false
			for _, node := range flag.SelectNodes {
				if node.Description != "" {
					hasDesc = true
					break
				}
			}

			// 对于有描述的情况，生成带描述的格式
			// zsh 格式: ((value1\:desc1 value2\:desc2))
			if hasDesc {
				var pairs []string
				for _, node := range flag.SelectNodes {
					if node.Description != "" {
						// 转义冒号和空格
						desc := strings.ReplaceAll(node.Description, ":", "\\:")
						desc = strings.ReplaceAll(desc, " ", "\\ ")
						pairs = append(pairs, fmt.Sprintf("%s\\:%s", node.Value, desc))
					} else {
						pairs = append(pairs, node.Value)
					}
				}
				return fmt.Sprintf("((%s))", strings.Join(pairs, " "))
			} else {
				// 只有值，没有描述
				var values []string
				for _, node := range flag.SelectNodes {
					values = append(values, node.Value)
				}
				return fmt.Sprintf("(%s)", strings.Join(values, " "))
			}
		}
	}
	return ""
}

// flattenCommands 递归扁平化命令树，收集所有命令及其完整路径
func (g *ScriptGenerator) flattenCommands() []*FlatCommand {
	var result []*FlatCommand

	// 递归处理每个顶级命令
	// 如果顶级命令的名称是程序名或者是 CommandLine/app/rootCmd 等根命令变量，
	// 则直接处理其子命令，不包含它自己
	for _, cmd := range g.cliInfo.Commands {
		if cmd.Name == g.commandName || cmd.VarName == "CommandLine" ||
			cmd.VarName == "app" || cmd.VarName == "rootCmd" {
			// 这是根命令，直接处理其子命令
			for _, subCmd := range cmd.SubCommands {
				g.flattenCommandsRecursive(subCmd, []string{}, &result)
			}
		} else {
			// 这是一个普通顶级命令
			g.flattenCommandsRecursive(cmd, []string{}, &result)
		}
	}

	return result
}

// flattenCommandsRecursive 递归处理命令树
func (g *ScriptGenerator) flattenCommandsRecursive(cmd *CommandInfo, parentPath []string, result *[]*FlatCommand) {
	// 构建当前命令的路径
	currentPath := append(parentPath, cmd.Name)

	flatCmd := &FlatCommand{
		Name:        cmd.Name,
		FullPath:    strings.Join(currentPath, " "),
		PathParts:   currentPath,
		Flags:       cmd.Flags,
		SubCommands: cmd.SubCommands,
		Instruction: cmd.Instruction,
		Depth:       len(currentPath),
		Hidden:      cmd.Hidden,
	}

	*result = append(*result, flatCmd)

	// 递归处理子命令
	for _, subCmd := range cmd.SubCommands {
		g.flattenCommandsRecursive(subCmd, currentPath, result)
	}
}

// getAllCommandNames 获取所有命令名称（用于bash补全识别）
func getAllCommandNames(commands []*FlatCommand) []string {
	names := make([]string, 0, len(commands))
	seen := make(map[string]bool)

	for _, cmd := range commands {
		if !seen[cmd.Name] {
			names = append(names, cmd.Name)
			seen[cmd.Name] = true
		}
	}

	return names
}

// getFirstLevelCommands 获取第一级子命令（不包含根命令本身）
func (g *ScriptGenerator) getFirstLevelCommands() []*CommandInfo {
	var firstLevel []*CommandInfo

	for _, cmd := range g.cliInfo.Commands {
		if cmd.Name == g.commandName || cmd.VarName == "CommandLine" ||
			cmd.VarName == "app" || cmd.VarName == "rootCmd" {
			// 这是根命令，返回其子命令
			return cmd.SubCommands
		} else {
			// 这是一个普通顶级命令
			firstLevel = append(firstLevel, cmd)
		}
	}

	return firstLevel
}

// GenerateBash 生成 Bash 补全脚本
func (g *ScriptGenerator) GenerateBash() string {
	tmpl := template.Must(template.New("bash").Funcs(templateFuncs).Parse(bashTemplate))

	allCommands := g.flattenCommands()
	firstLevel := g.getFirstLevelCommands()

	data := templateData{
		CommandName:        g.commandName,
		Commands:           g.cliInfo.Commands,
		RootFlags:          g.cliInfo.RootFlags,
		AllCommands:        allCommands,
		FirstLevelCommands: firstLevel,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("# Error generating bash completion: %v\n", err)
	}

	return buf.String()
}

// GenerateZsh 生成 Zsh 补全脚本
func (g *ScriptGenerator) GenerateZsh() string {
	tmpl := template.Must(template.New("zsh").Funcs(templateFuncs).Parse(zshTemplate))

	allCommands := g.flattenCommands()
	firstLevel := g.getFirstLevelCommands()

	data := templateData{
		CommandName:        g.commandName,
		Commands:           g.cliInfo.Commands,
		RootFlags:          g.cliInfo.RootFlags,
		AllCommands:        allCommands,
		FirstLevelCommands: firstLevel,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("# Error generating zsh completion: %v\n", err)
	}

	return buf.String()
}

// GenerateFish 生成 Fish 补全脚本
func (g *ScriptGenerator) GenerateFish() string {
	tmpl := template.Must(template.New("fish").Funcs(templateFuncs).Parse(fishTemplate))

	allCommands := g.flattenCommands()
	firstLevel := g.getFirstLevelCommands()

	data := templateData{
		CommandName:        g.commandName,
		Commands:           g.cliInfo.Commands,
		RootFlags:          g.cliInfo.RootFlags,
		AllCommands:        allCommands,
		FirstLevelCommands: firstLevel,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("# Error generating fish completion: %v\n", err)
	}

	return buf.String()
}
