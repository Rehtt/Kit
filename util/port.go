package util

import (
	"github.com/Rehtt/Kit/random"
)

const MaxPort uint16 = 1<<16 - 1

// 返回随机端口
func RandomPort(start, end uint16) uint16 {
	port := random.RandInt64(int64(start), int64(end))
	return uint16(port)
}
