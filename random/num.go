package random

import (
	"crypto/rand"
	"math/big"
	mrand "math/rand/v2"
	"time"
)

// unsafe 为true时，使用伪随机
func RandInt(min, max int, unsafe ...bool) int {
	if min >= max {
		return min
	}
	if len(unsafe) > 0 && unsafe[0] {
		return min + mrand.N(max-min+1)
	}
	out := RandInt64(int64(min), int64(max))
	return int(out)
}

// unsafe 为true时，使用伪随机
func RandInt64(min, max int64, unsafe ...bool) int64 {
	if min >= max {
		return min
	}
	if len(unsafe) > 0 && unsafe[0] {
		return min + mrand.N(max-min+1)
	}
	return RandBigInt(min, max).Int64()
}

func RandBigInt(min, max int64) *big.Int {
	num, _ := rand.Int(rand.Reader, big.NewInt(max-min+1))
	return num.Add(num, big.NewInt(min))
}

func RandDate(from, to time.Time, unsafe ...bool) time.Time {
	randDate := RandInt64(from.Unix(), to.Unix(), unsafe...)
	return time.Unix(randDate, 0)
}
