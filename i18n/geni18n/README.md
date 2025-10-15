# geni18n

自动从 `.go` 文件中提取 `i18n.GetText("")` 调用的字符串并生成 JSON 文件。

## 功能

- 扫描指定目录下的所有 Go 源文件
- 提取 `i18n.GetText()` 函数调用中的字符串字面量
- 生成格式化的 JSON 文件用于国际化翻译

## 安装

```bash
go install github.com/Rehtt/Kit/i18n/geni18n@latest
```

## 使用方法

### 基本用法

```bash
# 扫描当前目录
geni18n

# 扫描指定目录（支持短参数）
geni18n -p ./src

# 扫描指定目录（支持长参数）
geni18n --path ./src

# 指定输出目录和文件名
geni18n -p ./src -d ./locales -f zh-CN.json

# 详细输出模式
geni18n -v --path ./src

# 查看帮助
geni18n -h
geni18n --help
```

### 命令行参数

| 短参数 | 长参数 | 默认值 | 说明 |
|--------|--------|--------|------|
| `-p` | `--path` | `.` | 源代码路径 |
| `-d` | `--output-dir` | `i18n` | 输出目录 |
| `-f` | `--output-file` | `default.json` | 输出文件名 |
| `-i` | `--indent` | `true` | 是否格式化输出 JSON |
| `-v` | `--verbose` | `false` | 详细输出模式 |

## 示例

假设你的代码中有：

```go
package main

import "github.com/Rehtt/Kit/i18n"

func main() {
    msg := i18n.GetText("Hello, World!")
    greeting := i18n.GetText("Welcome to our application")
}
```

运行 `geni18n` 后，会在 `i18n/default.json` 生成：

```json
{
  "Hello, World!": "Hello, World!",
  "Welcome to our application": "Welcome to our application"
}
```

你可以复制这个文件并翻译成其他语言。
