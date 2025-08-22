### Kit/queue

轻量内存队列，支持阻塞/非阻塞获取、消息确认（ack）与超时回退（requeue）。默认以“至少一次投递”语义工作。

- **并发安全**：`Put`/`Get` 可并发调用
- **确认机制**：设置 `deadline` 后需手动 `Done(id)` 确认
- **超时回退**：未在截止时间内确认的消息将按策略回退（默认重新入队）

### 安装

```shell
go get github.com/Rehtt/Kit/queue
```

### 快速开始

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/Rehtt/Kit/queue"
)

func main() {
    q := queue.NewQueue()
    ctx := context.Background()

    // 生产
    go q.Put("hello")

    // 消费（阻塞获取），不需要确认
    id, data, ok := q.Get(ctx, nil, true)
    fmt.Println("no-ack:", id != 0, data, ok)

    // 消费（阻塞获取），设置确认超时，需 Done
    go q.Put("need-ack")
    dl := time.Now().Add(30 * time.Second)
    id2, data2, ok2 := q.Get(ctx, &dl, true)
    if ok2 {
        fmt.Println("ack:", id2 != 0, data2)
        q.Done(id2) // 完成确认
    }
}
```

### API 说明

```go
type Queue struct {
    DeadlineFunc func(queue *Queue, id uint64, data any, deadline time.Time)
}

func NewQueue() *Queue
// 推入队列
func (q *Queue) Put(data any)
// 获取：
// - deadline 非 nil：需要调用 Done(id) 确认；否则将按策略回退
// - block 可选：true 表示阻塞等待；默认非阻塞尝试一次
func (q *Queue) Get(ctx context.Context, deadline *time.Time, block ...bool) (id uint64, data any, ok bool)
// 确认完成（仅当 Get 时传入了非 nil deadline 才需要）
func (q *Queue) Done(id uint64)
// 默认超时处理：将未确认的消息重新放回队列
func DefaultDeadlineFunc() func(queue *Queue, id uint64, data any, deadline time.Time)
```

### 确认与超时语义

- **非确认模式**：`Get(ctx, nil, ...)` 获取到即出队，不需要 `Done`。
- **确认模式**：`Get(ctx, &deadline, ...)` 获取到后，消息会被标记为“投递中”。
  - 在 `deadline` 之前调用 `Done(id)` 完成确认并移除标记。
  - 若超时未确认，内部扫描协程会触发 `DeadlineFunc`（默认将消息重新 `Put`）。
- **扫描周期**：默认每 `5m` 扫描一次超时消息（内部变量 `scanTime`）。
  - 当前版本未导出配置项，若需要可提 Issue/PR 以支持外部配置。
- **阻塞/非阻塞**：`block=true` 时阻塞等待直到获取数据或 `ctx.Done()`；否则非阻塞尝试一次。
- **投递语义**：默认策略实现“至少一次投递”（可能重复，消费者需具备幂等性）。

### 自定义超时处理

```go
q := queue.NewQueue()
q.DeadlineFunc = func(q *queue.Queue, id uint64, data any, deadline time.Time) {
    // 自定义：记录日志、打点、丢弃或改写策略
    // 例如：丢弃而不重入队
    // 不做任何操作即相当于丢弃
}
```

若希望维持默认行为（超时重入队），可使用：

```go
q.DeadlineFunc = queue.DefaultDeadlineFunc()
```

### 非阻塞与上下文取消示例

```go
// 非阻塞获取
id, data, ok := q.Get(context.Background(), nil)

// 阻塞但可取消
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()
id, data, ok = q.Get(ctx, nil, true)
```

### 基准测试

项目内提供了简单基准用例：

```shell
go test -run ^$ -bench . -benchmem github.com/Rehtt/Kit/queue
```

包含：
- `BenchmarkNewNode`
- `BenchmarkQueuePutGet`
- `BenchmarkParallelQueuePutGet`

### 设计与实现要点

- 使用 `crypto/rand` + FNV-1a 生成 `uint64` 作为消息 `Id`，避免顺序自增带来的热键与猜测问题。
- 消息出队后：
  - 非确认模式：不进入“投递中”集合。
  - 确认模式：存入内部 `sync.Map` 追踪，超时后按策略处理。
- 内部队列基于 `github.com/Rehtt/Kit/channel` 的 `Chan[*Node]` 实现。

### 注意事项

- “至少一次投递”可能产生重复消费，消费者应实现幂等。
- 扫描周期为离散时间片，超时重新入队可能有最多一个扫描周期的延迟。
- 当前不提供持久化，进程退出会丢失内存中的待确认消息。

### 许可证

本项目基于 MIT License（见仓库根目录 `LICENSE`）。


