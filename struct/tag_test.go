package _struct

import (
	"fmt"
	"testing"
)

func TestGetTag(t *testing.T) {
	type test struct {
		A int    `S:"123"`
		B string `SS:"345"`
	}
	t1 := test{}
	out, err := GetTag(t1, "S")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(out)
	out2, err := GetTag(&t1, "S")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(out2)
}
