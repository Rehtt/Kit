package util

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func TestMap(t *testing.T) {
	m := NewMap()
	m.Set("123", 1234)
	m.Set("234", 4556)
	fmt.Println(m.Get("123"))
	fmt.Println(m.Get("234"))
	fmt.Println(m.Get("asdf"))
	m.Delete("123")
	fmt.Println(m.Get("123"))
}

func TestMapB(t *testing.T) {
	start := time.Now()
	m := NewMap()
	for i := 0; i < 9999999; i++ {
		m.Set(strconv.Itoa(i), i)
		if i%100000 == 0 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			fmt.Printf("%dMB %dMB\n", m.Alloc/1024/1024, m.TotalAlloc/1024/1024)
		}
	}
	fmt.Println("timeUnit", time.Since(start).Milliseconds(), "ms")
}
