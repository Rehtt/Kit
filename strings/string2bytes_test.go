package strings

import (
	"fmt"
	"testing"
)

func TestToBytes(t *testing.T) {
	a := "test test test"
	fmt.Println(UnsafeToBytes(a))
}
