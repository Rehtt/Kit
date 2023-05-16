package util

import (
	"crypto/sha256"
	"time"
)

func TimeToPrt(t time.Time) *time.Time {
	return &t
}

func Sha256(data []byte) []byte {
	m := sha256.New()
	m.Write(data)
	return m.Sum(nil)
}
