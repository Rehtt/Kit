# maps

`maps` 提供高性能、并发安全的泛型 Map 数据结构，包含：

- `RWMutexMap[T]`: 通过 `sync.RWMutex` 保护的单分片 Map，支持可选 TTL 与过期清理。
- `ConcurrentMap[T]`: 按 FNV32a 哈希分成 `SHARD_COUNT` 个分片（默认 32）的并发 Map，提升多核并发吞吐；支持全局或逐键 TTL，并自带定时清理协程。

## 特性

- 线程安全：读写分离锁/多分片削减锁竞争。
- TTL 过期：支持全局 TTL 与逐次 `Set` 指定 TTL；到期后 `Get` 不可见，`Clear` 会物理删除与复用对象。
- 零依赖：仅使用标准库；可与 `go test -race` 协同验证并发安全。

## 安装

```bash
go get github.com/Rehtt/Kit@latest
```

## 快速上手

### RWMutexMap

```go
package main

import (
    "fmt"
    "time"
    "github.com/Rehtt/Kit/maps"
)

func main() {
    // 无全局 TTL
    m := maps.NewRWMutexMap[int](0)

    m.Set("a", 10)
    if v, ok := m.Get("a"); ok {
        fmt.Println("a=", v)
    }

    // 针对该次写入指定 TTL
    m.Set("temp", 1, 1500*time.Millisecond)
    time.Sleep(1600 * time.Millisecond)
    m.Clear() // 清理到期键
    if _, ok := m.Get("temp"); !ok {
        fmt.Println("temp expired")
    }
}
```

### ConcurrentMap（分片并发）

```go
package main

import (
    "fmt"
    "time"
    "github.com/Rehtt/Kit/maps"
)

func main() {
    // 启用全局 TTL（示例 400ms）
    m := maps.NewConcurrentMap[int](maps.EnableExpired(400 * time.Millisecond))

    m.Set("foo", 42)
    if v, ok := m.Get("foo"); ok {
        fmt.Println("foo=", v)
    }

    time.Sleep(450 * time.Millisecond)
    if _, ok := m.Get("foo"); !ok {
        fmt.Println("foo expired (logical)")
    }

    // 分片内部可手动清理（通常不必，见“清理机制”）
    // for _, shard := range m.maps { shard.Clear() }
}
```

## API 概览

### RWMutexMap[T]

- `NewRWMutexMap[T](ttl time.Duration) *RWMutexMap[T]`
- `Set(key string, value T, ttl ...time.Duration)`
- `SetByFunc(key string, f func(old T) (new T), ttl ...time.Duration) (new T)`
- `Get(key string) (value T, ok bool)`
- `Delete(key string)`
- `Clear()`：清理内部最小堆里到期的键，并复用节点对象

说明：
- 若在构造时设置了全局 `ttl > 0`，每次 `Set/SetByFunc` 未显式传入 `ttl` 会使用全局 TTL。
- `Get` 对已到期键直接返回 `ok=false`；键的物理删除在 `Clear` 时完成。

### ConcurrentMap[T]

- 包级变量可配置：
  - `SHARD_COUNT uint32 = 32`（需在 `NewConcurrentMap` 之前修改）
  - `AUTO_CLEAR_INTERVAL = 10 * time.Minute`
- `NewConcurrentMap[T](options ...func(*Option)) *ConcurrentMap[T]`
- `EnableExpired(ttl time.Duration) func(*Option)`：启用全局 TTL
- `GetShard(key string) uint32`：返回分片索引（FNV32a 哈希）
- `Set(key string, value T, ttl ...time.Duration)`
- `SetByFunc(key string, f func(old T) (new T), ttl ...time.Duration) (new T)`
- `Get(key string) (value T, ok bool)`
- `Delete(key string)`

## 清理机制与到期可见性

- 到期可见性：
  - `Get` 对到期键返回 `ok=false`，即使键尚未被物理删除。
  - 物理删除在 `Clear` 触发：通过小顶堆按过期时间批量 `heap.Pop` 并 `delete`。
- 自动清理：
  - 若开启全局 TTL，`NewConcurrentMap` 会启动一个定时器协程，按 `AUTO_CLEAR_INTERVAL` 周期对每个分片调用 `Clear()`。
  - 若在单次 `Set` 中指定了自定义 `ttl`，也会确保自动清理协程已启动。
- 手动清理：
  - 你可在高写入压力场景下，按批量写入频率定期调用 `Clear()` 以降低内存峰值与堆大小。

## 并发与性能建议

- 多核写多读场景优先使用 `ConcurrentMap`（减少全局锁竞争）。
- 读远多于写、键数较少时可使用 `RWMutexMap`，代码更简单。
- 开启 `-race` 运行测试以验证并发安全：`go test -race ./...`。

## 常见问题（FAQ）

1. 为什么 `Get` 返回 `ok=false` 但键还在内存里？
   - 设计即如此：到期后立刻对外“不可见”，物理删除在 `Clear` 执行，避免每次访问都做 O(logN) 的堆维护。

2. 我需要手动调用 `Clear` 吗？
   - 若你启用了全局 TTL，默认会有定时清理。
   - 追求更低内存/更稳定延迟时，可在写入批次间歇调用 `Clear`。

3. 能否为不同键设置不同 TTL？
   - 可以。`Set/SetByFunc` 的 `ttl ...time.Duration` 会覆盖全局 TTL（若传入且 >0）。

## 运行测试

```bash
go test ./maps -race -v
```

## 许可证

MIT License，参见仓库根目录 `LICENSE`。


