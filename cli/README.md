### `cli` —— 轻量级命令行工具集合（基于标准库 flag）

`cli` 包是对 Go 标准库 `flag` 的薄封装，提供更友好的子命令组织、嵌套解析与帮助信息输出。适合快速构建多层级命令行工具，同时保持零第三方依赖、可读性强的代码风格。

---

### 安装

```bash
go get github.com/Rehtt/Kit/cli
```

在代码中：

```go
import "github.com/Rehtt/Kit/cli"
```

---

### 快速上手

```go
package main

import (
    "flag"
    "fmt"
    "os"

    kitcli "github.com/Rehtt/Kit/cli"
)

func main() {
    // 根命令
    root := kitcli.NewCLI("app", "示例应用", flag.ContinueOnError)
    root.Usage = " [command] [flags]"  // 设置用法说明
    var verbose bool
    root.BoolVar(&verbose, "v", false, "开启详细输出")

    // 子命令：hello
    hello := kitcli.NewCLI("hello", "打印问候语", flag.ContinueOnError)
    hello.Usage = " [flags]"  // 子命令的用法说明
    name := hello.String("name", "world", "名字")
    hello.CommandFunc = func(args []string) error {
        if verbose {
            fmt.Println("verbose on")
        }
        fmt.Printf("Hello, %s\n", *name)
        return nil
    }

    // 二级子命令：user add
    user := kitcli.NewCLI("user", "用户操作", flag.ContinueOnError)
    user.Usage = " <subcommand> [flags]"  // 显示需要子命令
    add := kitcli.NewCLI("add", "添加用户", flag.ContinueOnError)
    add.Usage = " [flags]"
    uname := add.String("name", "", "用户名")
    add.CommandFunc = func(args []string) error {
        fmt.Println("add", *uname)
        return nil
    }
    _ = user.AddCommand(add)

    // 绑定子命令到根命令
    _ = root.AddCommand(hello, user)

    // 解析并执行命令
    if err := root.Run(os.Args[1:]); err != nil {
        // 根据 errorHandling 决定行为；这里选择自行处理
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(2)
    }
}
```

运行体验（示例）：

```text
$ app -h
示例应用

Usage: app [command] [flags]
  -v    开启详细输出

Subcommands:
  hello  打印问候语
  user   用户操作

$ app hello -h
打印问候语

Usage: hello [flags]
  -name string
        名字 (default "world")

$ app hello -v -name Alice
verbose on
Hello, Alice

$ app user add -name Bob
add Bob
```

---

### 核心概念与 API

- **CLI 结构体**：表示一个命令。字段含义：
  - `Use`: 命令名（如 `hello`、`user`）
  - `Instruction`: 简要说明（用于帮助信息的开头，在 Usage 行之前显示）
  - `Usage`: 用法字符串，附加在 "Usage: 命令名" 之后显示。例如设置为 `" [flags]"` 将显示 `Usage: hello [flags]`，设置为 `" <subcommand> [flags]"` 将显示 `Usage: user <subcommand> [flags]`
  - `CommandFunc`: 命令执行函数，签名为 `func(args []string) error`，返回错误时会传递给调用者
  - `FlagSet`: 该命令的 `*flag.FlagSet`，用于定义与解析参数
  - `SubCommands`: 子命令映射（`map[string]*CLI`）

- **构造函数**
  - `NewCLI(use, instruction string, errorHandling flag.ErrorHandling) *CLI`
    - `errorHandling` 透传给 `flag.NewFlagSet`，常用值：
      - `flag.ContinueOnError`: 出错返回错误，便于外层处理
      - `flag.ExitOnError`: 出错直接退出程序

- **方法**
  - `AddCommand(cli ...*CLI) error`: 挂载一个或多个子命令，命名重复会报错
  - `Help()`: 打印当前命令的帮助与子命令列表
  - `Parse(arguments []string) error`: 解析参数；若存在子命令并检测到首个非 flag 参数为子命令名，则递归解析对应子命令
  - `Run(arguments []string) error`: `Parse` 的别名，语义上更清晰地表达"执行命令"

行为细节：

- `Parse` 内部会将 `c.FlagSet.Usage` 指向 `c.Help`，因此 `-h/--help` 会打印包含子命令的帮助信息。
- 当存在子命令且第一个位置参数匹配子命令时，会递归进入子命令解析。
- 未匹配到子命令时会打印帮助并返回错误（包装了 `flag.ErrHelp`），调用方可据此设置退出码（例如 2）。
- `-h/--help` 或 `CommandFunc` 返回 `flag.ErrHelp` 时，返回 `nil`（仅展示帮助，不视为错误）。
- `Help` 使用 `FlagSet.Output()` 的 `io.Writer` 输出；可通过 `FlagSet.SetOutput(w)` 重定向到自定义 writer。

---

### Usage 字段的使用方法

`Usage` 字段用于定制帮助信息中的用法提示行，格式为：`Usage: <命令名><Usage字段内容>`

**常见用法模式：**

```go
// 1. 仅接受 flags 的命令
cmd := cli.NewCLI("serve", "启动服务", flag.ContinueOnError)
cmd.Usage = " [flags]"
// 帮助输出：Usage: serve [flags]

// 2. 需要子命令的命令
cmd := cli.NewCLI("git", "版本控制工具", flag.ContinueOnError)
cmd.Usage = " <subcommand> [flags]"
// 帮助输出：Usage: git <subcommand> [flags]

// 3. 需要位置参数的命令
cmd := cli.NewCLI("copy", "复制文件", flag.ContinueOnError)
cmd.Usage = " <source> <dest> [flags]"
// 帮助输出：Usage: copy <source> <dest> [flags]

// 4. 可选位置参数
cmd := cli.NewCLI("build", "构建项目", flag.ContinueOnError)
cmd.Usage = " [target] [flags]"
// 帮助输出：Usage: build [target] [flags]

// 5. 多个参数的命令
cmd := cli.NewCLI("archive", "打包文件", flag.ContinueOnError)
cmd.Usage = " <output> <file1> [file2] ... [flags]"
// 帮助输出：Usage: archive <output> <file1> [file2] ... [flags]
```

**约定说明：**
- `<arg>` 表示必需参数
- `[arg]` 表示可选参数
- `...` 表示可重复的参数
- 建议 `[flags]` 放在最后

**完整示例：**

```go
package main

import (
    "flag"
    "fmt"
    "os"
    
    cli "github.com/Rehtt/Kit/cli"
)

func main() {
    root := cli.NewCLI("fileutil", "文件处理工具", flag.ContinueOnError)
    root.Usage = " <command> [flags]"
    
    // copy 命令：需要源和目标参数
    copy := cli.NewCLI("copy", "复制文件", flag.ContinueOnError)
    copy.Usage = " <source> <dest> [flags]"
    force := copy.Bool("f", false, "强制覆盖")
    copy.CommandFunc = func(args []string) error {
        if len(args) < 2 {
            fmt.Fprintln(os.Stderr, "错误: 需要指定源文件和目标文件")
            copy.Help()
            return fmt.Errorf("参数不足")
        }
        fmt.Printf("复制 %s -> %s (force=%v)\n", args[0], args[1], *force)
        return nil
    }
    
    // list 命令：可选目录参数
    list := cli.NewCLI("list", "列出文件", flag.ContinueOnError)
    list.Usage = " [directory] [flags]"
    all := list.Bool("a", false, "显示隐藏文件")
    list.CommandFunc = func(args []string) error {
        dir := "."
        if len(args) > 0 {
            dir = args[0]
        }
        fmt.Printf("列出 %s (all=%v)\n", dir, *all)
        return nil
    }
    
    _ = root.AddCommand(copy, list)
    
    if err := root.Run(os.Args[1:]); err != nil {
        os.Exit(1)
    }
}
```

运行效果：

```text
$ fileutil copy -h
复制文件

Usage: copy <source> <dest> [flags]
  -f    强制覆盖

$ fileutil list -h
列出文件

Usage: list [directory] [flags]
  -a    显示隐藏文件
```

---

### 默认实例与包装 API

本包提供一个默认的全局实例 `cli.CommandLine` 及其包装函数，便于快速开发简单工具：

```go
package main

import (
    "flag"
    "fmt"
    "os"

    cli "github.com/Rehtt/Kit/cli"
)

func main() {
    // 根级 flag（对所有命令生效）
    verbose := cli.Bool("v", false, "开启详细输出")

    // 定义一个子命令并绑定其专属 flag
    hello := cli.NewCLI("hello", "打印问候语", flag.ContinueOnError)
    name := hello.String("name", "world", "名字")
    hello.CommandFunc = func(args []string) error {
        if *verbose { fmt.Println("verbose on") }
        fmt.Printf("Hello, %s\n", *name)
        return nil
    }
    _ = cli.AddCommand(hello)

    if err := cli.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(2)
    }
}
```

包装函数速览（均作用于 `cli.CommandLine`）：

- 命令执行：`Parse() error`（解析并执行）、`Run() error`（Parse 的别名）
- 状态查询：`Parsed() bool`、`Args() []string`、`NArg() int`
- 命令管理：`AddCommand(cli ...*CLI) error`（添加子命令到默认实例）
- 标准类型：`BoolVar/Bool`、`StringVar/String`、`IntVar/Int`、`Int64Var/Int64`、`UintVar/Uint`、`Uint64Var/Uint64`、`Float64Var/Float64`、`DurationVar/Duration`
- 自定义类型：`TextVar(p encoding.TextUnmarshaler, name string, value encoding.TextMarshaler, usage string)`、`Var(p flag.Value, name, usage string)`
- 自定义解析回调：`Func(name, usage string, fn func(string) error)`、`BoolFunc(name, usage string, fn func(string) error)`
- 访问与遍历：`VisitAll(fn)`、`Visit(fn)`、`Lookup(name)`、`Set(name, value)`

注意：根级定义的 flag（通过包装函数定义）适用于所有命令；子命令私有的 flag 应在对应的 `*CLI` 上定义。

---

### 使用建议（Best Practices）

- **选择合适的 error handling**：
  - 推荐根命令使用 `flag.ContinueOnError`，让错误以返回值形式暴露，便于统一处理与测试。
- **按层定义各自的 Flag**：
  - 每个命令有独立的 `FlagSet`，避免不同层级的 flag 混淆。
- **合理设置 Usage 字段**：
  - 始终设置 `Usage` 字段，让用户清楚知道命令的使用方式
  - 对于有子命令的命令，使用 `" <subcommand> [flags]"` 格式
  - 对于需要位置参数的命令，明确标注必需参数 `<arg>` 和可选参数 `[arg]`
  - Usage 开头记得加空格，如 `" [flags]"` 而非 `"[flags]"`
- **参数验证**：
  - 在 `CommandFunc` 中验证位置参数数量，不足时调用 `Help()` 并返回错误
- **稳定输出**：
  - 如需稳定的子命令显示顺序，可在调用 `AddCommand` 前对切片排序，或在 `Help` 中对 `SubCommands` 的键排序后输出。

---

### 常见问题（FAQ）

- **Usage 字段不设置会怎样？**
  - 如果不设置 `Usage` 字段（默认为空字符串），帮助信息会显示为 `Usage: 命令名`，不会有后续的参数说明。建议总是设置该字段以提供清晰的使用指引。
- **如何实现命令别名？**
  - 目前 `SubCommands` 使用 map，键即命令名。可手动插入多个键映射到同一 `*CLI` 以实现别名。例如：
  ```go
  cmd := cli.NewCLI("list", "列出项目", flag.ContinueOnError)
  root.SubCommands["list"] = cmd
  root.SubCommands["ls"] = cmd  // 别名
  ```
- **如何让父命令的 flag 影响子命令？**
  - 在父命令的变量中保存状态（如 `verbose`），子命令执行时按需读取（示例见上）。
- **未知子命令是否会返回错误？**
  - 会。未知子命令会打印帮助并返回一个包装了 `flag.ErrHelp` 的错误（形如 `unknown subcommand "xxx": flag: help requested`）。
    调用方可据此区分“展示帮助”与“输入错误”，并设置合适的退出码（例如 2）。
- **如何处理位置参数？**
  - 在 `CommandFunc` 中，`args` 参数包含所有解析后的非 flag 参数。通过检查 `len(args)` 和访问 `args[0]`, `args[1]` 等来获取位置参数。示例见上方 Usage 章节的 `copy` 命令。

---

### 特性说明

- **错误处理**：`CommandFunc` 返回 `error` 类型，支持完整的错误传递链，便于外层统一处理错误情况。
- **灵活的 API**：既提供 `NewCLI` 构造独立实例的方式，也提供基于全局 `CommandLine` 的包装函数，适应不同复杂度的项目需求。
- **Parse vs Run**：两者功能相同，`Run` 是 `Parse` 的别名，语义上更清晰地表达"执行命令"的意图。

### 已知限制与改进空间

- **帮助输出顺序**：`SubCommands` 为 map，Help 中遍历无序。如需稳定顺序，可在 `Help()` 方法外自行对子命令排序后展示。
- **未知子命令的返回值**：当前返回 `nil` 并显示帮助信息，调用方难以区分"正常展示帮助"与"输入错误"。如需严格错误处理，可在外层检查参数或扩展 `Parse` 行为。

---

### 许可协议

本包遵循仓库根目录的许可协议，详见 `LICENSE`。


