# Completion

CLI 命令行补全功能模块，支持 Bash、Zsh、Fish 等 Shell 的自动补全。

## 功能特性

- **命令补全** - 子命令自动补全
- **参数补全** - 长短参数名补全  
- **文件补全** - 文件路径和扩展名过滤
- **目录补全** - 目录路径补全
- **自定义补全** - 支持自定义补全逻辑
- **描述支持** - Zsh/Fish 显示补全项描述

## 快速开始

```go
import "github.com/Rehtt/Kit/cli/completion"

// 创建补全管理器
root := cli.NewCLI("myapp", "示例应用")
cm := completion.New(root)

// 注册补全
cm.RegisterFileCompletion(root, "config", ".json", ".yaml")
cm.RegisterCustomCompletionPrefixMatches(root, "env", []completion.CompletionItem{
    {Value: "dev", Description: "开发环境"},
    {Value: "prod", Description: "生产环境"},
})

// 生成补全脚本
completion := cli.NewCLI("completion", "生成补全脚本")
completion.CommandFunc = func(args []string) error {
    return cm.GenerateCompletion(args[0]) // bash/zsh/fish
}
```

## API 参考

### 核心类型

```go
type CompletionItem struct {
    Value       string  // 补全值
    Description string  // 描述信息
}

type CompletionManager struct {
    // 补全管理器
}
```

### 主要方法

- `New(root *cli.CLI) *CompletionManager` - 创建补全管理器
- `RegisterFileCompletion(cli, flag, exts...)` - 注册文件补全
- `RegisterDirectoryCompletion(cli, flag)` - 注册目录补全
- `RegisterCustomCompletion(cli, flag, func)` - 注册自定义补全
- `RegisterCustomCompletionPrefixMatches(cli, flag, items)` - 注册前缀匹配补全
- `GenerateCompletion(shell)` - 生成 Shell 补全脚本

### 自定义补全函数

支持两种函数签名：

```go
// 简单补全
func(toComplete string) []string

// 带描述补全
func(toComplete string) []CompletionItem
```

## 文件结构

```
completion/
├── completion.go      # 主入口
├── types.go          # 基础类型定义
├── implementations.go # 补全实现
├── manager.go        # 补全管理器
└── shell_generators.go # Shell 脚本生成器
```

## 使用示例

参见 `../example/example_completion.go`
