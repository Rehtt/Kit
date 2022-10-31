package size

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type ByteSize uint64

var (
	ByteUnit = []string{"", "K", "M", "G", "T", "P"}
)

func ParseFromString(str string) (size ByteSize, err error) {
	var tmp bytes.Buffer
	var num = math.NaN()
	for i := range str {
		s := str[i]
		if s == ' ' {
			continue
		}
		if s >= 0x30 && s <= 0x39 || s == '.' {
			tmp.WriteByte(s)
			continue
		}
		if math.IsNaN(num) {
			num, err = strconv.ParseFloat(tmp.String(), 64)
			if err != nil {
				return 0, err
			}
			tmp.Reset()
		}
		var isUnit bool
		for ui, u := range ByteUnit {
			if string(s) == u || string(s) == strings.ToLower(u) {
				tmp.WriteString(strconv.Itoa(ui))
				isUnit = true
				break
			}
		}
		if isUnit {
			continue
		}
		unit, e := strconv.Atoi(tmp.String())
		if e != nil && tmp.Len() != 0 {
			err = fmt.Errorf("解析错误：%s", str)
			break
		}
		if s == 'i' {
			num = -(num * math.Pow(1024, float64(unit)))
			continue
		} else if s == 'B' {
			if num < 0 {
				size = ByteSize(-num)
			} else {
				size = ByteSize(num * math.Pow(1000, float64(unit)))
			}
			return
		} else if s == 'b' {
			if num < 0 {
				size = ByteSize(-num / 8)
			} else {
				size = ByteSize(num * math.Pow(1000, float64(unit)) / 8)
			}
			return
		}

	}
	err = fmt.Errorf("解析错误：%s", str)
	return 0, err
}
