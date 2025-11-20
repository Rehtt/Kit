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
    // 参数示例
    env := hello.String("env", "dev", "环境", cli.NewFlagItemSelectString("dev", "prod"))
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

支持使用混合短名：`app -vh 127.0.0.1 --port 9000`

注意使用混合短名时前面短名必须为bool类型，最后一个可以是value类型

例如: -abc 展开为 -a -b -c

例如: -abf value 展开为 -a -b -f value


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

提供两种补全方案，支持 Bash、Zsh、Fish。

#### 方案对比

| 特性 | 动态补全 (completion) | 静态补全 (completion_gen) |
|------|---------------------|--------------------------|
| 实现方式 | 程序运行时生成 | 静态脚本 |
| 自定义补全 | ✅ 支持复杂逻辑 | ❌ 仅支持基础类型 |
| 运行时开销 | 每次补全调用程序 | 纯脚本，无开销 |
| 动态数据 | ✅ 可查询实时数据 | ❌ 固定选项 |
| 维护成本 | 代码变更需重新安装 | 代码变更需重新生成 |
| 适用场景 | 需要动态数据/复杂逻辑 | 简单命令行工具 |

---

#### 方案一：动态补全 (completion)

**适合**：需要自定义补全逻辑、查询动态数据（如 Git 分支、数据库列表等）

```go
import "github.com/Rehtt/Kit/cli/completion"

root := cli.NewCLI("myapp", "我的应用")

// FlagItem 自动生成补全
root.StringVarShortLong(&config, "c", "config", "", "配置文件",
    cli.NewFlagItemFile())
root.StringVarShortLong(&env, "e", "env", "dev", "环境",
    cli.NewFlagItemSelectString("dev", "prod"))

// 创建补全管理器
cm := completion.New(root)

// 可选：自定义补全逻辑（支持动态数据）
cm.RegisterCustomCompletion(root, "branch", func(toComplete string) []string {
    return getGitBranches() // 实时查询 Git 分支
})

// 添加 completion 子命令
comp := cli.NewCLI("completion", "生成补全脚本")
comp.CommandFunc = func(args []string) error {
    if len(args) == 0 {
        return fmt.Errorf("用法: myapp completion [bash|zsh|fish]")
    }
    return cm.GenerateCompletion(args[0])
}
root.AddCommand(comp)
```

**安装**：

```bash
# 生成并安装（脚本会回调程序获取补全）
myapp completion bash | sudo tee /etc/bash_completion.d/myapp
myapp completion zsh > "${fpath[1]}/_myapp"
myapp completion fish > ~/.config/fish/completions/myapp.fish
```

**详见**：[completion/README.md](completion/README.md)

---

#### 方案二：静态补全 (completion_gen)

**适合**：简单工具、追求性能、无需动态数据

使用 `completion_gen` 工具从源码生成纯脚本补全文件：

```bash
# 安装工具
go install github.com/Rehtt/Kit/cli/completion_gen@latest

# 生成补全脚本（解析源码，一次生成）
./completion_gen -s bash -n myapp /path/to/project > myapp.bash
./completion_gen -s zsh -n myapp /path/to/project > _myapp
./completion_gen -s fish -n myapp /path/to/project > myapp.fish

# 安装
sudo cp myapp.bash /etc/bash_completion.d/
cp _myapp "${fpath[1]}/"
cp myapp.fish ~/.config/fish/completions/
```

**优点**：
- ✅ 纯脚本，补全速度极快
- ✅ 自动识别 `FlagItem` 类型（文件/目录/选项）
- ✅ 无需在程序中添加 completion 命令

**限制**：
- ❌ 不支持运行时动态补全
- ❌ 代码变更后需重新生成脚本

**详见**：[completion_gen/README.md](completion_gen/README.md)


