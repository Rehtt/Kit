package util

import (
	"errors"
	"strconv"
	"sync/atomic"
	"time"
)

type Snowflake struct {
	// 基准时间
	BaseTime time.Time
	// 逻辑id，可以是部署的机器id
	LogicalId int64

	// 所有bit总和不能超过64
	// 毫秒时间所占长度，默认41bit 大概可以在基准时间上用69年左右
	BaseTimeBit int64
	// 默认13bit
	LogicalIdBit int64
	// 自增序列，默认10bit
	AutoIncrementBit int64

	milliseconds  atomic.Int64
	autoIncrement atomic.Int64
}

func (s *Snowflake) handle() error {
	atomic.CompareAndSwapInt64(&s.BaseTimeBit, 0, 41)
	atomic.CompareAndSwapInt64(&s.LogicalIdBit, 0, 13)
	atomic.CompareAndSwapInt64(&s.AutoIncrementBit, 0, 10)
	if s.BaseTimeBit+s.LogicalIdBit+s.AutoIncrementBit > 64 {
		return errors.New("所有bit总和不能超过64")
	}
	if s.LogicalId >= (1 << s.LogicalIdBit) {
		return errors.New("LogicalId不能大于(1 << LogicalIdBit)=" + strconv.Itoa(1<<s.LogicalIdBit))
	}
	return nil
}

func (s *Snowflake) GenerateId() (int64, error) {
	if err := s.handle(); err != nil {
		return 0, err
	}

	milliseconds := time.Since(s.BaseTime).Milliseconds()
	// 根据毫秒重置自增序列，并保证线程安全
	if m := s.milliseconds.Load(); m != milliseconds {
		if s.milliseconds.CompareAndSwap(m, milliseconds) {
			s.autoIncrement.Store(0)
		}
	}
	autoIncrement := s.autoIncrement.Add(1)

	var out int64
	out = milliseconds << (64 - s.BaseTimeBit)
	out |= (s.LogicalId << (64 - s.BaseTimeBit - s.LogicalIdBit))
	out |= (autoIncrement % (1 << s.AutoIncrementBit))
	return out, nil
}
