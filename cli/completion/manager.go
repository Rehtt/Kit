package completion

import (
	"fmt"
	"strings"

	"github.com/Rehtt/Kit/cli"
)

// CompletionManager 补全管理器
type CompletionManager struct {
	hasInit     bool
	completions map[string]Completion
	sub         map[string]*CompletionManager
	cli         *cli.CLI

	allFlags map[cliPtr]map[string]Completion
}

// NewCompletionManager 创建补全管理器
func NewCompletionManager(cli *cli.CLI) *CompletionManager {
	return &CompletionManager{
		completions: make(map[string]Completion),
		sub:         make(map[string]*CompletionManager),
		cli:         cli,
	}
}

// RegisterCompletion 注册补全
func (cm *CompletionManager) RegisterCompletion(cli *cli.CLI, flagName string, completion Completion) {
	if cm.allFlags == nil {
		cm.allFlags = make(map[cliPtr]map[string]Completion)
	}
	if cm.allFlags[cli] == nil {
		cm.allFlags[cli] = make(map[string]Completion)
	}
	cm.allFlags[cli][flagName] = completion
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
	switch completionItems := completionItems.(type) {
	case []string:
		cis = make([]CompletionItem, 0, len(completionItems))
		for _, v := range completionItems {
			cis = append(cis, CompletionItem{Value: v})
		}
	case []CompletionItem:
		cis = completionItems
	}
	cm.RegisterCustomCompletion(cli, flagName, func(toComplete string) []CompletionItem {
		var matches []CompletionItem
		for _, t := range cis {
			if strings.HasPrefix(t.Value, toComplete) {
				matches = append(matches, t)
			}
		}
		return matches
	})
}

// Complete 执行补全
func (cm *CompletionManager) Complete(args []string, toComplete string) []string {
	if len(args) > 0 && cm.sub != nil {
		if subCmd, exists := cm.sub[args[0]]; exists {
			return subCmd.Complete(args[1:], toComplete)
		}
	}

	if strings.HasPrefix(toComplete, "-") {
		flagCompletion := NewFlagCompletion(cm.cli.FlagSet)
		return flagCompletion.Complete(args, toComplete)
	}

	if len(args) > 0 {
		lastArg := args[len(args)-1]
		if strings.HasPrefix(lastArg, "-") {
			flagName := strings.TrimPrefix(strings.TrimPrefix(lastArg, "--"), "-")

			if cm.cli.FlagSet.ShortLongMap != nil {
				if slValue, exists := cm.cli.FlagSet.ShortLongMap[flagName]; exists && slValue.LongName != "" {
					flagName = slValue.LongName
				}
			}

			if completion, exists := cm.completions[flagName]; exists {
				return completion.Complete(args, toComplete)
			}
		}
	}

	cmdCompletion := NewCommandCompletion(cm.cli)
	return cmdCompletion.Complete(args, toComplete)
}

// CompleteWithDesc 带描述的补全
func (cm *CompletionManager) CompleteWithDesc(args []string, toComplete string) []CompletionItem {
	if len(args) > 0 && cm.sub != nil {
		if subCmd, exists := cm.sub[args[0]]; exists {
			return subCmd.CompleteWithDesc(args[1:], toComplete)
		}
	}

	if strings.HasPrefix(toComplete, "-") {
		flagCompletion := NewFlagCompletion(cm.cli.FlagSet)
		return flagCompletion.CompleteWithDesc(args, toComplete)
	}

	if len(args) > 0 {
		lastArg := args[len(args)-1]
		if strings.HasPrefix(lastArg, "-") {
			flagName := strings.TrimPrefix(strings.TrimPrefix(lastArg, "--"), "-")

			if cm.cli.FlagSet.ShortLongMap != nil {
				if slValue, exists := cm.cli.FlagSet.ShortLongMap[flagName]; exists && slValue.LongName != "" {
					flagName = slValue.LongName
				}
			}

			if completion, exists := cm.completions[flagName]; exists {
				return completion.CompleteWithDesc(args, toComplete)
			}
		}
	}

	cmdCompletion := NewCommandCompletion(cm.cli)
	return cmdCompletion.CompleteWithDesc(args, toComplete)
}

// init 初始化补全管理器
func (cm *CompletionManager) init() {
	defer func() { cm.hasInit = true }()

	if cm.allFlags[cm.cli] != nil {
		for name, completion := range cm.allFlags[cm.cli] {
			cm.completions[name] = completion
		}
	}

	for _, cli := range cm.cli.SubCommands {
		if cli.Hidden {
			continue
		}
		if _, ok := cm.completions[cli.Use]; !ok {
			cm.completions[cli.Use] = NewCommandCompletion(cli)
			ncm := NewCompletionManager(cli)
			if cm.allFlags[cli] != nil {
				if ncm.allFlags == nil {
					ncm.allFlags = make(map[cliPtr]map[string]Completion)
				}
				ncm.allFlags[cli] = cm.allFlags[cli]
			}
			cm.sub[cli.Use] = ncm
			ncm.init()
		}
		// 子命令的参数补全只在子命令管理器中生效
	}
}

// HandleCompletion 处理补全请求
func (cm *CompletionManager) HandleCompletion(args []string) error {
	if !cm.hasInit {
		cm.init()
	}
	// 检查格式参数
	format := "simple"
	if len(args) > 0 && strings.HasPrefix(args[0], "--format=") {
		format = strings.TrimPrefix(args[0], "--format=")
		args = args[1:]
	}

	var toComplete string
	if len(args) > 0 {
		lastArg := args[len(args)-1]

		trimmedLastArg := strings.TrimSpace(lastArg)

		if strings.HasPrefix(lastArg, "-") && cm.cli.IsCompleteFlagInContext(lastArg, args) {
			toComplete = ""
		} else if trimmedLastArg == "" && len(args) >= 2 {
			secondLastArg := args[len(args)-2]
			if strings.HasPrefix(secondLastArg, "-") && cm.cli.IsCompleteFlagInContext(secondLastArg, args[:len(args)-1]) {
				toComplete = ""
				args = args[:len(args)-1]
			} else {
				toComplete = trimmedLastArg
				args = args[:len(args)-1]
			}
		} else {
			toComplete = trimmedLastArg
			args = args[:len(args)-1]
		}
	}

	switch format {
	case "zsh", "fish":
		items := cm.CompleteWithDesc(args, toComplete)
		for _, item := range items {
			if item.Description != "" {
				if format == "zsh" {
					fmt.Printf("%s:%s\n", item.Value, item.Description)
				} else { // fish
					fmt.Printf("%s\t%s\n", item.Value, item.Description)
				}
			} else {
				fmt.Println(item.Value)
			}
		}
	default:
		completions := cm.Complete(args, toComplete)
		for _, completion := range completions {
			fmt.Println(completion)
		}
	}
	return nil
}
