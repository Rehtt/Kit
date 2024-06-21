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
	s.autoIncrement.Add(1)
	return nil
}

func (s *Snowflake) GenerateId() (int64, error) {
	if err := s.handle(); err != nil {
		return 0, err
	}
	var out int64
	out = time.Now().Sub(s.BaseTime).Milliseconds() << (64 - s.BaseTimeBit)
	out |= (s.LogicalId << (64 - s.BaseTimeBit - s.LogicalIdBit))
	out |= (s.autoIncrement.Load() % (1 << s.AutoIncrementBit))
	return out, nil
}
