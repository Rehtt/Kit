package size

import (
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

	var num float64
	index := strings.IndexFunc(str, func(r rune) bool {
		return !(r >= 0x30 && r <= 0x39 || r == '.')
	})
	num, err = strconv.ParseFloat(str[:index], 64)
	if err != nil {
		return 0, fmt.Errorf("解析错误：%s", str[:index])
	}
	unitStr := strings.TrimSpace(str[index:])
	if len(str) < 1 {
		return 0, fmt.Errorf("解析错误:%s", str)
	}
	var unit, uindex int
	for i, u := range ByteUnit {
		if string(unitStr[uindex]) == u || string(unitStr[uindex]) == strings.ToLower(u) {
			unit = i
			uindex++
			break
		}
	}
	if unitStr[uindex] == 'i' {
		num *= math.Pow(1024, float64(unit))
		uindex++
	} else {
		num *= math.Pow(1000, float64(unit))
	}

	if unitStr[uindex] == 'B' {
	} else if unitStr[uindex] == 'b' {
		num /= 8
	} else {
		return 0, fmt.Errorf("解析错误：%s", str)
	}
	size = ByteSize(num)
	return
}
