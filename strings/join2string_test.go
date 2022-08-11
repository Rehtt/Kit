package strings

import (
	"fmt"
	"testing"
)

func TestJoinToString(t *testing.T) {
	a := []string{"a", "b", "c"}
	fmt.Println(JoinToString(a, ","))
	b := []int{1, 2, 3}
	fmt.Println(JoinToString(b, ","))
	c := []float64{1.234, 2.345, 3.456}
	fmt.Println(JoinToString(c, ","))
}
