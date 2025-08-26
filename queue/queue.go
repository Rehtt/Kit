package queue

import (
	"context"
	"crypto/rand"
	"hash/fnv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Rehtt/Kit/channel"
)

type Queue struct {
	getout       sync.Map
	queue        *channel.Chan[*Node]
	DeadlineFunc func(queue *Queue, id uint64, data any, deadline time.Time)
	scanInterval time.Duration
	close        atomic.Bool
}

type Node struct {
	Id       uint64
	Data     any
	Deadline *time.Time
}

var (
	nodePool = sync.Pool{
		New: func() any {
			return new(Node)
		},
	}
	scanTime = 5 * time.Minute
)

// NewQueue
//
//	@Description: 创建消息队列
//	@return *Queue
func NewQueue() *Queue {
	return NewQueueWithOptions(scanTime, DefaultDeadlineFunc())
}

// NewQueueWithOptions
//
//	@Description: 创建可配置扫描周期与超时回调的消息队列
//	@param scanInterval 扫描超时消息的周期，<=0 则使用默认值
//	@param deadlineFunc 可选：自定义超时回调，不传则使用默认回退策略
//	@return *Queue
func NewQueueWithOptions(scanInterval time.Duration, deadlineFunc ...func(queue *Queue, id uint64, data any, deadline time.Time)) *Queue {
	q := &Queue{
		queue: channel.New[*Node](),
	}
	if scanInterval <= 0 {
		q.scanInterval = scanTime
	} else {
		q.scanInterval = scanInterval
	}
	if len(deadlineFunc) > 0 && deadlineFunc[0] != nil {
		q.DeadlineFunc = deadlineFunc[0]
	} else {
		q.DeadlineFunc = DefaultDeadlineFunc()
	}
	go func() {
		for !q.close.Load() {
			time.Sleep(q.scanInterval)
			q.getout.Range(func(key, value any) bool {
				if v, ok := value.(*Node); ok && v.Deadline != nil && time.Until(*v.Deadline) < 0 {
					// 超时：从待确认集合移除，执行回调，并回收节点
					q.getout.Delete(v.Id)
					if q.DeadlineFunc != nil {
						q.DeadlineFunc(q, v.Id, v.Data, *v.Deadline)
					}
					nodePool.Put(v)
				}
				return true
			})
		}
	}()
	return q
}

// Get
//
//	@Description:	接收
//	@receiver q
//	@param deadline	消息确认超时，设置非nil后需要使用Done()进行消息确认
//	@param block	阻塞
//	@return id		队列id
//	@return data	内容
//	@return ok		是否获取到
func (q *Queue) Get(ctx context.Context, deadline *time.Time, block ...bool) (id uint64, data any, ok bool) {
	if q.close.Load() {
		return
	}
	var node *Node
	defer func() {
		if ok {
			id = node.Id
			data = node.Data
			if deadline != nil {
				// 拷贝时间值，避免持有调用方指针
				t := *deadline
				node.Deadline = &t
				q.getout.Store(node.Id, node)
			} else {
				// 非确认模式，立即回收节点
				nodePool.Put(node)
			}
		}
	}()
	if len(block) > 0 && block[0] {
		select {
		case <-ctx.Done():
		case node, ok = <-q.queue.Out:
		}
		return
	}
	select {
	case <-ctx.Done():
	case node, ok = <-q.queue.Out:
	default:
	}
	return
}

// Put
//
//	@Description: 推入队列
//	@receiver q
//	@param data
func (q *Queue) Put(data any) {
	if q.close.Load() {
		return
	}
	q.queue.In <- newNode(data)
}

// Done
//
//	@Description: 消息确认
//	@receiver q
//	@param id
func (q *Queue) Done(id uint64) {
	if v, ok := q.getout.LoadAndDelete(id); ok {
		if n, ok := v.(*Node); ok {
			nodePool.Put(n)
		}
	}
}

// DoneAll
//
//	@Description: 清空队列
func (q *Queue) DoneAll() {
	q.getout.Range(func(key, value any) bool {
		if n, ok := value.(*Node); ok {
			nodePool.Put(n)
		}
		q.getout.Delete(key)
		return true
	})
}

// Close
//
//	@Description: 关闭
func (q *Queue) Close() {
	if q.close.CompareAndSwap(false, true) {
		q.DoneAll()
		q.queue.Close()
	}
}

func newNode(data any) *Node {
	node := nodePool.Get().(*Node)
	tmp := make([]byte, 64)
	rand.Read(tmp)
	s := fnv.New64a()
	s.Write(tmp)
	node.Id = s.Sum64()
	node.Deadline = nil
	node.Data = data
	return node
}

// DefaultDeadlineFunc
//
//	@Description: 默认消息超时未确认处理，将超时任务重新退回队列
//	@return func(queue *Queue, id uint64, data any, deadline time.Time)
func DefaultDeadlineFunc() func(queue *Queue, id uint64, data any, deadline time.Time) {
	return func(queue *Queue, id uint64, data any, deadline time.Time) {
		queue.Put(data)
	}
}
