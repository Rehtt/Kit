package cli

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
)

func compareSlices(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if a == nil || b == nil {
		return len(a) == len(b)
	}
	return reflect.DeepEqual(a, b)
}

func TestCommandCompletion(t *testing.T) {
	root := NewCLI("app", "测试应用")
	root.AddCommand(
		NewCLI("hello", "问候命令"),
		NewCLI("world", "世界命令"),
		NewCLI("help", "帮助命令"),
	)

	completion := NewCommandCompletion(root)

	tests := []struct {
		name       string
		toComplete string
		expected   []string
	}{
		{"完整匹配", "hello", []string{"hello"}},
		{"前缀匹配", "h", []string{"hello", "help"}},
		{"无匹配", "xyz", nil},
		{"空输入", "", []string{"hello", "help", "world"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := completion.Complete([]string{}, tt.toComplete)
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !compareSlices(result, tt.expected) {
				t.Errorf("期望 %v，得到 %v", tt.expected, result)
			}
		})
	}
}

func TestFlagCompletion(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
	var config string
	var port int
	var verbose bool
	fs.StringVarShortLong(&config, "c", "config", "", "配置文件")
	fs.IntVarShortLong(&port, "p", "port", 8080, "端口号")
	fs.BoolVarShortLong(&verbose, "v", "verbose", false, "详细输出")

	completion := NewFlagCompletion(fs)

	tests := []struct {
		name       string
		toComplete string
		expected   []string
	}{
		{"长参数前缀", "--c", []string{"--config"}},
		{"短参数前缀", "-", []string{"--config", "--port", "--verbose", "-c", "-p", "-v"}},
		{"长参数完整", "--", []string{"--config", "--port", "--verbose"}},
		{"特定短参数", "-v", []string{"-v"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := completion.Complete([]string{}, tt.toComplete)
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !compareSlices(result, tt.expected) {
				t.Errorf("期望 %v，得到 %v", tt.expected, result)
			}
		})
	}
}

func TestFileCompletion(t *testing.T) {
	tmpDir := t.TempDir()
	testFiles := []string{"test.txt", "test.go", "example.json", "subdir/nested.txt"}

	for _, file := range testFiles {
		fullPath := filepath.Join(tmpDir, file)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	completion := NewFileCompletion()

	tests := []struct {
		name        string
		toComplete  string
		contains    []string
		notContains []string
	}{
		{"所有文件", "", []string{"test.txt", "test.go", "example.json", "subdir/"}, nil},
		{"txt文件前缀", "test", []string{"test.txt", "test.go"}, []string{"example.json"}},
		{"子目录", "sub", []string{"subdir/"}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := completion.Complete([]string{}, tt.toComplete)

			for _, item := range tt.contains {
				if !containsString(result, item) {
					t.Errorf("期望包含 %s，结果: %v", item, result)
				}
			}

			for _, item := range tt.notContains {
				if containsString(result, item) {
					t.Errorf("不应包含 %s，结果: %v", item, result)
				}
			}
		})
	}
}

func containsString(list []string, target string) bool {
	for _, s := range list {
		if strings.Contains(s, target) {
			return true
		}
	}
	return false
}

func TestFileCompletionWithExtensions(t *testing.T) {
	tmpDir := t.TempDir()
	testFiles := []string{"test.go", "test.txt", "example.json", "script.sh"}

	for _, file := range testFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, file), []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	result := NewFileCompletion(".go", ".txt").Complete([]string{}, "")

	for _, expected := range []string{"test.go", "test.txt"} {
		if !containsString(result, expected) {
			t.Errorf("期望包含 %s，结果: %v", expected, result)
		}
	}

	for _, notExpected := range []string{"example.json", "script.sh"} {
		if containsString(result, notExpected) {
			t.Errorf("不应包含 %s，结果: %v", notExpected, result)
		}
	}
}

func TestDirectoryCompletion(t *testing.T) {
	tmpDir := t.TempDir()
	dirs := []string{"dir1", "dir2", "subdir"}
	files := []string{"file1.txt", "file2.go"}

	for _, dir := range dirs {
		os.MkdirAll(filepath.Join(tmpDir, dir), 0o755)
	}
	for _, file := range files {
		os.WriteFile(filepath.Join(tmpDir, file), []byte("test"), 0o644)
	}

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	result := NewDirectoryCompletion().Complete([]string{}, "")

	for _, dir := range dirs {
		if !containsString(result, dir+"/") {
			t.Errorf("期望包含目录 %s/，结果: %v", dir, result)
		}
	}

	for _, file := range files {
		for _, r := range result {
			if strings.Contains(r, file) && !strings.HasSuffix(r, "/") {
				t.Errorf("不应包含文件 %s，结果: %v", file, result)
			}
		}
	}
}

func TestCustomCompletion(t *testing.T) {
	completion := NewCustomCompletion(func(toComplete string) []string {
		options := []string{"apple", "banana", "cherry", "date"}
		var matches []string
		for _, opt := range options {
			if strings.HasPrefix(opt, toComplete) {
				matches = append(matches, opt)
			}
		}
		return matches
	})

	tests := []struct {
		name       string
		toComplete string
		expected   []string
	}{
		{"前缀匹配", "a", []string{"apple"}},
		{"多个匹配", "", []string{"apple", "banana", "cherry", "date"}},
		{"无匹配", "xyz", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := completion.Complete([]string{}, tt.toComplete)
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !compareSlices(result, tt.expected) {
				t.Errorf("期望 %v，得到 %v", tt.expected, result)
			}
		})
	}
}

func TestCompletionManager(t *testing.T) {
	cli := NewCLI("test", "测试CLI")

	var config string
	var output string
	cli.StringVarShortLong(&config, "c", "config", "", "配置文件")
	cli.StringVarShortLong(&output, "o", "output", "", "输出目录")

	cli.RegisterFileCompletion("config", ".json", ".yaml")
	cli.RegisterDirectoryCompletion("output")

	tests := []struct {
		name       string
		args       []string
		toComplete string
		checkFunc  func([]string) bool
	}{
		{
			name:       "参数补全",
			args:       []string{},
			toComplete: "--c",
			checkFunc: func(results []string) bool {
				return slices.Contains(results, "--config")
			},
		},
		{
			name:       "短参数补全",
			args:       []string{},
			toComplete: "-",
			checkFunc: func(results []string) bool {
				hasC := false
				hasO := false
				for _, r := range results {
					if r == "-c" {
						hasC = true
					}
					if r == "-o" {
						hasO = true
					}
				}
				return hasC && hasO
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cli.CompletionManager.Complete(tt.args, tt.toComplete)
			if !tt.checkFunc(result) {
				t.Errorf("补全结果不符合预期: %v", result)
			}
		})
	}
}

func TestCompletionManagerWithShortFlag(t *testing.T) {
	tmpDir := t.TempDir()

	testFiles := []string{
		"config.json",
		"config.yaml",
		"data.txt",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tmpDir, file)
		if err := os.WriteFile(fullPath, []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	cli := NewCLI("test", "测试CLI")

	var config string
	var output string
	cli.StringVarShortLong(&config, "c", "config", "", "配置文件")
	cli.StringVarShortLong(&output, "o", "output", "", "输出目录")

	cli.RegisterFileCompletion("config", ".json", ".yaml")
	cli.RegisterDirectoryCompletion("output")

	tests := []struct {
		name       string
		args       []string
		toComplete string
		checkFunc  func([]string) bool
	}{
		{
			name:       "短参数文件补全",
			args:       []string{"-c"},
			toComplete: "",
			checkFunc: func(results []string) bool {
				hasJson := false
				hasYaml := false
				hasTxt := false
				for _, r := range results {
					if strings.Contains(r, ".json") {
						hasJson = true
					}
					if strings.Contains(r, ".yaml") {
						hasYaml = true
					}
					if strings.Contains(r, ".txt") {
						hasTxt = true
					}
				}
				// 应该有 json 和 yaml，但不应该有 txt
				return hasJson && hasYaml && !hasTxt
			},
		},
		{
			name:       "长参数文件补全",
			args:       []string{"--config"},
			toComplete: "",
			checkFunc: func(results []string) bool {
				hasJson := false
				hasYaml := false
				hasTxt := false
				for _, r := range results {
					if strings.Contains(r, ".json") {
						hasJson = true
					}
					if strings.Contains(r, ".yaml") {
						hasYaml = true
					}
					if strings.Contains(r, ".txt") {
						hasTxt = true
					}
				}
				return hasJson && hasYaml && !hasTxt
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cli.CompletionManager.Complete(tt.args, tt.toComplete)
			if !tt.checkFunc(result) {
				t.Errorf("补全结果不符合预期: %v", result)
			}
		})
	}
}

func TestShellCompletionGeneration(t *testing.T) {
	cli := NewCLI("myapp", "测试应用")

	shells := []string{"bash", "zsh", "fish"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			var buf bytes.Buffer
			var err error

			switch shell {
			case "bash":
				err = cli.CompletionManager.GenerateBashCompletion(&buf, "myapp")
			case "zsh":
				err = cli.CompletionManager.GenerateZshCompletion(&buf, "myapp")
			case "fish":
				err = cli.CompletionManager.GenerateFishCompletion(&buf, "myapp")
			}

			if err != nil {
				t.Errorf("生成 %s 补全脚本失败: %v", shell, err)
			}

			script := buf.String()
			if script == "" {
				t.Errorf("%s 补全脚本为空", shell)
			}

			// 检查脚本是否包含应用名称
			if !strings.Contains(script, "myapp") {
				t.Errorf("%s 补全脚本不包含应用名称", shell)
			}

			t.Logf("%s 补全脚本:\n%s", shell, script)
		})
	}
}

func TestCLICompletionIntegration(t *testing.T) {
	root := NewCLI("app", "测试应用")

	var config string
	var verbose bool
	root.StringVarShortLong(&config, "c", "config", "", "配置文件")
	root.BoolVarShortLong(&verbose, "v", "verbose", false, "详细输出")

	hello := NewCLI("hello", "问候命令")
	var name string
	hello.StringVar(&name, "name", "world", "名字")

	root.AddCommand(hello)
	root.RegisterFileCompletion("config", ".json")
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "参数补全",
			args: []string{"__complete", "--c"},
		},
		{
			name: "命令补全",
			args: []string{"__complete", "h"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := root.Parse(tt.args)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if err != nil {
				t.Errorf("补全处理失败: %v", err)
			}

			if output == "" {
				t.Errorf("补全输出为空")
			}

			t.Logf("补全输出: %s", output)
		})
	}
}

func TestCompletionWithDescription(t *testing.T) {
	root := NewCLI("test", "测试应用")
	root.AddCommand(NewCLI("hello", "问候命令"), NewCLI("world", "世界命令"))

	items := NewCommandCompletion(root).CompleteWithDesc([]string{}, "")

	if len(items) != 2 {
		t.Errorf("期望 2 个补全项，得到 %d 个", len(items))
	}

	for _, item := range items {
		if item.Value == "hello" && item.Description != "问候命令" {
			t.Errorf("hello 描述错误: %s", item.Description)
		}
		if item.Value == "world" && item.Description != "世界命令" {
			t.Errorf("world 描述错误: %s", item.Description)
		}
	}
}

func TestFlagCompletionWithDescription(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
	var config string
	var verbose bool
	fs.StringVarShortLong(&config, "c", "config", "", "配置文件路径")
	fs.BoolVarShortLong(&verbose, "v", "verbose", false, "详细输出模式")

	items := NewFlagCompletion(fs).CompleteWithDesc([]string{}, "--")

	if len(items) < 2 {
		t.Errorf("期望至少 2 个补全项，得到 %d 个", len(items))
	}

	foundConfig, foundVerbose := false, false
	for _, item := range items {
		if item.Value == "--config" {
			foundConfig = true
			if item.Description != "配置文件路径" {
				t.Errorf("--config 描述错误: %s", item.Description)
			}
		}
		if item.Value == "--verbose" {
			foundVerbose = true
			if item.Description != "详细输出模式" {
				t.Errorf("--verbose 描述错误: %s", item.Description)
			}
		}
	}

	if !foundConfig || !foundVerbose {
		t.Error("未找到预期的参数")
	}
}

func TestCustomCompletionWithDescription(t *testing.T) {
	completion := NewCustomCompletion(func(toComplete string) []CompletionItem {
		envs := []CompletionItem{
			{Value: "dev", Description: "开发环境"},
			{Value: "test", Description: "测试环境"},
			{Value: "prod", Description: "生产环境"},
		}
		var matches []CompletionItem
		for _, e := range envs {
			if strings.HasPrefix(e.Value, toComplete) {
				matches = append(matches, e)
			}
		}
		return matches
	})

	items := completion.CompleteWithDesc([]string{}, "")
	if len(items) != 3 {
		t.Errorf("期望 3 个补全项，得到 %d 个", len(items))
	}

	items = completion.CompleteWithDesc([]string{}, "d")
	if len(items) != 1 || items[0].Value != "dev" || items[0].Description != "开发环境" {
		t.Errorf("期望 dev:开发环境，得到 %v", items)
	}
}

func TestCompletionManagerWithDescription(t *testing.T) {
	cli := NewCLI("test", "测试CLI")

	var env string
	cli.StringVarShortLong(&env, "e", "env", "dev", "环境")

	cli.RegisterCustomCompletion("env", func(toComplete string) []CompletionItem {
		envs := []CompletionItem{
			{Value: "dev", Description: "开发环境"},
			{Value: "prod", Description: "生产环境"},
		}
		var matches []CompletionItem
		for _, e := range envs {
			if strings.HasPrefix(e.Value, toComplete) {
				matches = append(matches, e)
			}
		}
		return matches
	})

	// 测试带描述的补全
	items := cli.CompletionManager.CompleteWithDesc([]string{"-e"}, "")
	if len(items) != 2 {
		t.Errorf("期望 2 个补全项，但得到 %d 个", len(items))
	}

	for _, item := range items {
		if item.Value == "dev" && item.Description != "开发环境" {
			t.Errorf("dev 描述应该是 '开发环境'，但得到 '%s'", item.Description)
		}
		if item.Value == "prod" && item.Description != "生产环境" {
			t.Errorf("prod 描述应该是 '生产环境'，但得到 '%s'", item.Description)
		}
	}
}

func TestUnifiedCustomCompletion(t *testing.T) {
	t.Run("简单函数类型", func(t *testing.T) {
		completion := NewCustomCompletion(func(toComplete string) []string {
			var matches []string
			for _, opt := range []string{"apple", "banana", "cherry"} {
				if strings.HasPrefix(opt, toComplete) {
					matches = append(matches, opt)
				}
			}
			return matches
		})

		result := completion.Complete([]string{}, "a")
		if len(result) != 1 || result[0] != "apple" {
			t.Errorf("期望 [apple]，得到 %v", result)
		}

		items := completion.CompleteWithDesc([]string{}, "a")
		if len(items) != 1 || items[0].Value != "apple" || items[0].Description != "" {
			t.Errorf("期望 [{apple, ''}]，得到 %v", items)
		}
	})

	t.Run("带描述函数类型", func(t *testing.T) {
		completion := NewCustomCompletion(func(toComplete string) []CompletionItem {
			var matches []CompletionItem
			for _, opt := range []CompletionItem{
				{Value: "dev", Description: "开发环境"},
				{Value: "prod", Description: "生产环境"},
			} {
				if strings.HasPrefix(opt.Value, toComplete) {
					matches = append(matches, opt)
				}
			}
			return matches
		})

		result := completion.Complete([]string{}, "d")
		if len(result) != 1 || result[0] != "dev" {
			t.Errorf("期望 [dev]，得到 %v", result)
		}

		items := completion.CompleteWithDesc([]string{}, "d")
		if len(items) != 1 || items[0].Value != "dev" || items[0].Description != "开发环境" {
			t.Errorf("期望 [{dev, 开发环境}]，得到 %v", items)
		}
	})

	t.Run("CLI统一接口", func(t *testing.T) {
		cli := NewCLI("test", "测试")

		var simple string
		var withDesc string

		cli.StringVar(&simple, "simple", "", "简单补全")
		cli.StringVar(&withDesc, "desc", "", "带描述补全")

		// 注册简单补全
		cli.RegisterCustomCompletion("simple", func(toComplete string) []string {
			return []string{"option1", "option2"}
		})

		// 注册带描述补全
		cli.RegisterCustomCompletion("desc", func(toComplete string) []CompletionItem {
			return []CompletionItem{
				{Value: "opt1", Description: "选项1"},
				{Value: "opt2", Description: "选项2"},
			}
		})

		// 测试简单补全
		simpleResult := cli.CompletionManager.Complete([]string{"-simple"}, "")
		if len(simpleResult) != 2 {
			t.Errorf("简单补全期望 2 个结果，但得到 %d 个", len(simpleResult))
		}

		// 测试带描述补全
		descResult := cli.CompletionManager.CompleteWithDesc([]string{"-desc"}, "")
		if len(descResult) != 2 || descResult[0].Description == "" {
			t.Errorf("带描述补全期望 2 个带描述的结果，但得到 %v", descResult)
		}
	})
}

func TestSubCommandCompletion(t *testing.T) {
	// 创建根命令
	root := NewCLI("app", "测试应用")

	var rootEnv string
	root.StringVarShortLong(&rootEnv, "e", "env", "dev", "环境")
	root.RegisterCustomCompletion("env", func(toComplete string) []CompletionItem {
		return []CompletionItem{
			{Value: "dev", Description: "开发环境"},
			{Value: "prod", Description: "生产环境"},
		}
	})

	// 创建子命令
	build := NewCLI("build", "构建项目")
	var target string
	build.StringVar(&target, "target", "all", "构建目标")
	build.RegisterCustomCompletion("target", func(toComplete string) []CompletionItem {
		return []CompletionItem{
			{Value: "all", Description: "构建所有组件"},
			{Value: "frontend", Description: "只构建前端"},
			{Value: "backend", Description: "只构建后端"},
		}
	})

	root.AddCommand(build)

	tests := []struct {
		name     string
		args     []string
		complete string
		expected []string
		desc     bool
	}{
		{
			name:     "根命令环境补全",
			args:     []string{"-e"},
			complete: "",
			expected: []string{"dev", "prod"},
			desc:     false,
		},
		{
			name:     "子命令目标补全",
			args:     []string{"build", "--target"},
			complete: "",
			expected: []string{"all", "frontend", "backend"},
			desc:     false,
		},
		{
			name:     "子命令目标补全带描述",
			args:     []string{"build", "--target"},
			complete: "",
			expected: []string{"all", "frontend", "backend"},
			desc:     true,
		},
		{
			name:     "子命令参数补全",
			args:     []string{"build"},
			complete: "--",
			expected: []string{"--target"},
			desc:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []string
			var items []CompletionItem

			if tt.desc {
				items = root.CompletionManager.CompleteWithDesc(tt.args, tt.complete)
				result = make([]string, len(items))
				for i, item := range items {
					result[i] = item.Value
				}
			} else {
				result = root.CompletionManager.Complete(tt.args, tt.complete)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("期望 %d 个结果，但得到 %d 个: %v", len(tt.expected), len(result), result)
				return
			}

			for _, expected := range tt.expected {
				if !slices.Contains(result, expected) {
					t.Errorf("期望找到 %s，但结果中没有: %v", expected, result)
				}
			}

			// 如果测试带描述，验证描述不为空
			if tt.desc && len(items) > 0 {
				for _, item := range items {
					if item.Description == "" {
						t.Errorf("期望 %s 有描述，但描述为空", item.Value)
					}
				}
			}
		})
	}
}
