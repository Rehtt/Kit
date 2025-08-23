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
    var verbose bool
    root.BoolVar(&verbose, "v", false, "开启详细输出")

    // 子命令：hello
    hello := kitcli.NewCLI("hello", "打印问候语", flag.ContinueOnError)
    name := hello.String("name", "world", "名字")
    hello.CommandFunc = func(args []string)error {
        if verbose {
            fmt.Println("verbose on")
        }
        fmt.Printf("Hello, %s\n", *name)
        return nil
    }

    // 二级子命令：user add
    user := kitcli.NewCLI("user", "用户操作", flag.ContinueOnError)
    add := kitcli.NewCLI("add", "添加用户", flag.ContinueOnError)
    uname := add.String("name", "", "用户名")
    add.CommandFunc = func(args []string)error {
        fmt.Println("add", *uname)
    }
    _ = user.AddCommand(add)

    // 绑定子命令到根命令
    _ = root.AddCommand(hello, user)

    // 解析命令
    if err := root.Parse(os.Args[1:]); err != nil {
        // 根据 errorHandling 决定行为；这里选择自行处理
        os.Exit(2)
    }
    return nil
}
```

运行体验（示例）：

```text
$ app -h
Usage of app:
  -v    开启详细输出

Subcommands:
  hello  打印问候语
  user   用户操作

$ app hello -h
Usage of hello:
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
  - `Instruction`: 简要说明（用于帮助信息）
  - `CommandFunc`: 命令执行函数，签名为 `func(args []string)`
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

行为细节：

- `Parse` 内部会将 `c.FlagSet.Usage` 指向 `c.Help`，因此 `-h/--help` 会打印包含子命令的帮助信息。
- 当存在子命令且第一个位置参数匹配子命令时，会递归进入子命令解析。
- 未匹配到子命令、或未设置 `Func` 时会打印帮助信息并返回 `nil`。
- `Help` 使用 `FlagSet.Output()` 的 `io.Writer` 输出；可通过 `FlagSet.SetOutput(w)` 重定向到自定义 writer。

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
    hello.CommandFunc = func(args []string)return {
        if *verbose { fmt.Println("verbose on") }
        fmt.Printf("Hello, %s\n", *name)
        return nil
    }
    _ = cli.AddCommand(hello)

    if err := cli.Parse(); err != nil { os.Exit(2) }
}
```

包装函数速览（均作用于 `cli.CommandLine`）：

- `Parse() error`、`Parsed() bool`、`Args() []string`、`NArg() int`
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
- **稳定输出**：
  - 如需稳定的子命令显示顺序，可在调用 `AddCommand` 前对切片排序，或在 `Help` 中对 `SubCommands` 的键排序后输出。

---

### 常见问题（FAQ）

- **如何实现命令别名？**
  - 目前 `SubCommands` 使用 map，键即命令名。可手动插入多个键映射到同一 `*CLI` 以实现别名。
- **如何让父命令的 flag 影响子命令？**
  - 在父命令的变量中保存状态（如 `verbose`），子命令执行时按需读取（示例见上）。
- **未知子命令是否会返回错误？**
  - 当前实现会打印帮助并返回 `nil`。如需报错，可在外层检查 `c.Arg(0)` 或自行扩展 `Parse` 行为。

---

### 代码改进建议（基于当前实现）

- **帮助输出顺序**：`SubCommands` 为 map，Help 中遍历无序。建议在输出前对键排序，或改为切片存储以保证稳定顺序。
- **错误传递能力**：`CommandFunc` 无返回值，无法向外传递错误。可新增 `FuncE func(args []string) error` 或在 `Parse` 中统一调用错误处理回调。
- **未知子命令的返回值**：当前返回 `nil`，调用方难以区分“正常展示帮助”与“输入错误”。建议在未知子命令时返回可识别的错误值。

---

### 许可协议

本包遵循仓库根目录的许可协议，详见 `LICENSE`。


