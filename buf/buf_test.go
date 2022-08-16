package buf

import (
	"fmt"
	"github.com/Rehtt/Kit/vt/color"
	"testing"
)

func TestNewBuf(t *testing.T) {
	buf := NewBuf().WriteString("test\n").WriteColor("test Coloe", color.FgGreen, color.BgHiRed)
	fmt.Println(buf.ToString())
}
