package completion

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/Rehtt/Kit/cli"
)

func TestCommandCompletion(t *testing.T) {
	// 创建测试 CLI
	root := cli.NewCLI("test", "test app")
	sub1 := cli.NewCLI("command1", "first command")
	sub2 := cli.NewCLI("command2", "second command")
	hidden := cli.NewCLI("hidden", "hidden command")
	hidden.Hidden = true

	root.AddCommand(sub1, sub2, hidden)

	completion := NewCommandCompletion(root)

	tests := []struct {
		name       string
		toComplete string
		expected   []string
	}{
		{
			name:       "empty completion",
			toComplete: "",
			expected:   []string{"command1", "command2"},
		},
		{
			name:       "prefix match",
			toComplete: "command",
			expected:   []string{"command1", "command2"},
		},
		{
			name:       "specific match",
			toComplete: "command1",
			expected:   []string{"command1"},
		},
		{
			name:       "no match",
			toComplete: "xyz",
			expected:   []string{}, // 空切片而不是 nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := completion.Complete([]string{}, tt.toComplete)
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCommandCompletionWithDesc(t *testing.T) {
	root := cli.NewCLI("test", "test app")
	sub1 := cli.NewCLI("command1", "first command")
	sub2 := cli.NewCLI("command2", "second command")
	root.AddCommand(sub1, sub2)

	completion := NewCommandCompletion(root)
	result := completion.CompleteWithDesc([]string{}, "command")

	expected := []CompletionItem{
		{Value: "command1", Description: "first command"},
		{Value: "command2", Description: "second command"},
	}

	if len(result) != len(expected) {
		t.Errorf("expected %d items, got %d", len(expected), len(result))
		return
	}

	// 排序以确保比较一致性
	sort.Slice(result, func(i, j int) bool { return result[i].Value < result[j].Value })
	sort.Slice(expected, func(i, j int) bool { return expected[i].Value < expected[j].Value })

	for i, exp := range expected {
		if result[i].Value != exp.Value || result[i].Description != exp.Description {
			t.Errorf("expected %+v, got %+v", exp, result[i])
		}
	}
}

func TestFlagCompletion(t *testing.T) {
	root := cli.NewCLI("test", "test app")
	root.String("config", "", "config file")
	root.StringVarShortLong(new(string), "v", "verbose", "", "verbose mode")

	completion := NewFlagCompletion(root.FlagSet)

	tests := []struct {
		name       string
		toComplete string
		expected   []string
	}{
		{
			name:       "all flags",
			toComplete: "--",
			expected:   []string{"--config", "--verbose"},
		},
		{
			name:       "short flags",
			toComplete: "-v",
			expected:   []string{"-v"},
		},
		{
			name:       "specific flag",
			toComplete: "--con",
			expected:   []string{"--config"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := completion.Complete([]string{}, tt.toComplete)
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFileCompletion(t *testing.T) {
	// 创建临时目录和文件进行测试
	tmpDir := t.TempDir()

	// 创建测试文件
	testFiles := []string{"test.json", "test.yaml", "test.txt", "subdir"}
	for _, file := range testFiles[:3] {
		f, err := os.Create(filepath.Join(tmpDir, file))
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}

	// 创建子目录
	err := os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	// 切换到临时目录
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	tests := []struct {
		name       string
		extensions []string
		toComplete string
		wantCount  int
	}{
		{
			name:       "all files",
			extensions: []string{},
			toComplete: "",
			wantCount:  4, // 3 files + 1 directory
		},
		{
			name:       "json files only",
			extensions: []string{".json"},
			toComplete: "",
			wantCount:  2, // 1 json file + 1 directory
		},
		{
			name:       "prefix match",
			extensions: []string{},
			toComplete: "test",
			wantCount:  3, // 3 files matching "test" prefix
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completion := NewFileCompletion(tt.extensions...)
			result := completion.Complete([]string{}, tt.toComplete)

			if len(result) != tt.wantCount {
				t.Errorf("expected %d results, got %d: %v", tt.wantCount, len(result), result)
			}
		})
	}
}

func TestDirectoryCompletion(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建测试文件和目录
	os.Create(filepath.Join(tmpDir, "file.txt"))
	os.Mkdir(filepath.Join(tmpDir, "dir1"), 0755)
	os.Mkdir(filepath.Join(tmpDir, "dir2"), 0755)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	completion := NewDirectoryCompletion()
	result := completion.Complete([]string{}, "")

	// 应该只返回目录，不包含文件
	expectedCount := 2 // dir1/, dir2/
	if len(result) != expectedCount {
		t.Errorf("expected %d directories, got %d: %v", expectedCount, len(result), result)
	}

	// 检查所有结果都以 / 结尾（表示目录）
	for _, item := range result {
		if !filepath.IsAbs(item) && !strings.HasSuffix(item, "/") {
			t.Errorf("directory completion should end with '/': %s", item)
		}
	}
}

func TestCustomCompletion(t *testing.T) {
	// 测试简单函数
	simpleFunc := func(s string) []string {
		return []string{"option1", "option2"}
	}

	completion := NewCustomCompletion(simpleFunc)
	result := completion.Complete([]string{}, "opt")
	expected := []string{"option1", "option2"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// 测试带描述的函数
	descFunc := func(s string) []CompletionItem {
		return []CompletionItem{
			{Value: "item1", Description: "first item"},
			{Value: "item2", Description: "second item"},
		}
	}

	completion2 := NewCustomCompletion(descFunc)
	resultWithDesc := completion2.CompleteWithDesc([]string{}, "item")
	expectedWithDesc := []CompletionItem{
		{Value: "item1", Description: "first item"},
		{Value: "item2", Description: "second item"},
	}

	if !reflect.DeepEqual(resultWithDesc, expectedWithDesc) {
		t.Errorf("expected %v, got %v", expectedWithDesc, resultWithDesc)
	}
}

func TestCompletionTypes(t *testing.T) {
	tests := []struct {
		completion Completion
		expected   CompletionType
	}{
		{NewCommandCompletion(cli.NewCLI("test", "")), CompletionTypeCommand},
		{NewFlagCompletion(cli.NewCLI("test", "").FlagSet), CompletionTypeFlag},
		{NewFileCompletion(), CompletionTypeFile},
		{NewDirectoryCompletion(), CompletionTypeDirectory},
		{NewCustomCompletion(func(string) []string { return nil }), CompletionTypeCustom},
	}

	for i, tt := range tests {
		t.Run("", func(t *testing.T) {
			if tt.completion.GetType() != tt.expected {
				t.Errorf("test %d: expected type %d, got %d", i, tt.expected, tt.completion.GetType())
			}
		})
	}
}
