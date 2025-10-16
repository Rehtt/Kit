package completion

import "github.com/Rehtt/Kit/cli"

// CompletionType 补全类型
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

// Completion 补全接口
type Completion interface {
	Complete(args []string, toComplete string) []string
	CompleteWithDesc(args []string, toComplete string) []CompletionItem
	GetType() CompletionType
}

// CompletionFunc 自定义补全函数
type CompletionFunc any

// cliPtr CLI 指针类型
type cliPtr *cli.CLI

// normalizeCompletionFunc 标准化补全函数
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
