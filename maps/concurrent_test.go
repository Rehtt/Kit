package maps

import (
	"fmt"
	"strconv"
	"testing"
)

func TestMap(t *testing.T) {
	m := NewConcurrentMap[string]()
	m.Set("123", "1234")
	m.Set("234", "234")
	fmt.Println(m.Get("123"))
	fmt.Println(m.Get("234"))
	fmt.Println(m.Get("asdf"))
	m.Delete("123")
	fmt.Println(m.Get("123"))
}

func BenchmarkMap(b *testing.B) {
	m := NewConcurrentMap[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set(strconv.Itoa(i), i)
	}
	b.StopTimer()
}
