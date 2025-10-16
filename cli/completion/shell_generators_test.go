package completion

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Rehtt/Kit/cli"
)

func TestCompletionManager_GenerateBashCompletion(t *testing.T) {
	root := cli.NewCLI("myapp", "test app")
	cm := NewCompletionManager(root)

	var buf bytes.Buffer
	err := cm.GenerateBashCompletion(&buf, "myapp")
	if err != nil {
		t.Fatalf("GenerateBashCompletion failed: %v", err)
	}

	output := buf.String()

	// 检查关键内容
	expectedContents := []string{
		"# bash completion for myapp",
		"_myapp_completion()",
		"myapp __complete",
		"complete -F _myapp_completion myapp",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(output, expected) {
			t.Errorf("bash completion script should contain: %s", expected)
		}
	}
}

func TestCompletionManager_GenerateZshCompletion(t *testing.T) {
	root := cli.NewCLI("testcmd", "test command")
	cm := NewCompletionManager(root)

	var buf bytes.Buffer
	err := cm.GenerateZshCompletion(&buf, "testcmd")
	if err != nil {
		t.Fatalf("GenerateZshCompletion failed: %v", err)
	}

	output := buf.String()

	expectedContents := []string{
		"#compdef testcmd",
		"_testcmd()",
		"testcmd __complete --format=zsh",
		"_testcmd \"$@\"",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(output, expected) {
			t.Errorf("zsh completion script should contain: %s", expected)
		}
	}
}

func TestCompletionManager_GenerateFishCompletion(t *testing.T) {
	root := cli.NewCLI("fishapp", "fish test app")
	cm := NewCompletionManager(root)

	var buf bytes.Buffer
	err := cm.GenerateFishCompletion(&buf, "fishapp")
	if err != nil {
		t.Fatalf("GenerateFishCompletion failed: %v", err)
	}

	output := buf.String()

	expectedContents := []string{
		"# fish completion for fishapp",
		"function __fishapp_complete",
		"fishapp __complete --format=fish",
		"complete -c fishapp -f -a \"(__fishapp_complete)\"",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(output, expected) {
			t.Errorf("fish completion script should contain: %s", expected)
		}
	}
}

func TestCompletionManager_GenerateCompletion(t *testing.T) {
	root := cli.NewCLI("testapp", "test application")
	cm := NewCompletionManager(root)

	tests := []struct {
		name    string
		shell   string
		wantErr bool
		cmdName string
	}{
		{
			name:    "bash completion",
			shell:   "bash",
			wantErr: false,
			cmdName: "testapp",
		},
		{
			name:    "zsh completion",
			shell:   "zsh",
			wantErr: false,
			cmdName: "testapp",
		},
		{
			name:    "fish completion",
			shell:   "fish",
			wantErr: false,
			cmdName: "testapp",
		},
		{
			name:    "unsupported shell",
			shell:   "powershell",
			wantErr: true,
			cmdName: "testapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重定向 stdout 到 buffer 进行测试
			var buf bytes.Buffer

			// 由于 GenerateCompletion 直接写入 os.Stdout，
			// 我们需要测试具体的生成方法
			var err error
			switch tt.shell {
			case "bash":
				err = cm.GenerateBashCompletion(&buf, tt.cmdName)
			case "zsh":
				err = cm.GenerateZshCompletion(&buf, tt.cmdName)
			case "fish":
				err = cm.GenerateFishCompletion(&buf, tt.cmdName)
			default:
				// 测试错误情况
				err = cm.GenerateCompletion(tt.shell, tt.cmdName)
			}

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// 检查输出不为空
				if buf.Len() == 0 && tt.shell != "powershell" {
					t.Error("expected non-empty output")
				}
			}
		})
	}
}

func TestShellCompletionScriptContent(t *testing.T) {
	root := cli.NewCLI("myapp", "my application")
	cm := NewCompletionManager(root)

	tests := []struct {
		name   string
		shell  string
		checks []string
	}{
		{
			name:  "bash script structure",
			shell: "bash",
			checks: []string{
				"COMPREPLY=()",
				"_init_completion",
				"compgen -W",
			},
		},
		{
			name:  "zsh script structure",
			shell: "zsh",
			checks: []string{
				"local -a completions",
				"_describe 'completions'",
				"${(f)output}",
			},
		},
		{
			name:  "fish script structure",
			shell: "fish",
			checks: []string{
				"commandline -opc",
				"commandline -ct",
				"set -e cmd[1]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var err error

			switch tt.shell {
			case "bash":
				err = cm.GenerateBashCompletion(&buf, "myapp")
			case "zsh":
				err = cm.GenerateZshCompletion(&buf, "myapp")
			case "fish":
				err = cm.GenerateFishCompletion(&buf, "myapp")
			}

			if err != nil {
				t.Fatalf("failed to generate %s completion: %v", tt.shell, err)
			}

			output := buf.String()
			for _, check := range tt.checks {
				if !strings.Contains(output, check) {
					t.Errorf("%s completion should contain: %s", tt.shell, check)
				}
			}
		})
	}
}

func TestCompletionScriptCustomization(t *testing.T) {
	root := cli.NewCLI("customapp", "custom application")
	cm := NewCompletionManager(root)

	// 测试自定义命令名
	var buf bytes.Buffer
	err := cm.GenerateBashCompletion(&buf, "custom-name")
	if err != nil {
		t.Fatalf("failed to generate completion: %v", err)
	}

	output := buf.String()

	// 检查自定义命令名是否正确使用
	if !strings.Contains(output, "custom-name __complete") {
		t.Error("completion script should use custom command name")
	}

	if !strings.Contains(output, "_custom-name_completion") {
		t.Error("completion function should use custom command name")
	}

	if !strings.Contains(output, "complete -F _custom-name_completion custom-name") {
		t.Error("complete command should use custom command name")
	}
}
