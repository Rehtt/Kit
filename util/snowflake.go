package util

import (
	"errors"
	"runtime"
	"sync/atomic"
	"time"
)

type Snowflake struct {
	// 基准时间
	baseTime time.Time
	// 逻辑id，可以是部署的机器id
	// 默认13bit
	logicalId    int64
	logicalIdBit uint

	// 所有bit总和不能超过64
	// 毫秒时间所占长度，默认41bit 大概可以在基准时间上用69年左右
	timeBit uint
	// 自增序列，默认10bit
	counterBit  uint
	counterMask int64

	autoIncrement atomic.Int64
}

func NewSnowflake(baseTime time.Time, logicalId uint) (*Snowflake, error) {
	var (
		// baseTimeBit  int64 = 41
		logicalIdBit uint = 13
		counterBit   uint = 10
	)

	if baseTime.After(time.Now()) {
		return nil, errors.New("baseTime 必须小于当前时间")
	}
	if logicalId < 0 || logicalId >= 1<<logicalIdBit {
		return nil, errors.New("logicalId 大小超出范围")
	}

	return &Snowflake{
		baseTime: baseTime,

		timeBit:      logicalIdBit + counterBit,
		logicalIdBit: logicalIdBit,
		counterBit:   counterBit,

		counterMask: 1<<counterBit - 1,

		logicalId: int64(logicalId) << counterBit,
	}, nil
}

func (s *Snowflake) GenerateId() int64 {
	var milliseconds int64
	for {
		milliseconds = time.Since(s.baseTime).Milliseconds()
		cur := s.autoIncrement.Load()
		curMs := cur >> s.timeBit
		curCnt := cur & s.counterMask

		var next int64
		if milliseconds < curMs {
			// 时钟短回退：保持不回退（用 curMs），避免重复
			milliseconds = curMs
		}

		if milliseconds == curMs {
			// 同一毫秒，尝试 ++
			if curCnt >= s.counterMask {
				// 计数耗尽：等待下一毫秒再试
				// 优化调度，释放当前的调度
				runtime.Gosched()
				time.Sleep(time.Microsecond)
				continue
			}
			next = (milliseconds << s.timeBit) | s.logicalId | (curCnt + 1)
		} else {
			// 新的毫秒，计数从 1 开始
			next = (milliseconds << s.timeBit) | s.logicalId | 1
		}

		if s.autoIncrement.CompareAndSwap(cur, next) {
			return next
		}
	}
}

func (s *Snowflake) ParseInfo(id int64) (milliseconds time.Duration, logicalId int64, counter int64) {
	milliseconds = time.Duration((id >> s.timeBit)) * time.Millisecond
	logicalId = (id >> s.counterBit) & (1<<s.logicalIdBit - 1)
	counter = id & s.counterMask
	return
}
