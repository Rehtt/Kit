package util

import (
	"fmt"
	"time"
)

const (
	m = 60 * time.Second
	h = 60 * m
	d = 24 * h
)

func Duration2String(t time.Duration, secondPrecision ...string) string {
	if t < m {
		if len(secondPrecision) == 0 {
			secondPrecision = []string{"%.0f"}
		}
		return fmt.Sprintf(secondPrecision[0]+"s", t.Seconds())
	}
	if t < h {
		tmp := t / m
		return fmt.Sprintf("%dm %s", tmp, Duration2String(t-(tmp*m)))
	}
	if t < d {
		tmp := t / h
		return fmt.Sprintf("%dh %s", tmp, Duration2String(t-(tmp*h)))
	}
	tmp := t / d
	return fmt.Sprintf("%dd %s", tmp, Duration2String(t-(tmp*d)))
}
