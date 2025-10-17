# Completion

CLI 命令行动态补全功能模块，支持 Bash、Zsh、Fish 等 Shell 的自动补全。

## 功能特性

- **命令补全** - 子命令自动补全
- **参数补全** - 长短参数名补全  
- **文件补全** - 文件路径和扩展名过滤
- **目录补全** - 目录路径补全
- **自定义补全** - 支持自定义补全逻辑
- **描述支持** - Zsh/Fish 显示补全项描述

## 快速开始

```go
import (
    "github.com/Rehtt/Kit/cli"
    "github.com/Rehtt/Kit/cli/completion"
)

root := cli.NewCLI("app", "我的应用")

// 定义带 FlagItem 的参数（自动生成补全）
root.FlagSet.String("config", "", "配置文件", 
    cli.NewFlagItemFile())
    
root.FlagSet.String("format", "json", "输出格式", 
    cli.NewFlagItemSelectString("json", "yaml", "xml"))
    
root.FlagSet.String("dir", ".", "工作目录", 
    cli.NewFlagItemDir())

// 创建补全管理器
cm := completion.New(root)

// 生成补全脚本
cm.GenerateCompletion("bash")  // 或 zsh、fish
```

## FlagItem 类型

```go
// 文件补全
cli.NewFlagItemFile()

// 目录补全
cli.NewFlagItemDir()

// 选项补全（带描述）
cli.NewFlagItemSelect(
    cli.FlagItemNode{Value: "dev", Description: "开发环境"},
    cli.FlagItemNode{Value: "prod", Description: "生产环境"},
)

// 选项补全（简单）
cli.NewFlagItemSelectString("option1", "option2", "option3")
```

## 手动覆盖补全

```go
cm := completion.New(root)

// 文件补全（指定扩展名）
cm.RegisterFileCompletion(root, "config", ".json", ".yaml")

// 自定义补全
cm.RegisterCustomCompletion(root, "branch", func(toComplete string) []string {
    return gitBranches()
})

// 前缀匹配补全
cm.RegisterCustomCompletionPrefixMatches(root, "env", []completion.CompletionItem{
    {Value: "dev", Description: "开发环境"},
    {Value: "prod", Description: "生产环境"},
})
```

## Shell 集成

```bash
# Bash
app completion bash > /etc/bash_completion.d/app

# Zsh
app completion zsh > "${fpath[1]}/_app"

# Fish
app completion fish > ~/.config/fish/completions/app.fish
```

## API

### CompletionManager

- `New(root)` - 创建管理器
- `RegisterFileCompletion(cli, flag, exts...)` - 文件补全
- `RegisterDirectoryCompletion(cli, flag)` - 目录补全
- `RegisterCustomCompletion(cli, flag, fn)` - 自定义补全
- `RegisterCustomCompletionPrefixMatches(cli, flag, items)` - 前缀匹配
- `GenerateCompletion(shell, cmdName...)` - 生成脚本

### 自定义函数签名

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
