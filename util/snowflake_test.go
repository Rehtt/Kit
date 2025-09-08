package util

import (
	"sync"
	"testing"
	"time"
)

// helper: 快速构造一个可调位宽的 Snowflake，用于测试（直接构造 struct，测试在同包）
func newTestSnowflake(baseTime time.Time, logicalId int64, logicalBits uint, counterBits uint) *Snowflake {
	s := &Snowflake{
		baseTime:     baseTime,
		logicalIdBit: logicalBits,
		counterBit:   counterBits,
		timeBit:      logicalBits + counterBits,
		counterMask:  int64((1 << counterBits) - 1),
		logicalId:    logicalId << counterBits,
	}
	return s
}

func TestNewSnowflakeLogicalIdBoundaries(t *testing.T) {
	// 用于与 NewSnowflake 一致的常量
	const logicalIdBits uint = 13
	max := 1<<logicalIdBits - 1

	// 合法
	if _, err := NewSnowflake(time.Now().Add(-time.Second), 0); err != nil {
		t.Fatalf("unexpected error for logicalId=0: %v", err)
	}

	// 达到边界：等于 2^13 应该报错
	if _, err := NewSnowflake(time.Now().Add(-time.Second), max+1); err == nil {
		t.Fatalf("expected error for logicalId >= 2^13, got nil")
	}
}

func TestNewSnowflakeBaseTimeInFuture(t *testing.T) {
	_, err := NewSnowflake(time.Now().Add(time.Hour), 1) // baseTime 在未来
	if err == nil {
		t.Fatalf("expected error when baseTime is in the future, got nil")
	}
}

func TestGenerateIdSingleThread_UniqueMonotonic(t *testing.T) {
	s := newTestSnowflake(time.Now().Add(-time.Second), 7, 13, 10)

	const n = 5000
	ids := make([]int64, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, s.GenerateId())
	}

	// 唯一性
	seen := make(map[int64]struct{}, n)
	for i, id := range ids {
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate id at index %d: %d", i, id)
		}
		seen[id] = struct{}{}
	}

	// 单调性（严格递增）
	for i := 1; i < len(ids); i++ {
		if ids[i] <= ids[i-1] {
			t.Fatalf("ids not strictly increasing at %d: %d <= %d", i, ids[i], ids[i-1])
		}
	}
}

func TestGenerateIdConcurrent_Unique(t *testing.T) {
	s := newTestSnowflake(time.Now().Add(-time.Second), 123, 13, 10)

	const goroutines = 50
	const perG = 200 // 每个 goroutine 生成的 id 数量（总共 10k）
	total := goroutines * perG

	var wg sync.WaitGroup
	ch := make(chan int64, total)

	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				ch <- s.GenerateId()
			}
		}()
	}

	wg.Wait()
	close(ch)

	seen := make(map[int64]struct{}, total)
	count := 0
	for id := range ch {
		count++
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate id found: %d", id)
		}
		seen[id] = struct{}{}
	}
	if count != total {
		t.Fatalf("expected %d ids, got %d", total, count)
	}
}

func TestParseInfoConsistency(t *testing.T) {
	base := time.Now().Add(-time.Second)
	logical := int64(42)
	s := newTestSnowflake(base, logical, 13, 10)

	id := s.GenerateId()
	msDur, parsedLogical, cnt := s.ParseInfo(id)

	// logicalId must match
	if parsedLogical != logical {
		t.Fatalf("parsed logicalId mismatch: want %d got %d", logical, parsedLogical)
	}

	// counter must be >= 0 and <= mask
	if cnt < 0 || cnt > s.counterMask {
		t.Fatalf("parsed counter out of range: %d", cnt)
	}

	// timestamp: 因为生成后有短时间延迟，允许小范围差异（比如 <= 5ms）
	nowMs := time.Since(base)
	diff := nowMs - msDur
	if diff < 0 {
		diff = -diff
	}
	if diff > 5*time.Millisecond {
		t.Fatalf("parsed milliseconds not close enough: parsed=%v now=%v diff=%v", msDur, nowMs, diff)
	}
}

func TestCounterRolloverBehavior(t *testing.T) {
	// 使用很小的 counterBits（2 bits -> mask=3）来强制快速耗尽
	const logicalBits uint = 3
	const counterBits uint = 2 // 只能 4 个 counter 值（0..3 或 1..4 取决实现）
	base := time.Now().Add(-time.Second)
	logical := int64(1)
	s := newTestSnowflake(base, logical, logicalBits, counterBits)

	// 生成超过 counterMask 的数量，观察是否能继续产生唯一且时间位上升
	genCount := int(s.counterMask) + 3 // 比可用 counter 多几个
	ids := make([]int64, 0, genCount)
	for i := 0; i < genCount; i++ {
		ids = append(ids, s.GenerateId())
	}

	// 唯一性
	seen := make(map[int64]struct{}, genCount)
	for i, id := range ids {
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate id at index %d: %d", i, id)
		}
		seen[id] = struct{}{}
	}

	// 找到第一个 id 的 ms 部分和在计数耗尽后某个 id 的 ms 部分，确保在耗尽后 ms 有上升（说明推进到了下一毫秒）
	firstMs, _, _ := s.ParseInfo(ids[0])
	// 找到最后一个的 ms
	lastMs, _, _ := s.ParseInfo(ids[len(ids)-1])

	if lastMs <= firstMs {
		t.Fatalf("expected lastMs > firstMs after rollover, but lastMs=%v firstMs=%v", lastMs, firstMs)
	}
}

func BenchmarkGenerateId(b *testing.B) {
	s := newTestSnowflake(time.Now().Add(-time.Second), 7, 13, 10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = s.GenerateId()
		}
	})
}
