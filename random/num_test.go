package random

import (
	"testing"
)

func TestRandInt64(t *testing.T) {
	data := []struct {
		min int64
		max int64
	}{
		{min: 123, max: 235},
		{min: 0, max: 1},
		{min: -90, max: -80},
	}
	for i := 0; i < 100; i++ {
		for _, v := range data {
			out := RandInt64(v.min, v.max)
			if out < v.min || out > v.max {
				t.Fatalf("error min: %d, max: %d, out: %d\n", v.min, v.max, out)
			}
			t.Logf("min: %d, max: %d, out: %d\n", v.min, v.max, out)
		}
	}
}
