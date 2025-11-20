package completion

import (
	"fmt"
	"strings"

	"github.com/Rehtt/Kit/cli"
)

// CompletionManager 补全管理器
type CompletionManager struct {
	hasInit           bool
	completions       map[string]Completion
	sub               map[string]*CompletionManager
	cli               *cli.CLI
	manualCompletions map[cliPtr]map[string]Completion
}

// NewCompletionManager 创建补全管理器
func NewCompletionManager(cli *cli.CLI) *CompletionManager {
	return &CompletionManager{
		completions:       make(map[string]Completion),
		sub:               make(map[string]*CompletionManager),
		cli:               cli,
		manualCompletions: make(map[cliPtr]map[string]Completion),
	}
}

// RegisterCompletion 注册补全（可覆盖自动生成的补全）
func (cm *CompletionManager) RegisterCompletion(cli *cli.CLI, flagName string, completion Completion) {
	if cm.manualCompletions[cli] == nil {
		cm.manualCompletions[cli] = make(map[string]Completion)
	}
	cm.manualCompletions[cli][flagName] = completion
}

// RegisterFileCompletion 注册文件补全
func (cm *CompletionManager) RegisterFileCompletion(cli *cli.CLI, flagName string, extensions ...string) {
	cm.RegisterCompletion(cli, flagName, NewFileCompletion(extensions...))
}

// RegisterDirectoryCompletion 注册目录补全
func (cm *CompletionManager) RegisterDirectoryCompletion(cli *cli.CLI, flagName string) {
	cm.RegisterCompletion(cli, flagName, NewDirectoryCompletion())
}

// RegisterCustomCompletion 注册自定义补全
func (cm *CompletionManager) RegisterCustomCompletion(cli *cli.CLI, flagName string, fn CompletionFunc) {
	cm.RegisterCompletion(cli, flagName, NewCustomCompletion(fn))
}

// RegisterCustomCompletionPrefixMatches 注册前缀匹配补全
func (cm *CompletionManager) RegisterCustomCompletionPrefixMatches(cli *cli.CLI, flagName string, completionItems any) {
	var cis []CompletionItem
	switch items := completionItems.(type) {
	case []string:
		cis = make([]CompletionItem, 0, len(items))
		for _, v := range items {
			cis = append(cis, CompletionItem{Value: v})
		}
	case []CompletionItem:
		cis = items
	default:
		return
	}

	cm.RegisterCustomCompletion(cli, flagName, func(toComplete string) []CompletionItem {
		var matches []CompletionItem
		for _, item := range cis {
			if strings.HasPrefix(item.Value, toComplete) {
				matches = append(matches, item)
			}
		}
		return matches
	})
}

// createCompletionFromFlagItem 根据 FlagItem 创建 Completion
func createCompletionFromFlagItem(item cli.FlagItem) Completion {
	switch item.Type {
	case cli.FlagItemFile:
		return NewFileCompletion()
	case cli.FlagItemDir:
		return NewDirectoryCompletion()
	case cli.FlagItemSelect:
		items := make([]CompletionItem, 0, len(item.Nodes))
		for _, node := range item.Nodes {
			items = append(items, CompletionItem{
				Value:       node.Value,
				Description: node.Description,
			})
		}
		return NewCustomCompletion(func(toComplete string) []CompletionItem {
			var matches []CompletionItem
			for _, ci := range items {
				if strings.HasPrefix(ci.Value, toComplete) {
					matches = append(matches, ci)
				}
			}
			return matches
		})
	default:
		return nil
	}
}

// normalizeFlagName 标准化为长参数名
func (cm *CompletionManager) normalizeFlagName(flagName string) string {
	if cm.cli.FlagSet.ShortLongMap != nil {
		if slValue, exists := cm.cli.FlagSet.ShortLongMap[flagName]; exists && slValue.LongName != "" {
			return slValue.LongName
		}
	}
	return flagName
}

// findSubCommand 查找子命令
func (cm *CompletionManager) findSubCommand(args []string) (subCmd *CompletionManager, remainingArgs []string) {
	if cm.sub != nil {
		for i, arg := range args {
			if sub, exists := cm.sub[arg]; exists {
				return sub, args[i+1:]
			}
		}
	}
	return nil, args
}

// Complete 执行补全
func (cm *CompletionManager) Complete(args []string, toComplete string) []string {
	if subCmd, remaining := cm.findSubCommand(args); subCmd != nil {
		return subCmd.Complete(remaining, toComplete)
	}

	if strings.HasPrefix(toComplete, "-") {
		return NewFlagCompletion(cm.cli.FlagSet).Complete(args, toComplete)
	}

	if len(args) > 0 {
		lastArg := args[len(args)-1]
		if strings.HasPrefix(lastArg, "-") {
			flagName := cm.normalizeFlagName(strings.TrimPrefix(strings.TrimPrefix(lastArg, "--"), "-"))
			if completion, exists := cm.completions[flagName]; exists {
				return completion.Complete(args, toComplete)
			}
		}
	}

	return NewCommandCompletion(cm.cli).Complete(args, toComplete)
}

// CompleteWithDesc 带描述补全
func (cm *CompletionManager) CompleteWithDesc(args []string, toComplete string) []CompletionItem {
	if subCmd, remaining := cm.findSubCommand(args); subCmd != nil {
		return subCmd.CompleteWithDesc(remaining, toComplete)
	}
	if strings.HasPrefix(toComplete, "-") {
		return NewFlagCompletion(cm.cli.FlagSet).CompleteWithDesc(args, toComplete)
	}
	if len(args) > 0 {
		lastArg := args[len(args)-1]
		if strings.HasPrefix(lastArg, "-") {
			flagName := cm.normalizeFlagName(strings.TrimPrefix(strings.TrimPrefix(lastArg, "--"), "-"))
			if completion, exists := cm.completions[flagName]; exists {
				return completion.CompleteWithDesc(args, toComplete)
			}
		}
	}
	return NewCommandCompletion(cm.cli).CompleteWithDesc(args, toComplete)
}

// initSubCommand 初始化子命令补全
func (cm *CompletionManager) initSubCommand(cli *cli.CLI) *CompletionManager {
	ncm := NewCompletionManager(cli)
	if cm.manualCompletions[cli] != nil {
		ncm.manualCompletions = make(map[cliPtr]map[string]Completion)
		ncm.manualCompletions[cli] = cm.manualCompletions[cli]
	}
	ncm.init()
	return ncm
}

// init 初始化补全管理器
func (cm *CompletionManager) init() {
	defer func() { cm.hasInit = true }()

	if cm.cli.FlagSet.Item != nil {
		for slValue, flagItem := range cm.cli.FlagSet.Item {
			flagName := slValue.LongName
			if flagName == "" {
				flagName = slValue.ShortName
			}

			if completion := createCompletionFromFlagItem(flagItem); completion != nil {
				cm.completions[flagName] = completion
			}
		}
	}

	for name, completion := range cm.manualCompletions[cm.cli] {
		cm.completions[name] = completion
	}

	cm.cli.SubCommands.Range(func(subCli *cli.CLI) bool {
		if !subCli.Hidden {
			cm.sub[subCli.Use] = cm.initSubCommand(subCli)
		}
		return true
	})
}

// parseCompletionArgs 解析补全参数
func (cm *CompletionManager) parseCompletionArgs(args []string) (remainingArgs []string, toComplete string) {
	if len(args) == 0 {
		return args, ""
	}

	lastArg := args[len(args)-1]
	trimmedLastArg := strings.TrimSpace(lastArg)

	if strings.HasPrefix(lastArg, "-") && cm.cli.IsCompleteFlagInContext(lastArg, args) {
		return args, ""
	}

	if trimmedLastArg == "" && len(args) >= 2 {
		secondLastArg := args[len(args)-2]
		if strings.HasPrefix(secondLastArg, "-") && cm.cli.IsCompleteFlagInContext(secondLastArg, args[:len(args)-1]) {
			return args[:len(args)-1], ""
		}
	}

	return args[:len(args)-1], trimmedLastArg
}

// printCompletions 打印补全结果
func (cm *CompletionManager) printCompletions(items []CompletionItem, format string) {
	for _, item := range items {
		if item.Description != "" && (format == "zsh" || format == "fish") {
			separator := ":"
			if format == "fish" {
				separator = "\t"
			}
			fmt.Printf("%s%s%s\n", item.Value, separator, item.Description)
		} else {
			fmt.Println(item.Value)
		}
	}
}

// HandleCompletion 处理补全请求
func (cm *CompletionManager) HandleCompletion(args []string) error {
	if !cm.hasInit {
		cm.init()
	}

	format := "simple"
	if len(args) > 0 && strings.HasPrefix(args[0], "--format=") {
		format = strings.TrimPrefix(args[0], "--format=")
		args = args[1:]
	}

	remainingArgs, toComplete := cm.parseCompletionArgs(args)

	if format == "zsh" || format == "fish" {
		cm.printCompletions(cm.CompleteWithDesc(remainingArgs, toComplete), format)
	} else {
		completions := cm.Complete(remainingArgs, toComplete)
		for _, c := range completions {
			fmt.Println(c)
		}
	}
	return nil
}
