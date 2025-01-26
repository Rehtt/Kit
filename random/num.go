package random

import (
	"crypto/rand"
	"math/big"
	"time"
)

func RandInt(min, max int) int {
	out := RandInt64(int64(min), int64(max))
	return int(out)
}

func RandInt64(min, max int64) int64 {
	if min >= max {
		return min
	}
	return RandBigInt(min, max).Int64()
}

func RandBigInt(min, max int64) *big.Int {
	num, _ := rand.Int(rand.Reader, big.NewInt(max-min+1))
	return num.Add(num, big.NewInt(min))
}

func RandDate(from, to time.Time) time.Time {
	randDate := RandInt64(from.Unix(), to.Unix())
	return time.Unix(randDate, 0)
}
