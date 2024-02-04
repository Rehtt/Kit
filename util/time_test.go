package util

import (
	"fmt"
	"testing"
	"time"
)

func TestDuration2String(t *testing.T) {
	a := Duration2String(340221 * time.Second)
	fmt.Println(a)
}
