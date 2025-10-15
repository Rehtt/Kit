# Link 链表模块

链表数据结构的Go实现，包含单向链表排序和双向循环链表两个主要功能。

## 功能特性

### 1. 单向链表归并排序 (MergeSort)
- 泛型支持，适用于任意类型
- 稳定排序算法
- 时间复杂度 O(n log n)
- 空间复杂度 O(1)（原地排序）

### 2. 双向循环链表 (DLink)
- 泛型支持
- 自动扩容机制
- 循环覆盖模式
- 内存池优化
- 非并发安全（需要外部同步）

## 安装使用

```go
import "github.com/rehtt/Kit/link"
```

## API 文档

### 单向链表排序

#### 接口定义

```go
type OnewayLinkInterface[T any] interface {
    Next() OnewayLinkInterface[T]
    SetNext(OnewayLinkInterface[T])
    Val() T
}

type CompareFn[T any] func(a, b T) bool
```

#### 主要函数

```go
func MergeSort[T any](head OnewayLinkInterface[T], less CompareFn[T]) OnewayLinkInterface[T]
```

对链表进行稳定排序，返回排序后的新表头。

**参数：**
- `head`: 链表头节点
- `less`: 比较函数，返回 true 表示 a < b

**返回值：**
- 排序后的链表头节点

### 双向循环链表

#### 结构体定义

```go
type DLink[T any] struct {
    AutoLen  bool           // 自动扩容开关
    OnCover  func(value T)  // 覆盖回调函数
}
```

#### 主要方法

```go
func NewDLink[T any]() *DLink[T]                    // 创建新链表
func (l *DLink[T]) Size(size int64) error           // 设置容量
func (l *DLink[T]) Push(value T)                    // 添加元素
func (l *DLink[T]) Pull() T                         // 取出元素
func (l *DLink[T]) Peek() T                         // 查看顶部元素
func (l *DLink[T]) Range() []T                      // 获取所有元素
func (l *DLink[T]) Len() int64                      // 获取长度
func (l *DLink[T]) Cap() int64                      // 获取容量
```

## 使用示例

### 单向链表排序示例

```go
package main

import (
    "fmt"
    "github.com/rehtt/Kit/link"
)

// 实现链表节点
type Node struct {
    value int
    next  *Node
}

func (n *Node) Next() link.OnewayLinkInterface[int] {
    if n.next == nil {
        return nil
    }
    return n.next
}

func (n *Node) SetNext(next link.OnewayLinkInterface[int]) {
    if next == nil {
        n.next = nil
    } else {
        n.next = next.(*Node)
    }
}

func (n *Node) Val() int {
    return n.value
}

func main() {
    // 创建链表: 3 -> 1 -> 4 -> 2
    head := &Node{value: 3}
    head.next = &Node{value: 1}
    head.next.next = &Node{value: 4}
    head.next.next.next = &Node{value: 2}

    // 定义比较函数
    less := func(a, b int) bool {
        return a < b
    }

    // 排序
    sorted := link.MergeSort[int](head, less)

    // 打印结果: 1 -> 2 -> 3 -> 4
    for p := sorted; p != nil; p = p.Next() {
        fmt.Print(p.Val(), " ")
    }
    fmt.Println()
}
```

### 双向循环链表示例

```go
package main

import (
    "fmt"
    "github.com/rehtt/Kit/link"
)

func main() {
    // 创建链表
    dl := link.NewDLink[int]()
    
    // 设置初始容量
    dl.Size(3)
    
    // 添加元素
    dl.Push(1)
    dl.Push(2)
    dl.Push(3)
    
    fmt.Println("当前元素:", dl.Range()) // [1, 2, 3]
    fmt.Println("长度:", dl.Len())       // 3
    fmt.Println("容量:", dl.Cap())       // 3
    
    // 取出元素 (FIFO)
    fmt.Println("取出:", dl.Pull())      // 1
    fmt.Println("剩余:", dl.Range())     // [2, 3]
    
    // 查看顶部元素
    fmt.Println("顶部:", dl.Peek())      // 2
}
```

### 循环覆盖模式示例

```go
package main

import (
    "fmt"
    "github.com/rehtt/Kit/link"
)

func main() {
    dl := link.NewDLink[string]()
    dl.Size(2)
    dl.AutoLen = false // 关闭自动扩容
    
    // 设置覆盖回调
    dl.OnCover = func(value string) {
        fmt.Printf("被覆盖的值: %s\n", value)
    }
    
    dl.Push("A")
    dl.Push("B")
    dl.Push("C") // 输出: 被覆盖的值: A
    
    fmt.Println("当前元素:", dl.Range()) // [B, C]
}
```

### 自动扩容示例

```go
package main

import (
    "fmt"
    "github.com/rehtt/Kit/link"
)

func main() {
    dl := link.NewDLink[int]()
    dl.Size(2)
    dl.AutoLen = true // 开启自动扩容（默认）
    
    dl.Push(1)
    dl.Push(2)
    fmt.Println("容量:", dl.Cap()) // 2
    
    dl.Push(3) // 触发自动扩容
    fmt.Println("容量:", dl.Cap()) // 7 (2 + 5)
    
    fmt.Println("元素:", dl.Range()) // [1, 2, 3]
}
```

## 性能特点

### 单向链表排序
- **时间复杂度**: O(n log n)
- **空间复杂度**: O(1)
- **稳定性**: 稳定排序
- **适用场景**: 大型链表排序，内存受限环境

### 双向循环链表
- **Push操作**: O(1)
- **Pull操作**: O(1)
- **Range操作**: O(n)
- **内存优化**: 使用对象池减少GC压力
- **适用场景**: 队列、缓冲区、循环缓存

## 注意事项

1. **线程安全**: 双向循环链表不是并发安全的，多线程环境下需要外部同步
2. **内存管理**: 链表使用内存池优化，自动回收节点对象
3. **自动扩容**: 默认开启，每次扩容增加5个节点
4. **覆盖模式**: 关闭自动扩容后，新元素会覆盖最旧的元素

## 测试

运行测试：

```bash
go test ./link
```

运行基准测试：

```bash
go test -bench=. ./link
```
