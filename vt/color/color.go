package color

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// todo 完善

type Color int

// 字体颜色
const (
	FgBlack Color = iota + 30
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
)

// 背景色
const (
	BgBlack Color = iota + 40
	BgRed
	BgGreen
	BgYellow
	BgBlue
	BgMagenta
	BgCyan
	BgWhite
)

// 字体颜色高亮
const (
	FgHiBlack Color = iota + 90
	FgHiRed
	FgHiGreen
	FgHiYellow
	FgHiBlue
	FgHiMagenta
	FgHiCyan
	FgHiWhite
)

// 背景色高亮
const (
	BgHiBlack Color = iota + 100
	BgHiRed
	BgHiGreen
	BgHiYellow
	BgHiBlue
	BgHiMagenta
	BgHiCyan
	BgHiWhite
)

type Colors struct {
	colors []Color
}

func NewColors(colors ...Color) *Colors {
	c := new(Colors)
	c.colors = colors
	return c
}

func (c *Colors) hasColors() bool {
	if c.colors == nil || len(c.colors) == 0 {
		return false
	}
	return true
}
func (c *Colors) startColor() string {
	if !c.hasColors() {
		return ""
	}
	tmp := make([]string, 0, len(c.colors))
	for _, v := range c.colors {
		tmp = append(tmp, strconv.Itoa(int(v)))
	}
	return fmt.Sprintf("\033[%sm", strings.Join(tmp, ";"))
}
func (c *Colors) endColor() string {
	if !c.hasColors() {
		return ""
	}
	return "\033[0m"
}

func (c *Colors) startColorWriter(w io.Writer) {
	io.WriteString(w, c.startColor())
}
func (c *Colors) endColorWriter(w io.Writer) {
	io.WriteString(w, c.endColor())
}

// \033[字背景颜色;字体颜色m 字符串\033[0m
func (c *Colors) Fprint(w io.Writer, a ...interface{}) (n int, err error) {
	c.startColorWriter(w)
	defer c.endColorWriter(w)
	return fmt.Fprint(w, a...)
}

func (c *Colors) Sprint(a ...interface{}) string {
	return c.startColor() + fmt.Sprint(a...) + c.endColor()
}
