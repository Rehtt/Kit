## requester

轻量级 HTTP 请求构造器与执行器，提供链式 API 来简化 GET/POST/PUT/DELETE 请求、JSON 编解码、Header 设置与调试输出。内部使用 `sync.Pool` 复用实例，降低临时分配与 GC 压力。

### 安装

```bash
go get github.com/Rehtt/Kit@latest
```

导入：

```go
import "github.com/Rehtt/Kit/requester"
```

### 快速开始

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/Rehtt/Kit/requester"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 简单 GET，并以字符串返回
    r := requester.NewRequester()
    defer r.Close() // 使用完归还池子（会安全关闭 response）

    body := r.Get("https://httpbin.org/get").AsString(ctx)
    fmt.Println(body)
}
```

#### 发送 JSON 并解析 JSON 响应

```go
type Resp struct {
    JSON map[string]any `json:"json"`
}

ctx := context.Background()
r := requester.NewRequester()
defer r.Close()

var out Resp
err := r.PostJSON("https://httpbin.org/post", map[string]any{"hello":"world"}).AsJSON(ctx, &out)
if err != nil {
    panic(err)
}
fmt.Println(out.JSON["hello"]) // world
```

#### 自定义 Header 与调试输出

```go
r := requester.NewRequester().
    Debug(true).                               // 打印请求与响应
    SetHead("Authorization", "Bearer xxx").    // 覆盖式设置
    AddHead("X-Trace-Id", "abc-123")          // 追加式设置

bytes := r.Post("https://httpbin.org/anything", nil).AsBytes(context.Background())
fmt.Println(string(bytes))
r.Close()
```

### API 说明

- `NewRequester() *Requester`
  - 从池中获取一个新的 `Requester` 并清空状态。

- `(*Requester) Get(u string)` / `Post(u string, body io.Reader)` / `Put(u string, body io.Reader)` / `Delete(u string, body io.Reader)`
  - 配置请求方法、URL 与请求体（可为 `nil`）。

- `(*Requester) RequestJSON(method, url string, obj any)`
  - 设置 `content-type: application/json`，当 `obj` 为 `string`/`[]byte` 时直接作为原始 body；否则会进行 JSON 编码。

- `(*Requester) PostJSON(url string, obj any)` / `PutJSON(url string, obj any)` / `DeleteJSON(url string, obj any)`
  - `RequestJSON` 的便捷方法。

- `(*Requester) AddHead(key, value string)` / `SetHead(key, value string)`
  - 追加/覆盖请求头。

- `(*Requester) Response(ctx context.Context) (*http.Response, error)`
  - 发送请求并返回底层 `*http.Response`；调用者需负责 `resp.Body.Close()` 或最终调用 `Requester.Close()`。

- `(*Requester) AsBytes(ctx context.Context) []byte`
  - 发送请求并返回响应体字节切片；内部会在读取后关闭 `resp.Body`。

- `(*Requester) AsString(ctx context.Context) string`
  - 等同 `string(AsBytes(ctx))`。

- `(*Requester) AsJSON(ctx context.Context, obj any) error`
  - 发送请求并将 JSON 响应解码到 `obj`；内部会在解码后关闭 `resp.Body`。

- `(*Requester) Debug(debug bool) *Requester`
  - 开启后会将请求与响应（含 Header 与 Body）以模板格式输出到标准输出，便于排错。

- `(*Requester) Clear() *Requester`
  - 重置内部状态（URL、方法、Body、Header、错误、响应指针等）。

- `(*Requester) Clone() *Requester`
  - 复制一个新的 `Requester`，会克隆 Header 并复制 URL、方法与 `body` 引用（注意：`body` 为同一引用）。

- `(*Requester) Close()`
  - 安全关闭已有响应体并将实例放回池中。建议在使用结束后调用；或复用请调用 `Clear()`。

- `(*Requester) GetErr() error`
  - 返回构建/执行过程中的最后一个错误（若有）。

### 重要说明与最佳实践

- **资源管理**：
  - 使用 `AsBytes` / `AsJSON` 会自动关闭响应体；
  - 使用 `Response` 获取底层响应时，请手动 `defer resp.Body.Close()`，或在最后 `Requester.Close()`。

- **实例复用与并发**：
  - `Requester` 通过 `sync.Pool` 复用，减少分配成本；
  - 单个实例不是并发安全的，请勿在多个 goroutine 间并发复用同一实例；
  - 推荐按请求生命周期获取、使用、`Close()` 归还。

- **Debug 输出**：
  - 开启 `Debug(true)` 后，会读取并回填请求/响应的 Body，保证后续流程仍可读取；
  - 输出到 `os.Stdout`，若不需要请关闭或自行重定向标准输出。

- **默认客户端**：
  - 当前使用 `http.DefaultClient` 发送请求；若需自定义超时、代理、重试等策略，可在外层使用 `context` 控制，或扩展本包以支持自定义 `http.Client`（欢迎 PR）。

### 错误处理示例

```go
r := requester.NewRequester()
defer r.Close()

resp, err := r.Get("https://httpbin.org/status/418").Response(context.Background())
if err != nil {
    // 构建/发送阶段错误
    panic(err)
}
defer resp.Body.Close()

if resp.StatusCode >= 400 {
    // 业务侧自定义错误处理
}
```

### 许可

遵循仓库根目录的 `LICENSE`。


