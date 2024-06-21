package util

import (
	"fmt"
	"testing"
	"time"
)

func TestSnowflake(t *testing.T) {
	s := Snowflake{
		BaseTime:  time.Date(2024, 6, 21, 0, 0, 0, 0, time.Local),
		LogicalId: 1234,
	}
	fmt.Println(s.GenerateId())
	time.Sleep(500 * time.Millisecond)
	fmt.Println(s.GenerateId())
	time.Sleep(time.Second)
	fmt.Println(s.GenerateId())
}
