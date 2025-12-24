package util

import (
	"errors"
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
	timeMaskBit uint
	// 自增序列，默认10bit
	counterBit  uint
	counterMask int64

	autoIncrement atomic.Int64
}

// NewSnowflake 创建一个新的 Snowflake
//
//	baseTime: 基准时间，用于计算时间差，必须小于当前时间
//	logicalId: 逻辑id，可以是部署的机器id，在集群中应该是唯一的，默认13bit
func NewSnowflake(baseTime time.Time, logicalId int) (*Snowflake, error) {
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

		timeMaskBit:  logicalIdBit + counterBit,
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
		curMs := cur >> s.timeMaskBit
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
				time.Sleep(time.Microsecond)
				continue
			}
			next = (milliseconds << s.timeMaskBit) | s.logicalId | (curCnt + 1)
		} else {
			// 新的毫秒，计数从 1 开始
			next = (milliseconds << s.timeMaskBit) | s.logicalId | 1
		}

		if s.autoIncrement.CompareAndSwap(cur, next) {
			return next
		}
	}
}

// ParseInfo 将 id 解析成时间、逻辑id、计数
func (s *Snowflake) ParseInfo(id int64) (milliseconds time.Duration, logicalId int64, counter int64) {
	milliseconds = time.Duration((id >> s.timeMaskBit)) * time.Millisecond
	logicalId = (id >> s.counterBit) & (1<<s.logicalIdBit - 1)
	counter = id & s.counterMask
	return
}
