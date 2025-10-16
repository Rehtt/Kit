### `cli` —— 轻量级命令行工具集合

基于 Go 标准库 `flag` 的薄封装，提供子命令组织、长短参数支持和友好的帮助信息。

### 快速上手

```go
package main

import (
    "fmt"
    "os"
    cli "github.com/Rehtt/Kit/cli"
)

func main() {
    // 创建根命令
    root := cli.NewCLI("app", "示例应用")
    root.Usage = "[flags] [command]"
    
    // 定义参数
    var verbose bool
    root.BoolVar(&verbose, "v", false, "详细输出")

    // 子命令
    hello := cli.NewCLI("hello", "打印问候语")
    hello.Usage = "[flags]"
    name := hello.String("name", "world", "名字")
    hello.CommandFunc = func(args []string) error {
        if verbose {
            fmt.Println("verbose mode")
        }
        fmt.Printf("Hello, %s\n", *name)
        return nil
    }

    // 添加子命令
    _ = root.AddCommand(hello)

    // 运行
    if err := root.Run(os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(2)
    }
}
```

### 长短参数支持

#### 方式一：使用别名

```go
var config string
root.StringVar(&config, "config", "", "配置文件")
root.Alias("c", "config")  // -c 作为 --config 的别名
```

#### 方式二：ShortLong 方法

```go
var host string
var port int
var verbose bool

root.StringVarShortLong(&host, "h", "host", "localhost", "监听地址")
root.IntVarShortLong(&port, "p", "port", 8080, "监听端口")
root.BoolVarShortLong(&verbose, "v", "verbose", false, "详细输出")
```

支持混合使用：`app -h 127.0.0.1 --port 9000 -v`

#### 支持的类型

**原生类型**：String, Int, Int64, Uint, Uint64, Float64, Duration, Bool

**扩展类型**：Strings（字符串切片）, PasswordString（密码类型）

### 核心 API

- `NewCLI(use, instruction string) *CLI` - 创建命令
- `AddCommand(cli ...*CLI) error` - 添加子命令
- `Run(arguments []string) error` - 解析并执行
- `Alias(alias, original string)` - 添加参数别名

### 全局实例

```go
// 使用全局 CommandLine 实例
verbose := cli.Bool("v", false, "详细输出")
hello := cli.NewCLI("hello", "问候")
cli.AddCommand(hello)
cli.Run()
```

### 终端补全功能

支持命令、参数和文件路径的自动补全，Zsh/Fish 支持显示说明信息。

#### 使用示例

```go
root := cli.NewCLI("myapp", "我的应用")
var config, env string
root.StringVarShortLong(&config, "c", "config", "", "配置文件")
root.StringVarShortLong(&env, "e", "env", "dev", "环境")

// 注册补全（使用长参数名）
root.RegisterFileCompletion("config", ".json", ".yaml")
root.RegisterDirectoryCompletion("output")

// 自定义补全（推荐使用带描述版本）
root.RegisterCustomCompletion("env", func(toComplete string) []cli.CompletionItem {
    return []cli.CompletionItem{
        {Value: "dev", Description: "开发环境"},
        {Value: "prod", Description: "生产环境"},
    }
})
completion := cli.NewCLI("completion", "生成补全脚本")
completion.Usage = "[bash|zsh|fish]"
completion.CommandFunc = func(args []string) error {
	if len(args) == 0 {
		fmt.Println("用法: myapp completion [bash|zsh|fish]")
		return nil
	}
	switch args[0] {
	case "bash", "zsh", "fish":
		return root.GenerateCompletion(args[0])
	default:
		return fmt.Errorf("不支持的 shell: %s", args[0])
	}
}

```

#### 安装与使用

```bash
# 生成并安装补全脚本
myapp completion bash | sudo tee /etc/bash_completion.d/myapp
myapp completion zsh > "${fpath[1]}/_myapp"
myapp completion fish > ~/.config/fish/completions/myapp.fish

# 使用
myapp <TAB>        # 子命令补全
myapp -c <TAB>     # 文件补全
myapp --<TAB>      # 参数补全
```

#### API

- `RegisterFileCompletion(flag, exts...)` - 文件补全
- `RegisterDirectoryCompletion(flag)` - 目录补全  
- `RegisterCustomCompletion(flag, func)` - 自定义补全
  - `func(string) []string` - 简单补全
  - `func(string) []CompletionItem` - 带描述补全
- `RegisterCustomCompletionPrefixMatches(flag, valus)` - 自定义补全，简单前缀匹配
  - `[]string` - 简单补全
  - `[]CompletionItem` - 带描述补全
- `GenerateCompletion(shell, cmd)` - 生成脚本


