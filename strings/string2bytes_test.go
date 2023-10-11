package strings

import (
	"fmt"
	"testing"
)

func TestToBytes(t *testing.T) {
	a := "test test test"
	fmt.Println(ToBytes(a))
}

func TestToString(t *testing.T) {
	a := []byte{'t', 'e', 's', 't'}
	fmt.Println(ToString(a))
}
