package vt

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Rehtt/Kit/buf"
)

// 清屏
func Clear(w ...io.Writer) {
	b := buf.NewBuf().WriteString("\033c")
	if len(w) == 0 {
		Print(b)
		return
	}
	b.WriteTo(w[0], true)
}

// 清除从光标到行尾的内容
func ClearLine(w ...io.Writer) {
	b := buf.NewBuf().WriteString("\u001B[K")
	if len(w) == 0 {
		Print(b)
		return
	}
	b.WriteTo(w[0], true)
}

// 保存光标位置
func Save(w ...io.Writer) {
	b := buf.NewBuf().WriteString("\u001B[s")
	if len(w) == 0 {
		Print(b)
		return
	}
	b.WriteTo(w[0], true)
}

// 恢复光标位置
func Recover(w ...io.Writer) {
	b := buf.NewBuf().WriteString("\u001B[u")
	if len(w) == 0 {
		Print(b)
		return
	}
	b.WriteTo(w[0], true)
}

// 设置光标位置
func ToXY(x, y int, w ...io.Writer) {
	b := buf.NewBuf().WriteString("\033[").
		WriteString(strconv.Itoa(x)).
		WriteByte(';').
		WriteString(strconv.Itoa(y)).
		WriteByte('H')
	if len(w) == 0 {
		Print(b)
		return
	}
	b.WriteTo(w[0], true)
}

// 光标上移 n 行
func TopN(n int, w ...io.Writer) {
	b := buf.NewBuf().WriteString("\033[").
		WriteString(strconv.Itoa(n)).
		WriteByte('A')
	if len(w) == 0 {
		Print(b)
		return
	}
	b.WriteTo(w[0], true)
}

// 光标下移 n 行
func DownN(n int, w ...io.Writer) {
	b := buf.NewBuf().WriteString("\033[").
		WriteString(strconv.Itoa(n)).
		WriteByte('B')
	if len(w) == 0 {
		Print(b)
		return
	}
	b.WriteTo(w[0], true)
}

// 光标右移 n 行
func RightN(n int, w ...io.Writer) {
	b := buf.NewBuf().WriteString("\033[").
		WriteString(strconv.Itoa(n)).
		WriteByte('C')
	if len(w) == 0 {
		Print(b)
		return
	}
	b.WriteTo(w[0], true)
}

// 光标左移 n 行
func LeftN(n int, w ...io.Writer) {
	b := buf.NewBuf().WriteString("\033[").
		WriteString(strconv.Itoa(n)).
		WriteByte('D')
	if len(w) == 0 {
		Print(b)
		return
	}
	b.WriteTo(w[0], true)
}

// 隐藏光标
func HideCursor(w ...io.Writer) {
	b := buf.NewBuf().WriteString("\033[?25l")
	if len(w) == 0 {
		Print(b)
		return
	}
	b.WriteTo(w[0], true)
}

// 显示光标
func ShowCursor(w ...io.Writer) {
	b := buf.NewBuf().WriteString("\033[?25h")
	if len(w) == 0 {
		Print(b)
		return
	}
	b.WriteTo(w[0], true)
}

func Print(buf *buf.Buf) {
	fmt.Println(buf.ToString(true))
}
