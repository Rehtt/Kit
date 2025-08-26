package queue

import (
	"context"
	"testing"
	"time"

	"github.com/Rehtt/Kit/channel"
)

func TestGetNonBlockingEmpty(t *testing.T) {
	q := NewQueue()
	ctx := context.Background()
	id, data, ok := q.Get(ctx, nil)
	if ok || id != 0 || data != nil {
		t.Error("expected no data; got something")
	}
}

func TestPutAndGet(t *testing.T) {
	q := NewQueue()
	ctx := context.Background()
	data := "test-data"
	go q.Put(data)
	id, got, ok := q.Get(ctx, nil, true)
	if !ok {
		t.Fatal("expected ok true")
	}
	if got.(string) != data {
		t.Errorf("expected data %v; got %v", data, got)
	}
	if id == 0 {
		t.Errorf("expected id != 0; got %v", id)
	}
}

func TestGetWithDeadlineAndDone(t *testing.T) {
	q := NewQueue()
	ctx := context.Background()
	data := "deadline-data"
	go q.Put(data)
	deadline := time.Now().Add(100 * time.Millisecond)
	id, got, ok := q.Get(ctx, &deadline, true)
	if !ok {
		t.Fatal("expected ok true")
	}
	if got.(string) != data {
		t.Errorf("expected data %v; got %v", data, got)
	}
	q.Done(id)
	var found bool
	q.getout.Range(func(k, v any) bool {
		if k.(uint64) == id {
			found = true
			return false
		}
		return true
	})
	if found {
		t.Error("expected id to be removed from getout")
	}
}

func TestDefaultDeadlineFunc(t *testing.T) {
	q := &Queue{queue: channel.New[*Node]()}
	fn := DefaultDeadlineFunc()
	data := "default-data"
	fn(q, 1, data, time.Now())
	ctx := context.Background()
	id, got, ok := q.Get(ctx, nil, true)
	if !ok {
		t.Fatal("expected ok true")
	}
	if got.(string) != data {
		t.Errorf("expected data %v; got %v", data, got)
	}
	if id == 0 {
		t.Error("expected non-empty id")
	}
}

func TestScanRequeue(t *testing.T) {
	old := scanTime
	scanTime = 10 * time.Millisecond
	defer func() { scanTime = old }()
	q := NewQueue()
	ctx := context.Background()
	data := "scan-data"
	go q.Put(data)
	past := time.Now().Add(-10 * time.Millisecond)
	_, got, ok := q.Get(ctx, &past, true)
	if !ok {
		t.Fatal("expected ok true")
	}
	if got.(string) != data {
		t.Errorf("expected %v; got %v", data, got)
	}
	// wait for requeue
	time.Sleep(50 * time.Millisecond)
	_, got2, ok2 := q.Get(ctx, nil, true)
	if !ok2 {
		t.Fatal("expected ok true after scan requeue")
	}
	if got2.(string) != data {
		t.Errorf("expected %v; got %v", data, got2)
	}
}

func BenchmarkNewNode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = newNode("benchmark")
	}
}

func BenchmarkQueuePutGet(b *testing.B) {
	q := NewQueue()
	ctx := context.Background()
	data := "benchmark-data"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go q.Put(data)
		q.Get(ctx, nil, true)
	}
}

func BenchmarkParallelQueuePutGet(b *testing.B) {
	q := NewQueue()
	ctx := context.Background()
	data := "benchmark-data"
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			go q.Put(data)
			q.Get(ctx, nil, true)
		}
	})
}

func TestCloseBehavior(t *testing.T) {
	q := NewQueue()
	ctx := context.Background()
	// 关闭后：Get 返回空，Put 无效
	q.Close()
	q.Put("after-close")
	id, data, ok := q.Get(ctx, nil)
	if ok || id != 0 || data != nil {
		t.Error("expected empty result after Close")
	}
}

func TestDoneAll(t *testing.T) {
	q := NewQueue()
	ctx := context.Background()
	// 放入两条需要确认的消息
	go q.Put("a")
	go q.Put("b")
	dl := time.Now().Add(10 * time.Second)
	_, _, ok1 := q.Get(ctx, &dl, true)
	if !ok1 {
		t.Fatal("expected first ok true")
	}
	_, _, ok2 := q.Get(ctx, &dl, true)
	if !ok2 {
		t.Fatal("expected second ok true")
	}
	// 确认待处理集合非空
	var had bool
	q.getout.Range(func(k, v any) bool {
		had = true
		return false
	})
	if !had {
		t.Fatal("expected getout not empty before DoneAll")
	}
	// 清理
	q.DoneAll()
	var found bool
	q.getout.Range(func(k, v any) bool {
		found = true
		return false
	})
	if found {
		t.Fatal("expected getout empty after DoneAll")
	}
}

func TestBlockingGetCancelable(t *testing.T) {
	q := NewQueue()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	id, data, ok := q.Get(ctx, nil, true)
	if ok || id != 0 || data != nil {
		t.Error("expected canceled blocking get to return empty result")
	}
}

func TestCustomDeadlineFuncCalled(t *testing.T) {
	old := scanTime
	scanTime = 10 * time.Millisecond
	defer func() { scanTime = old }()

	q := NewQueue()
	called := make(chan struct{}, 1)
	q.DeadlineFunc = func(_ *Queue, _ uint64, _ any, _ time.Time) {
		select {
		case called <- struct{}{}:
		default:
		}
	}

	ctx := context.Background()
	go q.Put("x")
	past := time.Now().Add(-10 * time.Millisecond)
	_, _, ok := q.Get(ctx, &past, true)
	if !ok {
		t.Fatal("expected ok true")
	}

	select {
	case <-called:
		// ok
	case <-time.After(300 * time.Millisecond):
		t.Fatal("expected custom DeadlineFunc to be called")
	}
}
