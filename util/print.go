// Copyright (c) 2025 Rehtt
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// @Author: Rehtt dsreshiram@gmail.com
// @Date: 2025/11/25

package util

import (
	"fmt"
	"io"
	"os"
	"time"
	"unicode/utf8"

	"github.com/Rehtt/Kit/random"
	"github.com/Rehtt/Kit/vt"
)

type EffectsPrintOptions struct {
	Out              io.Writer     // 默认 os.Stdout
	EffectsPrintTime time.Duration // 默认 30ms
}

// EffectsPrint 打印字符串，加滚动特效
func EffectsPrint(str string, n int, opts ...EffectsPrintOptions) {
	var out io.Writer
	et := 30 * time.Millisecond
	out = os.Stdout

	if len(opts) > 0 {
		if opts[0].Out != nil {
			out = opts[0].Out
		}
		if opts[0].EffectsPrintTime > 0 {
			et = opts[0].EffectsPrintTime
		}
	}

	vt.HideCursor(out)
	defer vt.ShowCursor(out)

	for _, r := range str {
		if r == ' ' || r == '\n' || r == '\t' {
			fmt.Fprintf(out, "%c", r)
			continue
		}

		isWide := utf8.RuneLen(r) > 1
		for range n {
			if isWide {
				// 输出汉字(宽2) -> 回退2格
				fmt.Fprintf(out, "%c", random.RandHanzi(true))
				time.Sleep(et)
				vt.LeftN(2, out)
			} else {
				// 输出符号(宽1) -> 回退1格
				fmt.Fprintf(out, "%c", random.RandASCII(true))
				time.Sleep(et)
				vt.LeftN(1, out)
			}
		}
		fmt.Fprintf(out, "%c", r)
	}
}
