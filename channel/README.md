### channel/unlimited：无界通道（Unbounded Channel）

简单、泛型、安全的“无界缓冲”通道实现。通过内部 goroutine + 链式队列将生产者与消费者解耦：写入侧永远向 `In` 写入，读取侧从 `Out` 读取。内部使用 `link.DLink[T]` 作为队列，保证 FIFO 顺序，支持在关闭后排空并优雅关闭 `Out`。

#### 特性
- **FIFO 顺序保证**：读出顺序与写入顺序一致。
- **无界缓冲**：不需要预估容量，突发流量也不会阻塞写入（但请注意内存占用风险）。
- **优雅关闭**：`Close()` 只会关闭 `In`，待内部队列数据全部发送完毕后自动关闭 `Out`；多次调用安全。
- **简洁 API**：`New/Len/Cap/Close` + `In/Out` 两个方向清晰易读。

#### 安装与导入
```go
import "github.com/Rehtt/Kit/channel"
```

#### 快速开始
```go
package main

import (
    "fmt"
    "github.com/Rehtt/Kit/channel"
)

func main() {
    c := channel.New[int]()
    defer c.Close()

    // 生产者
    go func() {
        for i := 0; i < 10; i++ {
            c.In <- i
        }
        // 关闭 In，触发排空后自动关闭 Out
        c.Close()
    }()

    // 消费者（单消费可严格保证顺序）
    for v := range c.Out {
        fmt.Println(v)
    }
}
```

#### API
```go
// New 创建无界通道。
func New[T any]() *Chan[T]

type Chan[T any] struct {
    In  chan<- T // 写入端（只写）
    Out <-chan T // 读取端（只读）
}

// 当前排队中的元素个数（可能滞后于最新写入片刻，取决于内部 goroutine 调度）。
func (c *Chan[T]) Len() int64

// 历史上曾达到的最大容量上界（由内部队列报告，便于观测峰值）。
func (c *Chan[T]) Cap() int64

// 关闭写入端 In，并在排空后自动关闭 Out；可安全重复调用。
func (c *Chan[T]) Close()
```

#### 行为与注意事项
- **多生产者**：可同时向同一个 `In` 并发写入，遵循 Go 原生 `chan` 的并发安全语义。
- **多消费者**：可以并发从 `Out` 读，但数据会在消费者之间分摊；若需要广播/一条消息被多个消费者同时处理，请使用上层的广播组件（例如本仓库中的 `util/chan_broadcaster`）。
- **内存占用**：该通道为“无界”，若消费者处理不及时，队列会增长并占用更多内存。建议：
  - 监控 `Len()`/`Cap()` 指标；
  - 对上游做背压/限速；
  - 在关闭时确保消费者尽快排空。
- **关闭语义**：`Close()` 后不再接受写入；内部会继续将队列中剩余数据发送到 `Out`，随后自动关闭 `Out`。这避免了消费者端的“读到一半通道即被关闭”的问题。

#### 与 `make(chan T, n)` 的对比
- **优点**：
  - 无需容量估算；
  - 突发写入不会因缓冲满而阻塞；
  - 关闭语义更友好（排空再关 `Out`）。
- **取舍**：
  - 无上限意味着需要自行约束上游速率或保障消费能力，以免内存增长。

#### 基准测试
项目内提供了两个基准：
- `BenchmarkSerial`：单写单读的吞吐性能；
- `BenchmarkParallel`：并行写入，观测内部队列与调度的争用情况。

运行：
```bash
go test -run ^$ -bench . -benchtime=2s ./channel
```

#### 适用场景
- **削峰填谷**：应对偶发的写入洪峰；
- **解耦生产/消费速率**：消费者较慢但不希望阻塞生产者；
- **简化容量管理**：无需纠结缓冲区大小选择。

#### 许可证
遵循仓库根目录的 `LICENSE`。


