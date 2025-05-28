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
	if ok || id != "" || data != nil {
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
	if len(id) != 64 {
		t.Errorf("expected id length 64; got %d", len(id))
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
		if k.(string) == id {
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
	q := &Queue{queue: channel.New()}
	fn := DefaultDeadlineFunc()
	data := "default-data"
	fn(q, "ignored-id", data, time.Now())
	ctx := context.Background()
	id, got, ok := q.Get(ctx, nil, true)
	if !ok {
		t.Fatal("expected ok true")
	}
	if got.(string) != data {
		t.Errorf("expected data %v; got %v", data, got)
	}
	if id == "" {
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
