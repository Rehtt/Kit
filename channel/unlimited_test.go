package channel

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	c := New()

	go func(c *channel) {
		for i := 0; i < 20; i++ {
			c.In <- i + 1
		}
		close(c.In)
	}(c)
	for v := range c.Out {
		fmt.Println(v)

	}
	fmt.Println(111111111, c.Cap())
}
