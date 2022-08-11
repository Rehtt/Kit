package buf

import (
	"fmt"
	"github.com/fatih/color"
	"testing"
)

func TestNewBuf(t *testing.T) {
	buf := NewBuf().WriteString("test\n").WriteColor("test Coloe", color.FgHiRed)
	fmt.Println(buf.ToString())
}
