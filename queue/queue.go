package queue

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"github.com/Rehtt/Kit/channel"
	"sync"
	"time"
)

type Queue struct {
	getout       sync.Map
	queue        *channel.Chan
	DeadlineFunc func(queue *Queue, id string, data any, deadline time.Time)
}

type Node struct {
	Id       string
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
	q := &Queue{
		queue:        channel.New(),
		DeadlineFunc: DefaultDeadlineFunc(),
	}
	go func() {
		for {
			time.Sleep(scanTime)
			q.getout.Range(func(key, value any) bool {
				if v, ok := value.(*Node); ok && v.Deadline != nil && v.Deadline.Sub(time.Now()) < 0 {
					q.Done(v.Id)
					if q.DeadlineFunc != nil {
						q.DeadlineFunc(q, v.Id, v.Data, *v.Deadline)
					}
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
func (q *Queue) Get(deadline *time.Time, block ...bool) (id string, data any, ok bool) {
	var node *Node
	defer func() {
		if ok {
			id = node.Id
			data = node.Data
			if deadline != nil {
				node.Deadline = deadline
				q.getout.Store(node.Id, node)
			}
		}
	}()
	if len(block) > 0 && block[0] {
		n := <-q.queue.Out
		node, ok = n.(*Node)
		return
	}
	select {
	case n := <-q.queue.Out:
		node, ok = n.(*Node)
		return
	default:
		return "", nil, false
	}

}

// Put
//
//	@Description: 推入队列
//	@receiver q
//	@param data
func (q *Queue) Put(data any) {
	q.queue.In <- newNode(data)
}

// Done
//
//	@Description: 消息确认
//	@receiver q
//	@param id
func (q *Queue) Done(id string) {
	q.getout.Delete(id)
}
func newNode(data any) *Node {
	node := nodePool.Get().(*Node)
	var tmp = make([]byte, 512)
	rand.Read(tmp)
	s := sha256.New()
	s.Write(tmp)
	node.Id = hex.EncodeToString(s.Sum(nil))
	node.Deadline = nil
	node.Data = data
	return node
}

// DefaultDeadlineFunc
//
//	@Description: 默认消息超时未确认处理，将超时任务重新退回队列
//	@return func(queue *Queue, id string, data any, deadline time.Time)
func DefaultDeadlineFunc() func(queue *Queue, id string, data any, deadline time.Time) {
	return func(queue *Queue, id string, data any, deadline time.Time) {
		queue.Put(data)
	}
}
