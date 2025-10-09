package bytes

import (
	"fmt"
	"testing"
)

func TestToString(t *testing.T) {
	a := []byte("test test test")
	fmt.Println(UnsafeToString(a))
}
