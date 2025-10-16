package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

type CompletionType int

const (
	CompletionTypeCommand CompletionType = iota
	CompletionTypeFlag
	CompletionTypeFile
	CompletionTypeDirectory
	CompletionTypeCustom
)

// CompletionItem 补全项
type CompletionItem struct {
	Value       string
	Description string
}

// CompletionFunc 自定义补全函数，支持 []string 或 []CompletionItem 返回值
type CompletionFunc interface{}

func normalizeCompletionFunc(fn CompletionFunc) func(string) []CompletionItem {
	switch f := fn.(type) {
	case func(string) []string:
		return func(s string) []CompletionItem {
			values := f(s)
			items := make([]CompletionItem, len(values))
			for i, v := range values {
				items[i] = CompletionItem{Value: v}
			}
			return items
		}
	case func(string) []CompletionItem:
		return f
	default:
		return func(string) []CompletionItem { return nil }
	}
}

type Completion interface {
	Complete(args []string, toComplete string) []string
	CompleteWithDesc(args []string, toComplete string) []CompletionItem
	GetType() CompletionType
}

// CommandCompletion 命令补全
type CommandCompletion struct {
	cli *CLI
}

func NewCommandCompletion(cli *CLI) *CommandCompletion {
	return &CommandCompletion{cli: cli}
}

func (c *CommandCompletion) Complete(args []string, toComplete string) []string {
	var suggestions []string

	if c.cli.SubCommands != nil {
		for name := range c.cli.SubCommands {
			if strings.HasPrefix(name, toComplete) {
				suggestions = append(suggestions, name)
			}
		}
	}

	sort.Strings(suggestions)
	return suggestions
}

func (c *CommandCompletion) CompleteWithDesc(args []string, toComplete string) []CompletionItem {
	var items []CompletionItem

	if c.cli.SubCommands != nil {
		for name, subCmd := range c.cli.SubCommands {
			if strings.HasPrefix(name, toComplete) {
				items = append(items, CompletionItem{
					Value:       name,
					Description: subCmd.Instruction,
				})
			}
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Value < items[j].Value
	})
	return items
}

func (c *CommandCompletion) GetType() CompletionType {
	return CompletionTypeCommand
}

// FlagCompletion 参数补全
type FlagCompletion struct {
	flagSet *FlagSet
}

func NewFlagCompletion(flagSet *FlagSet) *FlagCompletion {
	return &FlagCompletion{flagSet: flagSet}
}

func (f *FlagCompletion) Complete(args []string, toComplete string) []string {
	var suggestions []string
	processed := make(map[string]bool)

	f.flagSet.VisitAll(func(flag *flag.Flag) {
		if processed[flag.Name] {
			return
		}

		if f.flagSet.shortLongMap != nil {
			if slValue, exists := f.flagSet.shortLongMap[flag.Name]; exists {
				if slValue.ShortName != "" {
					shortFlag := "-" + slValue.ShortName
					if strings.HasPrefix(shortFlag, toComplete) {
						suggestions = append(suggestions, shortFlag)
					}
					processed[slValue.ShortName] = true
				}

				if slValue.LongName != "" {
					longFlag := "--" + slValue.LongName
					if strings.HasPrefix(longFlag, toComplete) {
						suggestions = append(suggestions, longFlag)
					}
					processed[slValue.LongName] = true
				}
				processed[flag.Name] = true
				return
			}
		}

		longFlag := "--" + flag.Name
		if strings.HasPrefix(longFlag, toComplete) {
			suggestions = append(suggestions, longFlag)
		}
		processed[flag.Name] = true
	})

	sort.Strings(suggestions)
	return suggestions
}

func (f *FlagCompletion) CompleteWithDesc(args []string, toComplete string) []CompletionItem {
	var items []CompletionItem
	processed := make(map[string]bool)

	f.flagSet.VisitAll(func(flag *flag.Flag) {
		if processed[flag.Name] {
			return
		}

		if f.flagSet.shortLongMap != nil {
			if slValue, exists := f.flagSet.shortLongMap[flag.Name]; exists {
				if slValue.ShortName != "" {
					shortFlag := "-" + slValue.ShortName
					if strings.HasPrefix(shortFlag, toComplete) {
						items = append(items, CompletionItem{
							Value:       shortFlag,
							Description: flag.Usage,
						})
					}
					processed[slValue.ShortName] = true
				}

				if slValue.LongName != "" {
					longFlag := "--" + slValue.LongName
					if strings.HasPrefix(longFlag, toComplete) {
						items = append(items, CompletionItem{
							Value:       longFlag,
							Description: flag.Usage,
						})
					}
					processed[slValue.LongName] = true
				}
				processed[flag.Name] = true
				return
			}
		}

		longFlag := "--" + flag.Name
		if strings.HasPrefix(longFlag, toComplete) {
			items = append(items, CompletionItem{
				Value:       longFlag,
				Description: flag.Usage,
			})
		}
		processed[flag.Name] = true
	})

	sort.Slice(items, func(i, j int) bool {
		return items[i].Value < items[j].Value
	})
	return items
}

func (f *FlagCompletion) GetType() CompletionType {
	return CompletionTypeFlag
}

// FileCompletion 文件路径补全
type FileCompletion struct {
	extensions []string
	dirOnly    bool
}

func NewFileCompletion(extensions ...string) *FileCompletion {
	return &FileCompletion{extensions: extensions}
}

func NewDirectoryCompletion() *FileCompletion {
	return &FileCompletion{dirOnly: true}
}

func (f *FileCompletion) Complete(args []string, toComplete string) []string {
	var suggestions []string

	dir := filepath.Dir(toComplete)
	base := filepath.Base(toComplete)

	if base == "." {
		base = ""
	}

	if dir == "." && !strings.Contains(toComplete, "/") {
		dir = "."
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return suggestions
	}

	for _, entry := range entries {
		name := entry.Name()

		if strings.HasPrefix(name, ".") && !strings.HasPrefix(toComplete, ".") {
			continue
		}

		if base != "" && !strings.HasPrefix(name, base) {
			continue
		}

		var fullPath string
		if dir == "." {
			fullPath = name
		} else {
			fullPath = filepath.Join(dir, name)
		}

		if entry.IsDir() {
			suggestions = append(suggestions, fullPath+"/")
		} else if !f.dirOnly {
			if len(f.extensions) == 0 {
				suggestions = append(suggestions, fullPath)
			} else {
				ext := filepath.Ext(name)
				if slices.Contains(f.extensions, ext) {
					suggestions = append(suggestions, fullPath)
				}
			}
		}
	}

	sort.Strings(suggestions)
	return suggestions
}

func (f *FileCompletion) CompleteWithDesc(args []string, toComplete string) []CompletionItem {
	suggestions := f.Complete(args, toComplete)
	items := make([]CompletionItem, len(suggestions))
	for i, s := range suggestions {
		desc := "文件"
		if strings.HasSuffix(s, "/") {
			desc = "目录"
		} else if ext := filepath.Ext(s); ext != "" {
			desc = ext + " 文件"
		}
		items[i] = CompletionItem{Value: s, Description: desc}
	}
	return items
}

func (f *FileCompletion) GetType() CompletionType {
	if f.dirOnly {
		return CompletionTypeDirectory
	}
	return CompletionTypeFile
}

// CustomCompletion 自定义补全
type CustomCompletion struct {
	fn func(string) []CompletionItem
}

// NewCustomCompletion 创建自定义补全，支持 func(string)[]string 或 func(string)[]CompletionItem
func NewCustomCompletion(fn CompletionFunc) *CustomCompletion {
	return &CustomCompletion{fn: normalizeCompletionFunc(fn)}
}

func (c *CustomCompletion) Complete(args []string, toComplete string) []string {
	items := c.fn(toComplete)
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = item.Value
	}
	return result
}

func (c *CustomCompletion) CompleteWithDesc(args []string, toComplete string) []CompletionItem {
	return c.fn(toComplete)
}

func (c *CustomCompletion) GetType() CompletionType {
	return CompletionTypeCustom
}

// CompletionManager 补全管理
type CompletionManager struct {
	completions map[string]Completion
	cli         *CLI
}

func NewCompletionManager(cli *CLI) *CompletionManager {
	return &CompletionManager{
		completions: make(map[string]Completion),
		cli:         cli,
	}
}

func (cm *CompletionManager) RegisterCompletion(flagName string, completion Completion) {
	cm.completions[flagName] = completion
}

func (cm *CompletionManager) RegisterFileCompletion(flagName string, extensions ...string) {
	cm.RegisterCompletion(flagName, NewFileCompletion(extensions...))
}

func (cm *CompletionManager) RegisterDirectoryCompletion(flagName string) {
	cm.RegisterCompletion(flagName, NewDirectoryCompletion())
}

// RegisterCustomCompletion 注册自定义补全，支持 func(string)[]string 或 func(string)[]CompletionItem
func (cm *CompletionManager) RegisterCustomCompletion(flagName string, fn CompletionFunc) {
	cm.RegisterCompletion(flagName, NewCustomCompletion(fn))
}

func (cm *CompletionManager) Complete(args []string, toComplete string) []string {
	if len(args) > 0 && cm.cli.SubCommands != nil {
		if subCmd, exists := cm.cli.SubCommands[args[0]]; exists {
			return subCmd.CompletionManager.Complete(args[1:], toComplete)
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

			if cm.cli.FlagSet.shortLongMap != nil {
				if slValue, exists := cm.cli.FlagSet.shortLongMap[flagName]; exists && slValue.LongName != "" {
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

func (cm *CompletionManager) CompleteWithDesc(args []string, toComplete string) []CompletionItem {
	if len(args) > 0 && cm.cli.SubCommands != nil {
		if subCmd, exists := cm.cli.SubCommands[args[0]]; exists {
			return subCmd.CompletionManager.CompleteWithDesc(args[1:], toComplete)
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

			if cm.cli.FlagSet.shortLongMap != nil {
				if slValue, exists := cm.cli.FlagSet.shortLongMap[flagName]; exists && slValue.LongName != "" {
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

func (cm *CompletionManager) GenerateBashCompletion(w io.Writer, cmdName string) error {
	script := fmt.Sprintf(`# bash completion for %s
_%s_completion() {
    local cur prev words cword
    
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion || return
    else
        COMPREPLY=()
        cur="${COMP_WORDS[COMP_CWORD]}"
        prev="${COMP_WORDS[COMP_CWORD-1]}"
    fi

    local args=()
    for ((i=1; i<COMP_CWORD; i++)); do
        args+=("${COMP_WORDS[i]}")
    done

    local completions
    completions=$(%s __complete "${args[@]}" "$cur" 2>/dev/null)
    
    if [[ $? -eq 0 ]]; then
        while IFS= read -r line; do
            COMPREPLY+=("$line")
        done < <(compgen -W "$completions" -- "$cur")
    fi
}

complete -F _%s_completion %s
`, cmdName, cmdName, cmdName, cmdName, cmdName)

	_, err := w.Write([]byte(script))
	return err
}

func (cm *CompletionManager) GenerateZshCompletion(w io.Writer, cmdName string) error {
	script := fmt.Sprintf(`#compdef %s

_%s() {
    local -a completions
    local curcontext="$curcontext" state line
    typeset -A opt_args

    local -a args
    local cur="${words[CURRENT]}"
    
    if (( CURRENT > 2 )); then
        args=("${words[@]:1:$((CURRENT-2))}")
    fi

    local output
    output=$(%s __complete --format=zsh "${args[@]}" "$cur" 2>/dev/null)
    
    if [[ $? -eq 0 ]]; then
        completions=(${(f)output})
        _describe 'completions' completions
    fi
}

_%s "$@"
`, cmdName, cmdName, cmdName, cmdName)

	_, err := w.Write([]byte(script))
	return err
}

func (cm *CompletionManager) GenerateFishCompletion(w io.Writer, cmdName string) error {
	script := fmt.Sprintf(`# fish completion for %s
function __%s_complete
    set -l cmd (commandline -opc)
    set -l cur (commandline -ct)
    set -e cmd[1]
    %s __complete --format=fish $cmd $cur 2>/dev/null
end

complete -c %s -f -a "(__%s_complete)"
`, cmdName, cmdName, cmdName, cmdName, cmdName)

	_, err := w.Write([]byte(script))
	return err
}
