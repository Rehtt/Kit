# I18n 国际化模块

轻量级的Go国际化(i18n)库，支持多语言文本管理和动态语言切换。

## 功能特性

- 🌍 支持多语言文本管理
- 🔄 动态语言切换
- 📁 基于JSON文件的语言包
- 🚀 高性能文本查找
- 💾 自动缓存机制
- 🎯 简单易用的API

## 安装使用

```go
import "github.com/rehtt/Kit/i18n"
```

## 快速开始

### 1. 准备语言文件

在项目根目录创建 `i18n` 文件夹，并添加语言文件：

**i18n/default.json** (默认语言)
```json
{
  "hello": "Hello",
  "welcome": "Welcome to our application",
  "goodbye": "Goodbye",
  "user_not_found": "User not found"
}
```

**i18n/zh-CN.json** (中文)
```json
{
  "hello": "你好",
  "welcome": "欢迎使用我们的应用",
  "goodbye": "再见",
  "user_not_found": "用户未找到"
}
```

**i18n/ja.json** (日语)
```json
{
  "hello": "こんにちは",
  "welcome": "私たちのアプリケーションへようこそ",
  "goodbye": "さようなら",
  "user_not_found": "ユーザーが見つかりません"
}
```

### 2. 基本使用

```go
package main

import (
    "fmt"
    "golang.org/x/text/language"
    "github.com/rehtt/Kit/i18n"
)

func main() {
    // 获取文本（使用默认语言）
    fmt.Println(i18n.GetText("hello"))    // 输出: Hello
    fmt.Println(i18n.GetText("welcome"))  // 输出: Welcome to our application
    
    // 获取指定语言的文本
    zh := language.Chinese
    fmt.Println(i18n.GetText("hello", zh))    // 输出: 你好
    fmt.Println(i18n.GetText("welcome", zh))  // 输出: 欢迎使用我们的应用
    
    ja := language.Japanese
    fmt.Println(i18n.GetText("hello", ja))    // 输出: こんにちは
}
```

## API 文档

### 核心函数

#### SetLang
```go
func SetLang(l *language.Tag) error
```
设置默认语言。

**参数：**
- `l`: 语言标签，传入 `nil` 使用默认语言

**返回值：**
- `error`: 错误信息，如果语言文件不存在或格式错误

**示例：**
```go
// 设置默认语言为中文
zh := language.Chinese
err := i18n.SetLang(&zh)
if err != nil {
    log.Fatal(err)
}

// 重置为默认语言
err = i18n.SetLang(nil)
```

#### GetText
```go
func GetText(str string, lang ...language.Tag) string
```
获取指定键的本地化文本。

**参数：**
- `str`: 文本键
- `lang`: 可选的语言标签，如果不指定则使用默认语言

**返回值：**
- `string`: 本地化文本，如果找不到对应翻译则返回原始键值

**示例：**
```go
// 使用默认语言
text := i18n.GetText("hello")

// 使用指定语言
zh := language.Chinese
text := i18n.GetText("hello", zh)

// 如果键不存在，返回键本身
text := i18n.GetText("non_existent_key") // 返回: "non_existent_key"
```

#### SetPath
```go
func SetPath(path string)
```
设置语言文件目录路径。

**参数：**
- `path`: 语言文件目录路径

**示例：**
```go
// 设置自定义语言文件路径
i18n.SetPath("./locales")
```

## 使用示例

### 基本多语言支持

```go
package main

import (
    "fmt"
    "golang.org/x/text/language"
    "github.com/rehtt/Kit/i18n"
)

func main() {
    // 支持的语言
    languages := []language.Tag{
        language.English,
        language.Chinese,
        language.Japanese,
    }
    
    key := "welcome"
    
    for _, lang := range languages {
        text := i18n.GetText(key, lang)
        fmt.Printf("%s: %s\n", lang.String(), text)
    }
    
    // 输出:
    // en: Welcome to our application
    // zh: 欢迎使用我们的应用  
    // ja: 私たちのアプリケーションへようこそ
}
```

### Web应用中的使用

```go
package main

import (
    "net/http"
    "golang.org/x/text/language"
    "github.com/rehtt/Kit/i18n"
)

func handler(w http.ResponseWriter, r *http.Request) {
    // 从请求头获取用户首选语言
    acceptLang := r.Header.Get("Accept-Language")
    tags, _, _ := language.ParseAcceptLanguage(acceptLang)
    
    var userLang language.Tag
    if len(tags) > 0 {
        userLang = tags[0]
    }
    
    // 获取本地化消息
    message := i18n.GetText("welcome", userLang)
    
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.Write([]byte(message))
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
```

### 错误消息国际化

```go
package main

import (
    "errors"
    "fmt"
    "golang.org/x/text/language"
    "github.com/rehtt/Kit/i18n"
)

type LocalizedError struct {
    Key  string
    Lang language.Tag
}

func (e LocalizedError) Error() string {
    return i18n.GetText(e.Key, e.Lang)
}

func findUser(id int, lang language.Tag) error {
    // 模拟用户查找失败
    if id <= 0 {
        return LocalizedError{
            Key:  "user_not_found",
            Lang: lang,
        }
    }
    return nil
}

func main() {
    zh := language.Chinese
    en := language.English
    
    err1 := findUser(-1, zh)
    fmt.Println(err1) // 输出: 用户未找到
    
    err2 := findUser(-1, en)
    fmt.Println(err2) // 输出: User not found
}
```

### 配置管理

```go
package main

import (
    "fmt"
    "golang.org/x/text/language"
    "github.com/rehtt/Kit/i18n"
)

type Config struct {
    DefaultLang language.Tag
    LangPath    string
}

func initI18n(config Config) error {
    // 设置语言文件路径
    if config.LangPath != "" {
        i18n.SetPath(config.LangPath)
    }
    
    // 设置默认语言
    return i18n.SetLang(&config.DefaultLang)
}

func main() {
    config := Config{
        DefaultLang: language.Chinese,
        LangPath:    "./locales",
    }
    
    err := initI18n(config)
    if err != nil {
        fmt.Printf("初始化i18n失败: %v\n", err)
        return
    }
    
    // 现在默认使用中文
    fmt.Println(i18n.GetText("hello")) // 输出: 你好
}
```

## 语言文件格式

### JSON格式要求

语言文件必须是有效的JSON格式，键值对都是字符串：

```json
{
  "key1": "value1",
  "key2": "value2",
  "nested.key": "nested value"
}
```

### 文件命名规范

- `default.json`: 默认语言文件
- `{language-tag}.json`: 特定语言文件

支持的语言标签格式：
- `en`: 英语
- `zh`: 中文
- `zh-CN`: 简体中文
- `zh-TW`: 繁体中文
- `ja`: 日语
- `ko`: 韩语
- `fr`: 法语
- `de`: 德语
- `es`: 西班牙语

### 优化技巧

1. **避免重复翻译**: 如果某个键的值与键名相同，系统会自动忽略该条目
2. **使用嵌套键名**: 可以使用点号分隔的键名来组织翻译，如 `"user.profile.name"`
3. **保持键名一致**: 确保所有语言文件中的键名保持一致

## 性能特点

- **缓存机制**: 已加载的语言文件会被缓存，避免重复读取
- **延迟加载**: 只有在首次使用时才加载语言文件
- **内存优化**: 自动清理相同键值对，减少内存占用
- **快速查找**: 使用map结构实现O(1)时间复杂度的文本查找

## 注意事项

1. **文件路径**: 默认在 `i18n` 目录下查找语言文件
2. **错误处理**: 如果语言文件不存在或格式错误，相关函数会返回错误
3. **回退机制**: 如果指定语言的翻译不存在，会返回原始键名
4. **线程安全**: 模块是线程安全的，可以在并发环境中使用

## 测试

创建测试文件和语言文件后运行：

```bash
go test ./i18n
```
