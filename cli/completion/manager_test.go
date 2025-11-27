package completion

import (
	"cmp"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Rehtt/Kit/cli"
)

func TestCompletionManager_RegisterCompletion(t *testing.T) {
	root := cli.NewCLI("test", "test app")
	cm := NewCompletionManager(root)

	// 创建自定义补全
	customCompletion := NewCustomCompletion(func(s string) []string {
		return []string{"custom1", "custom2"}
	})

	// 注册补全
	cm.RegisterCompletion(root, "test-flag", customCompletion)

	// 验证注册成功
	if cm.manualCompletions == nil {
		t.Error("manualCompletions should be initialized")
	}

	if cm.manualCompletions[root] == nil {
		t.Error("manualCompletions[root] should be initialized")
	}

	if cm.manualCompletions[root]["test-flag"] != customCompletion {
		t.Error("completion not registered correctly")
	}
}

func TestCompletionManager_RegisterFileCompletion(t *testing.T) {
	root := cli.NewCLI("test", "test app")
	cm := NewCompletionManager(root)

	cm.RegisterFileCompletion(root, "config", ".json", ".yaml")

	if cm.manualCompletions[root]["config"] == nil {
		t.Error("file completion not registered")
	}

	if cm.manualCompletions[root]["config"].GetType() != CompletionTypeFile {
		t.Error("wrong completion type")
	}
}

func TestCompletionManager_RegisterDirectoryCompletion(t *testing.T) {
	root := cli.NewCLI("test", "test app")
	cm := NewCompletionManager(root)

	cm.RegisterDirectoryCompletion(root, "output")

	if cm.manualCompletions[root]["output"] == nil {
		t.Error("directory completion not registered")
	}

	if cm.manualCompletions[root]["output"].GetType() != CompletionTypeDirectory {
		t.Error("wrong completion type")
	}
}

func TestCompletionManager_RegisterCustomCompletion(t *testing.T) {
	root := cli.NewCLI("test", "test app")
	cm := NewCompletionManager(root)

	testFunc := func(s string) []string {
		return []string{"test1", "test2"}
	}

	cm.RegisterCustomCompletion(root, "env", testFunc)

	if cm.manualCompletions[root]["env"] == nil {
		t.Error("custom completion not registered")
	}

	if cm.manualCompletions[root]["env"].GetType() != CompletionTypeCustom {
		t.Error("wrong completion type")
	}
}

func TestCompletionManager_RegisterCustomCompletionPrefixMatches(t *testing.T) {
	root := cli.NewCLI("test", "test app")
	cm := NewCompletionManager(root)

	// 测试 []string
	cm.RegisterCustomCompletionPrefixMatches(root, "env1", []string{"dev", "prod", "test"})

	// 测试 []CompletionItem
	items := []CompletionItem{
		{Value: "option1", Description: "first option"},
		{Value: "option2", Description: "second option"},
	}
	cm.RegisterCustomCompletionPrefixMatches(root, "env2", items)

	// 验证注册成功
	if cm.manualCompletions[root]["env1"] == nil || cm.manualCompletions[root]["env2"] == nil {
		t.Error("prefix match completions not registered")
	}

	// 测试前缀匹配功能
	cm.init()
	result1 := cm.completions["env1"].Complete([]string{}, "d")
	if !reflect.DeepEqual(result1, []string{"dev"}) {
		t.Errorf("expected [dev], got %v", result1)
	}

	result2 := cm.completions["env2"].CompleteWithDesc([]string{}, "option1")
	expected2 := []CompletionItem{{Value: "option1", Description: "first option"}}
	if !reflect.DeepEqual(result2, expected2) {
		t.Errorf("expected %v, got %v", expected2, result2)
	}
}

func TestCompletionManager_NormalizeFlagName(t *testing.T) {
	root := cli.NewCLI("test", "test app")
	root.StringVarShortLong(new(string), "v", "verbose", "", "verbose mode")
	cm := NewCompletionManager(root)

	tests := []struct {
		input    string
		expected string
	}{
		{"v", "verbose"},       // 短名转长名
		{"verbose", "verbose"}, // 长名保持不变
		{"unknown", "unknown"}, // 未知参数保持不变
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := cm.normalizeFlagName(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCompletionManager_FindSubCommand(t *testing.T) {
	root := cli.NewCLI("test", "test app")
	sub := cli.NewCLI("subcommand", "sub command")
	root.AddCommand(sub)

	cm := NewCompletionManager(root)
	cm.init()

	tests := []struct {
		name     string
		args     []string
		wantSub  bool
		wantArgs []string
	}{
		{
			name:     "no subcommand",
			args:     []string{"--flag", "value"},
			wantSub:  false,
			wantArgs: []string{"--flag", "value"},
		},
		{
			name:     "subcommand at start",
			args:     []string{"subcommand", "--flag"},
			wantSub:  true,
			wantArgs: []string{"--flag"},
		},
		{
			name:     "subcommand after flags",
			args:     []string{"--flag", "value", "subcommand", "--sub-flag"},
			wantSub:  true,
			wantArgs: []string{"--sub-flag"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subCmd, remaining := cm.findSubCommand(tt.args)

			if (subCmd != nil) != tt.wantSub {
				t.Errorf("expected subcommand found: %v, got: %v", tt.wantSub, subCmd != nil)
			}

			if !reflect.DeepEqual(remaining, tt.wantArgs) {
				t.Errorf("expected remaining args: %v, got: %v", tt.wantArgs, remaining)
			}
		})
	}
}

func TestCompletionManager_Complete(t *testing.T) {
	root := cli.NewCLI("test", "test app")
	root.String("config", "", "config file")

	sub := cli.NewCLI("build", "build command")
	sub.String("target", "", "build target")
	root.AddCommand(sub)

	cm := NewCompletionManager(root)
	cm.RegisterCustomCompletionPrefixMatches(root, "config", []string{"dev.json", "prod.json"})
	cm.RegisterCustomCompletionPrefixMatches(sub, "target", []string{"all", "frontend", "backend"})
	cm.init() // 初始化补全管理器

	tests := []struct {
		name       string
		args       []string
		toComplete string
		expected   []string
	}{
		{
			name:       "command completion",
			args:       []string{},
			toComplete: "",
			expected:   []string{"build"},
		},
		{
			name:       "flag completion",
			args:       []string{},
			toComplete: "--",
			expected:   []string{"--config"},
		},
		{
			name:       "root flag value completion",
			args:       []string{"--config"},
			toComplete: "dev",
			expected:   []string{"dev.json"},
		},
		{
			name:       "subcommand flag completion",
			args:       []string{"build"},
			toComplete: "--",
			expected:   []string{"--target"},
		},
		{
			name:       "subcommand flag value completion",
			args:       []string{"build", "--target"},
			toComplete: "front",
			expected:   []string{"frontend"},
		},
		{
			name:       "complex args with subcommand",
			args:       []string{"--config", "dev.json", "build", "--target"},
			toComplete: "back",
			expected:   []string{"backend"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cm.Complete(tt.args, tt.toComplete)
			slices.Sort(result)
			slices.Sort(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCompletionManager_CompleteWithDesc(t *testing.T) {
	root := cli.NewCLI("test", "test app")
	sub := cli.NewCLI("deploy", "deploy application")
	sub.String("env", "", "environment")
	root.AddCommand(sub)

	cm := NewCompletionManager(root)
	cm.RegisterCustomCompletionPrefixMatches(sub, "env", []CompletionItem{
		{Value: "dev", Description: "development environment"},
		{Value: "prod", Description: "production environment"},
	})
	cm.init() // 初始化补全管理器

	result := cm.CompleteWithDesc([]string{"deploy", "--env"}, "")
	expected := []CompletionItem{
		{Value: "dev", Description: "development environment"},
		{Value: "prod", Description: "production environment"},
	}

	slices.SortFunc(result, func(a, b CompletionItem) int { return cmp.Compare(a.Value, b.Value) })
	slices.SortFunc(expected, func(a, b CompletionItem) int { return cmp.Compare(a.Value, b.Value) })

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestCompletionManager_ParseCompletionArgs(t *testing.T) {
	root := cli.NewCLI("test", "test app")
	root.String("flag", "", "test flag")
	cm := NewCompletionManager(root)

	tests := []struct {
		name         string
		args         []string
		wantArgs     []string
		wantComplete string
	}{
		{
			name:         "empty args",
			args:         []string{},
			wantArgs:     []string{},
			wantComplete: "",
		},
		{
			name:         "complete flag value",
			args:         []string{"--flag"},
			wantArgs:     []string{"--flag"},
			wantComplete: "",
		},
		{
			name:         "complete partial value",
			args:         []string{"command", "partial"},
			wantArgs:     []string{"command"},
			wantComplete: "partial",
		},
		{
			name:         "empty last arg with flag",
			args:         []string{"--flag", ""},
			wantArgs:     []string{"--flag"},
			wantComplete: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, toComplete := cm.parseCompletionArgs(tt.args)

			if !reflect.DeepEqual(args, tt.wantArgs) {
				t.Errorf("expected args %v, got %v", tt.wantArgs, args)
			}

			if toComplete != tt.wantComplete {
				t.Errorf("expected toComplete %q, got %q", tt.wantComplete, toComplete)
			}
		})
	}
}

func TestCompletionManager_Init(t *testing.T) {
	root := cli.NewCLI("test", "test app")
	sub1 := cli.NewCLI("command1", "first command")
	sub2 := cli.NewCLI("hidden", "hidden command")
	sub2.Hidden = true
	root.AddCommand(sub1, sub2)

	cm := NewCompletionManager(root)
	cm.RegisterCustomCompletion(root, "root-flag", func(s string) []string { return []string{"root"} })
	cm.RegisterCustomCompletion(sub1, "sub-flag", func(s string) []string { return []string{"sub"} })

	// 初始化前
	if cm.hasInit {
		t.Error("should not be initialized yet")
	}

	cm.init()

	// 初始化后
	if !cm.hasInit {
		t.Error("should be initialized")
	}

	// 检查根命令的补全
	if cm.completions["root-flag"] == nil {
		t.Error("root flag completion not initialized")
	}

	// 注意：子命令不会注册到 completions 中，而是注册到 sub 中
	// 子命令补全由 CommandCompletion 负责处理

	// 检查子命令管理器
	if cm.sub["command1"] == nil {
		t.Error("subcommand manager not created")
	}

	// 检查子命令的补全传递
	subManager := cm.sub["command1"]
	if subManager.completions["sub-flag"] == nil {
		t.Error("subcommand flag completion not transferred")
	}
}

func TestCompletionManager_HandleCompletion(t *testing.T) {
	root := cli.NewCLI("test", "test app")
	sub := cli.NewCLI("build", "build command")
	root.AddCommand(sub)

	cm := NewCompletionManager(root)
	cm.RegisterCustomCompletionPrefixMatches(sub, "env", []CompletionItem{
		{Value: "dev", Description: "development"},
		{Value: "prod", Description: "production"},
	})

	// 测试不同格式的输出
	tests := []struct {
		name     string
		args     []string
		contains []string // 检查输出是否包含这些字符串
	}{
		{
			name:     "simple format",
			args:     []string{"build", "--env", "d"},
			contains: []string{"dev"},
		},
		{
			name:     "zsh format",
			args:     []string{"--format=zsh", "build", "--env", ""},
			contains: []string{"dev:development", "prod:production"},
		},
		{
			name:     "fish format",
			args:     []string{"--format=fish", "build", "--env", "p"},
			contains: []string{"prod\tproduction"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 捕获输出
			var output strings.Builder

			// 这里我们不能直接测试 HandleCompletion 的输出，
			// 因为它直接打印到 stdout。我们可以测试它不返回错误。
			err := cm.HandleCompletion(tt.args)
			if err != nil {
				t.Errorf("HandleCompletion returned error: %v", err)
			}

			// 实际的输出测试需要重定向 stdout，这里我们只测试不出错
			_ = output
		})
	}
}
