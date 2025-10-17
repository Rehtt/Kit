package completion

import (
	"testing"

	"github.com/Rehtt/Kit/cli"
)

func TestNew(t *testing.T) {
	root := cli.NewCLI("testapp", "test application")

	// 调用 New 函数
	cm := New(root)

	// 验证返回的是 CompletionManager
	if cm == nil {
		t.Fatal("New() should return a CompletionManager")
	}

	// 验证 CLI 被正确设置
	if cm.cli != root {
		t.Error("CompletionManager should reference the root CLI")
	}

	// 验证基本字段被初始化
	if cm.completions == nil {
		t.Error("completions map should be initialized")
	}

	if cm.sub == nil {
		t.Error("sub map should be initialized")
	}

	// 验证 __complete 命令被添加
	if root.SubCommands == nil {
		t.Fatal("root should have subcommands after New()")
	}

	completeCmd, exists := root.SubCommands["__complete"]
	if !exists {
		t.Error("__complete command should be added to root")
	}

	// 验证 __complete 命令的属性
	if !completeCmd.Hidden {
		t.Error("__complete command should be hidden")
	}

	if !completeCmd.Raw {
		t.Error("__complete command should be raw")
	}

	if completeCmd.CommandFunc == nil {
		t.Error("__complete command should have a CommandFunc")
	}
}

func TestNewWithExistingSubcommands(t *testing.T) {
	root := cli.NewCLI("testapp", "test application")

	// 添加一些现有的子命令
	sub1 := cli.NewCLI("command1", "first command")
	sub2 := cli.NewCLI("command2", "second command")
	root.AddCommand(sub1, sub2)

	// 调用 New 函数
	cm := New(root)

	// 验证现有子命令仍然存在
	if _, exists := root.SubCommands["command1"]; !exists {
		t.Error("existing subcommand should still exist")
	}

	if _, exists := root.SubCommands["command2"]; !exists {
		t.Error("existing subcommand should still exist")
	}

	// 验证 __complete 命令被添加
	if _, exists := root.SubCommands["__complete"]; !exists {
		t.Error("__complete command should be added")
	}

	// 验证总的子命令数量
	expectedCount := 3 // command1, command2, __complete
	if len(root.SubCommands) != expectedCount {
		t.Errorf("expected %d subcommands, got %d", expectedCount, len(root.SubCommands))
	}

	// 验证 CompletionManager 被正确返回
	if cm == nil {
		t.Error("New() should return a CompletionManager")
	}
}

func TestNewIntegration(t *testing.T) {
	// 创建一个完整的 CLI 应用进行集成测试
	root := cli.NewCLI("myapp", "my application")
	root.String("config", "", "config file")

	build := cli.NewCLI("build", "build project")
	build.String("target", "", "build target")
	root.AddCommand(build)

	// 创建补全管理器
	cm := New(root)

	// 注册一些补全
	cm.RegisterFileCompletion(root, "config", ".json", ".yaml")
	cm.RegisterCustomCompletionPrefixMatches(build, "target", []string{"all", "frontend", "backend"})

	// 测试补全功能是否正常工作
	// 测试根命令的文件补全注册
	if cm.manualCompletions[root]["config"] == nil {
		t.Error("file completion should be registered for root command")
	}

	// 测试子命令的自定义补全注册
	if cm.manualCompletions[build]["target"] == nil {
		t.Error("custom completion should be registered for subcommand")
	}

	// 测试初始化后的补全功能
	cm.init()

	// 测试命令补全
	commands := cm.Complete([]string{}, "")
	found := false
	for _, cmd := range commands {
		if cmd == "build" {
			found = true
			break
		}
	}
	if !found {
		t.Error("build command should be in completion results")
	}

	// 测试子命令的补全
	subResult := cm.Complete([]string{"build", "--target"}, "front")
	if len(subResult) != 1 || subResult[0] != "frontend" {
		t.Errorf("expected [frontend], got %v", subResult)
	}
}
