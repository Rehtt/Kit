# completion_gen

CLI 命令行静态补全功能脚本生成，支持 Bash、Zsh、Fish 等 Shell 的自动补全。

> 用于从使用 `github.com/Rehtt/Kit/cli` 库的 Go 项目中自动生成 shell 补全脚本。

## 安装

```bash
go install github.com/Rehtt/Kit/cli/completion_gen@latest
```

## 使用方法

### 基本用法

```bash
# 为当前目录的项目生成 bash 补全脚本
./completion_gen .

# 指定 shell 类型
./completion_gen -s zsh /path/to/project

# 输出到文件
./completion_gen -s bash -o mycmd.bash /path/to/project

# 指定命令名称
./completion_gen -n mycommand -s fish /path/to/project
```

### 命令选项

```
-s, --shell string      生成的 shell 类型 (bash/zsh/fish，默认: bash)
-n, --name string       命令名称（默认从代码自动推断）
-o, --output string     输出文件路径（默认输出到 stdout）
-r, --recursive bool    是否递归扫描子目录（默认: true）
-v, --verbose bool      显示详细的解析过程（默认: false）
```

### 安装补全脚本

生成补全脚本后，需要将其安装到对应 shell 的补全目录：

#### Bash

```bash
# 生成补全脚本
./completion_gen -s bash -n mycmd /path/to/project > mycmd.bash

# 安装到系统
sudo cp mycmd.bash /etc/bash_completion.d/

# 或安装到用户目录
mkdir -p ~/.local/share/bash-completion/completions
cp mycmd.bash ~/.local/share/bash-completion/completions/mycmd

# 重新加载
source ~/.bashrc
```

#### Zsh

```bash
# 生成补全脚本
./completion_gen -s zsh -n mycmd /path/to/project > _mycmd

# 安装到 zsh 补全目录
sudo cp _mycmd /usr/local/share/zsh/site-functions/

# 或安装到用户目录
mkdir -p ~/.zsh/completions
cp _mycmd ~/.zsh/completions/

# 在 ~/.zshrc 中添加
fpath=(~/.zsh/completions $fpath)

# 重新加载
exec zsh
```

#### Fish

```bash
# 生成补全脚本
./completion_gen -s fish -n mycmd /path/to/project > mycmd.fish

# 安装到 fish 补全目录
cp mycmd.fish ~/.config/fish/completions/

# 重新加载
fish_update_completions
```

## 工作原理

`completion_gen` 通过 AST（抽象语法树）解析 Go 源代码，识别以下内容：

1. **命令定义**：`cli.NewCLI()` 调用
2. **Flag 注册**：`StringVar`、`IntVar`、`BoolVar` 等方法调用
3. **子命令关系**：`AddCommand()` 调用
4. **补全类型**：`NewFlagItemFile()`、`NewFlagItemDir()`、`NewFlagItemSelect()` 等

然后根据解析结果，使用内置模板生成对应 shell 的补全脚本。

## 示例

假设你有一个使用 `cli` 库的项目：

```go
package main

import "github.com/Rehtt/Kit/cli"

func main() {
    app := cli.NewCLI("mytool", "一个示例工具")
    
    app.StringVarShortLong(&config, "c", "config", "config.yaml", "配置文件路径",
        cli.NewFlagItemFile())
    app.BoolVarShortLong(&verbose, "v", "verbose", false, "详细输出")
    
    subCmd := cli.NewCLI("list", "列出所有项")
    app.AddCommand(subCmd)
    
    app.Parse(os.Args[1:])
}
```

生成补全脚本：

```bash
./completion_gen -s bash -n mytool . > mytool.bash
```

安装后，你就可以使用 Tab 键补全：

```bash
mytool <TAB>          # 显示: list
mytool -<TAB>         # 显示: -c --config -v --verbose
mytool --config <TAB> # 补全文件路径
```

## 注意事项

- 仅支持使用 `github.com/Rehtt/Kit/cli` 库的项目
- 需要能够访问项目的 Go 源代码
- 代码必须能够被正确解析（语法正确）
- 动态注册的命令可能无法识别

## 相关文档

- [cli 库文档](../README.md)
- [completion 包文档](../completion/README.md)

## License

MIT License - Copyright (c) 2025 Rehtt
