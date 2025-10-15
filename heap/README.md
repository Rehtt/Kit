# Heap 堆模块

高性能的堆数据结构Go实现，基于链表结构，支持内存池优化的FIFO队列操作。

## 功能特性

- 🚀 高性能FIFO队列操作
- 💾 内存池优化，减少GC压力  
- 📊 实时长度和元素计数
- 🔄 自动内存回收机制
- 🎯 简单易用的API
- ⚡ O(1)时间复杂度的Push/Pop操作

## 安装使用

```go
import "github.com/rehtt/Kit/heap"
```

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/rehtt/Kit/heap"
)

func main() {
    // 创建新的堆
    h := heap.NewHeap()
    
    // 添加元素
    h.Push(1)
    h.Push("hello")
    h.Push(3.14)
    
    fmt.Println("长度:", h.Len()) // 输出: 长度: 3
    
    // 取出元素 (FIFO顺序)
    fmt.Println(h.Pop()) // 输出: 1
    fmt.Println(h.Pop()) // 输出: hello
    fmt.Println(h.Pop()) // 输出: 3.14
    fmt.Println(h.Pop()) // 输出: <nil>
}
```

## API 文档

### 结构体

```go
type Heap struct {
    // 私有字段，通过方法访问
}
```

### 构造函数

#### NewHeap
```go
func NewHeap() *Heap
```
创建一个新的堆实例。

**返回值：**
- `*Heap`: 新创建的堆实例

**示例：**
```go
h := heap.NewHeap()
```

### 核心方法

#### Push
```go
func (h *Heap) Push(data any)
```
向堆中添加一个元素。

**参数：**
- `data`: 要添加的数据，可以是任意类型

**时间复杂度：** O(1)

**示例：**
```go
h.Push(42)
h.Push("world")
h.Push([]int{1, 2, 3})
```

#### Pop
```go
func (h *Heap) Pop() any
```
从堆中取出一个元素（FIFO顺序）。

**返回值：**
- `any`: 取出的元素，如果堆为空则返回 `nil`

**时间复杂度：** O(1)

**示例：**
```go
element := h.Pop()
if element != nil {
    fmt.Println("取出元素:", element)
} else {
    fmt.Println("堆为空")
}
```

#### Len
```go
func (h *Heap) Len() int32
```
获取堆中当前元素的数量。

**返回值：**
- `int32`: 当前元素数量

**示例：**
```go
fmt.Printf("当前有 %d 个元素\n", h.Len())
```

#### CountElem
```go
func (h *Heap) CountElem() int32
```
获取堆中总的节点数量（包括已回收的节点）。

**返回值：**
- `int32`: 总节点数量

**示例：**
```go
fmt.Printf("总共分配了 %d 个节点\n", h.CountElem())
```

## 使用示例

### 基本队列操作

```go
package main

import (
    "fmt"
    "github.com/rehtt/Kit/heap"
)

func main() {
    h := heap.NewHeap()
    
    // 添加不同类型的数据
    h.Push(1)
    h.Push("hello")
    h.Push([]int{1, 2, 3})
    h.Push(map[string]int{"key": 42})
    
    fmt.Printf("堆长度: %d\n", h.Len()) // 输出: 堆长度: 4
    
    // 按FIFO顺序取出
    for h.Len() > 0 {
        element := h.Pop()
        fmt.Printf("取出: %v (类型: %T)\n", element, element)
    }
}
```

### 任务队列示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/rehtt/Kit/heap"
)

type Task struct {
    ID      int
    Name    string
    Created time.Time
}

func (t Task) String() string {
    return fmt.Sprintf("Task{ID: %d, Name: %s}", t.ID, t.Name)
}

func main() {
    taskQueue := heap.NewHeap()
    
    // 添加任务
    tasks := []Task{
        {ID: 1, Name: "处理订单", Created: time.Now()},
        {ID: 2, Name: "发送邮件", Created: time.Now()},
        {ID: 3, Name: "生成报告", Created: time.Now()},
    }
    
    for _, task := range tasks {
        taskQueue.Push(task)
        fmt.Printf("添加任务: %s\n", task)
    }
    
    fmt.Printf("\n队列长度: %d\n\n", taskQueue.Len())
    
    // 处理任务
    for taskQueue.Len() > 0 {
        task := taskQueue.Pop().(Task)
        fmt.Printf("处理任务: %s\n", task)
        
        // 模拟任务处理时间
        time.Sleep(100 * time.Millisecond)
    }
    
    fmt.Println("\n所有任务处理完成")
}
```

### 缓冲区示例

```go
package main

import (
    "fmt"
    "github.com/rehtt/Kit/heap"
)

type Buffer struct {
    heap     *heap.Heap
    maxSize  int32
    overflow func(data any) // 溢出处理函数
}

func NewBuffer(maxSize int32) *Buffer {
    return &Buffer{
        heap:    heap.NewHeap(),
        maxSize: maxSize,
    }
}

func (b *Buffer) SetOverflowHandler(handler func(data any)) {
    b.overflow = handler
}

func (b *Buffer) Write(data any) {
    if b.heap.Len() >= b.maxSize {
        // 缓冲区满，移除最旧的元素
        old := b.heap.Pop()
        if b.overflow != nil {
            b.overflow(old)
        }
    }
    b.heap.Push(data)
}

func (b *Buffer) Read() any {
    return b.heap.Pop()
}

func (b *Buffer) Size() int32 {
    return b.heap.Len()
}

func main() {
    buffer := NewBuffer(3)
    
    // 设置溢出处理
    buffer.SetOverflowHandler(func(data any) {
        fmt.Printf("缓冲区溢出，丢弃: %v\n", data)
    })
    
    // 写入数据
    for i := 1; i <= 5; i++ {
        buffer.Write(fmt.Sprintf("数据%d", i))
        fmt.Printf("写入数据%d，缓冲区大小: %d\n", i, buffer.Size())
    }
    
    fmt.Println("\n读取缓冲区数据:")
    for buffer.Size() > 0 {
        data := buffer.Read()
        fmt.Printf("读取: %v\n", data)
    }
}
```

### 生产者-消费者模式

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/rehtt/Kit/heap"
)

func main() {
    h := heap.NewHeap()
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    // 生产者
    wg.Add(1)
    go func() {
        defer wg.Done()
        for i := 1; i <= 10; i++ {
            mu.Lock()
            h.Push(fmt.Sprintf("消息%d", i))
            fmt.Printf("生产: 消息%d\n", i)
            mu.Unlock()
            time.Sleep(100 * time.Millisecond)
        }
    }()
    
    // 消费者
    wg.Add(1)
    go func() {
        defer wg.Done()
        for {
            mu.Lock()
            if h.Len() > 0 {
                msg := h.Pop()
                fmt.Printf("消费: %v\n", msg)
                mu.Unlock()
            } else {
                mu.Unlock()
                time.Sleep(50 * time.Millisecond)
            }
            
            // 简单的退出条件
            mu.Lock()
            isEmpty := h.Len() == 0
            mu.Unlock()
            
            if isEmpty {
                time.Sleep(200 * time.Millisecond)
                mu.Lock()
                stillEmpty := h.Len() == 0
                mu.Unlock()
                if stillEmpty {
                    break
                }
            }
        }
    }()
    
    wg.Wait()
    fmt.Println("生产消费完成")
}
```

### 内存使用监控

```go
package main

import (
    "fmt"
    "github.com/rehtt/Kit/heap"
)

func main() {
    h := heap.NewHeap()
    
    fmt.Println("=== 内存使用监控 ===")
    
    // 添加元素
    fmt.Println("\n1. 添加元素:")
    for i := 1; i <= 5; i++ {
        h.Push(i)
        fmt.Printf("添加 %d - 长度: %d, 总节点: %d\n", 
            i, h.Len(), h.CountElem())
    }
    
    // 取出部分元素
    fmt.Println("\n2. 取出部分元素:")
    for i := 0; i < 3; i++ {
        val := h.Pop()
        fmt.Printf("取出 %v - 长度: %d, 总节点: %d\n", 
            val, h.Len(), h.CountElem())
    }
    
    // 再次添加元素（观察节点复用）
    fmt.Println("\n3. 再次添加元素（节点复用）:")
    for i := 6; i <= 8; i++ {
        h.Push(i)
        fmt.Printf("添加 %d - 长度: %d, 总节点: %d\n", 
            i, h.Len(), h.CountElem())
    }
    
    // 清空堆
    fmt.Println("\n4. 清空堆:")
    for h.Len() > 0 {
        val := h.Pop()
        fmt.Printf("取出 %v - 长度: %d, 总节点: %d\n", 
            val, h.Len(), h.CountElem())
    }
}
```

## 性能特点

### 时间复杂度
- **Push操作**: O(1) - 常数时间添加
- **Pop操作**: O(1) - 常数时间取出  
- **Len操作**: O(1) - 常数时间获取长度
- **CountElem操作**: O(1) - 常数时间获取总节点数

### 空间复杂度
- **存储空间**: O(n) - n为元素数量
- **内存优化**: 使用内存池减少GC压力
- **节点复用**: 自动回收和复用节点对象

### 性能优势
1. **内存池**: 减少频繁的内存分配和回收
2. **原子操作**: 使用原子计数器提高并发性能
3. **零拷贝**: 直接操作指针，避免数据拷贝
4. **GC友好**: 减少垃圾回收压力

## 适用场景

- ✅ **消息队列**: FIFO消息处理
- ✅ **任务调度**: 按顺序处理任务
- ✅ **缓冲区**: 临时数据存储
- ✅ **生产者-消费者**: 解耦生产和消费
- ✅ **事件处理**: 按顺序处理事件
- ✅ **数据流**: 流式数据处理

## 注意事项

1. **线程安全**: 该实现不是线程安全的，多线程环境需要外部同步
2. **内存管理**: 使用内存池优化，节点会被自动回收和复用
3. **数据类型**: 支持任意类型的数据存储
4. **空堆操作**: Pop空堆会返回nil，需要检查返回值
5. **内存泄漏**: 长期持有大对象可能导致内存无法回收

## 与标准库对比

| 特性 | Kit/heap | container/list | channel |
|------|----------|----------------|---------|
| 类型安全 | interface{} | interface{} | 泛型支持 |
| 内存优化 | ✅ 内存池 | ❌ | ✅ |
| 并发安全 | ❌ | ❌ | ✅ |
| 性能 | 高 | 中等 | 高 |
| 使用复杂度 | 简单 | 中等 | 简单 |

## 测试

运行测试：

```bash
go test ./heap
```

运行基准测试：

```bash
go test -bench=. ./heap
```
