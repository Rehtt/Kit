package completion

import (
	"flag"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/Rehtt/Kit/cli"
)

// CommandCompletion 命令补全
type CommandCompletion struct {
	cli *cli.CLI
}

// NewCommandCompletion 创建命令补全
func NewCommandCompletion(cli *cli.CLI) *CommandCompletion {
	return &CommandCompletion{cli: cli}
}

func (c *CommandCompletion) Complete(args []string, toComplete string) []string {
	if c.cli.SubCommands == nil {
		return nil
	}

	suggestions := make([]string, 0, len(c.cli.SubCommands))
	for name, subCmd := range c.cli.SubCommands {
		if !subCmd.Hidden && strings.HasPrefix(name, toComplete) {
			suggestions = append(suggestions, name)
		}
	}

	sort.Strings(suggestions)
	return suggestions
}

func (c *CommandCompletion) CompleteWithDesc(args []string, toComplete string) []CompletionItem {
	if c.cli.SubCommands == nil {
		return nil
	}

	items := make([]CompletionItem, 0, len(c.cli.SubCommands))
	for name, subCmd := range c.cli.SubCommands {
		if !subCmd.Hidden && strings.HasPrefix(name, toComplete) {
			items = append(items, CompletionItem{
				Value:       name,
				Description: subCmd.Instruction,
			})
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

// FlagCompletion flag 补全
type FlagCompletion struct {
	flagSet *cli.FlagSet
}

// NewFlagCompletion 创建参数补全
func NewFlagCompletion(flagSet *cli.FlagSet) *FlagCompletion {
	return &FlagCompletion{flagSet: flagSet}
}

func (f *FlagCompletion) addFlagSuggestion(suggestions *[]string, flagName, toComplete string) {
	if strings.HasPrefix(flagName, toComplete) {
		*suggestions = append(*suggestions, flagName)
	}
}

func (f *FlagCompletion) addFlagItem(items *[]CompletionItem, flagName, description, toComplete string) {
	if strings.HasPrefix(flagName, toComplete) {
		*items = append(*items, CompletionItem{
			Value:       flagName,
			Description: description,
		})
	}
}

func (f *FlagCompletion) Complete(args []string, toComplete string) []string {
	suggestions := make([]string, 0, 10)
	processed := make(map[string]bool)

	f.flagSet.VisitAll(func(flag *flag.Flag) {
		if processed[flag.Name] {
			return
		}
		if f.flagSet.ShortLongMap != nil {
			if slValue, exists := f.flagSet.ShortLongMap[flag.Name]; exists {
				if slValue.ShortName != "" {
					f.addFlagSuggestion(&suggestions, "-"+slValue.ShortName, toComplete)
					processed[slValue.ShortName] = true
				}
				if slValue.LongName != "" {
					f.addFlagSuggestion(&suggestions, "--"+slValue.LongName, toComplete)
					processed[slValue.LongName] = true
				}
				processed[flag.Name] = true
				return
			}
		}
		f.addFlagSuggestion(&suggestions, "--"+flag.Name, toComplete)
		processed[flag.Name] = true
	})

	sort.Strings(suggestions)
	return suggestions
}

func (f *FlagCompletion) CompleteWithDesc(args []string, toComplete string) []CompletionItem {
	items := make([]CompletionItem, 0, 10)
	processed := make(map[string]bool)

	f.flagSet.VisitAll(func(flag *flag.Flag) {
		if processed[flag.Name] {
			return
		}
		if f.flagSet.ShortLongMap != nil {
			if slValue, exists := f.flagSet.ShortLongMap[flag.Name]; exists {
				if slValue.ShortName != "" {
					f.addFlagItem(&items, "-"+slValue.ShortName, flag.Usage, toComplete)
					processed[slValue.ShortName] = true
				}
				if slValue.LongName != "" {
					f.addFlagItem(&items, "--"+slValue.LongName, flag.Usage, toComplete)
					processed[slValue.LongName] = true
				}
				processed[flag.Name] = true
				return
			}
		}
		f.addFlagItem(&items, "--"+flag.Name, flag.Usage, toComplete)
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

// NewFileCompletion 创建文件补全
func NewFileCompletion(extensions ...string) *FileCompletion {
	return &FileCompletion{extensions: extensions, dirOnly: false}
}

// NewDirectoryCompletion 创建目录补全
func NewDirectoryCompletion() *FileCompletion {
	return &FileCompletion{dirOnly: true}
}

func (f *FileCompletion) shouldIncludeEntry(entry os.DirEntry, baseName string) bool {
	name := entry.Name()
	if strings.HasPrefix(name, ".") && !strings.HasPrefix(baseName, ".") {
		return false
	}
	if baseName != "" && !strings.HasPrefix(name, baseName) {
		return false
	}
	if entry.IsDir() {
		return true
	}
	if f.dirOnly {
		return false
	}
	if len(f.extensions) == 0 {
		return true
	}
	return slices.Contains(f.extensions, filepath.Ext(name))
}

func (f *FileCompletion) Complete(args []string, toComplete string) []string {
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
		return nil
	}

	suggestions := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !f.shouldIncludeEntry(entry, base) {
			continue
		}

		var fullPath string
		if dir == "." {
			fullPath = entry.Name()
		} else {
			fullPath = filepath.Join(dir, entry.Name())
		}
		if entry.IsDir() {
			fullPath += "/"
		}
		suggestions = append(suggestions, fullPath)
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

// NewCustomCompletion 创建自定义补全
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
