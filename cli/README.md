### `cli` —— 轻量级命令行工具集合

基于 Go 标准库 `flag` 的薄封装，提供子命令组织、长短参数支持和友好的帮助信息。

### 安装

```bash
go get github.com/Rehtt/Kit/cli
```

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


