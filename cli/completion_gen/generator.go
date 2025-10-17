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
	CommandName string
	Commands    []*CommandInfo
	RootFlags   []*FlagInfo
}

// 模板辅助函数
var templateFuncs = template.FuncMap{
	"sub": func(a, b int) int {
		return a - b
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
			spec = fmt.Sprintf("'(-%s ---%s)", flag.Short, flag.Long)
			if flag.Type == "bool" {
				spec += fmt.Sprintf("{-%s,--%s}'[%s]", flag.Short, flag.Long, flag.Usage)
			} else {
				spec += fmt.Sprintf("(-%s,--%s)'[%s]:", flag.Short, flag.Long, flag.Usage)
				spec += getZshFlagValueSpec(flag) + "'"
			}
		} else if flag.Long != "" {
			if flag.Type == "bool" {
				spec = fmt.Sprintf("'--%s[%s]'", flag.Long, flag.Usage)
			} else {
				spec = fmt.Sprintf("'--%s[%s]:", flag.Long, flag.Usage)
				spec += getZshFlagValueSpec(flag) + "'"
			}
		} else if flag.Short != "" {
			if flag.Type == "bool" {
				spec = fmt.Sprintf("'-%s[%s]'", flag.Short, flag.Usage)
			} else {
				spec = fmt.Sprintf("'-%s[%s]:", flag.Short, flag.Usage)
				spec += getZshFlagValueSpec(flag) + "'"
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
				parts = append(parts, fmt.Sprintf("-r -a '%s'", strings.Join(flag.SelectNodes, " ")))
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
			return fmt.Sprintf("(%s)", strings.Join(flag.SelectNodes, " "))
		}
	}
	return ""
}

// GenerateBash 生成 Bash 补全脚本
func (g *ScriptGenerator) GenerateBash() string {
	tmpl := template.Must(template.New("bash").Funcs(templateFuncs).Parse(bashTemplate))

	data := templateData{
		CommandName: g.commandName,
		Commands:    g.cliInfo.Commands,
		RootFlags:   g.cliInfo.RootFlags,
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

	data := templateData{
		CommandName: g.commandName,
		Commands:    g.cliInfo.Commands,
		RootFlags:   g.cliInfo.RootFlags,
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

	data := templateData{
		CommandName: g.commandName,
		Commands:    g.cliInfo.Commands,
		RootFlags:   g.cliInfo.RootFlags,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("# Error generating fish completion: %v\n", err)
	}

	return buf.String()
}
