package util

import (
	"math/rand"
	"time"
)

// 返回随机端口
func RandomPort(start, end int, seed int64) int {
	if end > 65535 {
		return 0
	}
	rand.Seed(time.Now().UnixNano() + seed)
	return rand.Intn(end-start) + start
}
