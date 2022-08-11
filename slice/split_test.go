package slice

import (
	"fmt"
	"testing"
)

func TestSplit(t *testing.T) {
	a := []int{1, 2, 3}
	for _, v := range Split(a, 2) {
		fmt.Println(v.([]int))
	}
	for _, v := range Split(&a, 2) {
		fmt.Println(v.(*[]int))
	}

	b := []string{"a", "b", "c"}
	for _, v := range Split(b, 2) {
		fmt.Println(v.([]string))
	}

	c := []float64{1.234, 3.2345, 5.123}
	for _, v := range Split(c, 2) {
		fmt.Println(v.([]float64))
	}
}
