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
	for i := 0; i < b.N; i++ {
		m.Set(strconv.Itoa(i), i)
	}
}

func BenchmarkConcurrentMap(b *testing.B) {
	m := NewConcurrentMap[int]()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Set(strconv.Itoa(b.N), b.N)
		}
	})
}
