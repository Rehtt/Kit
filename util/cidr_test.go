package util

import (
	"fmt"
	"testing"
)

func TestCalculateCIDR(t *testing.T) {
	cidr, err := CalculateCIDR("1.0.0.0", "1.3.0.3")
	if err != nil {
		t.Errorf("err %s", err)
	}
	fmt.Println(cidr)
}
