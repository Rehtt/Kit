package size

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type ByteSize uint64
type ByteSizeUnit struct {
	Size float64
	Unit string
}

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
func (s ByteSize) PiB() ByteSizeUnit {
	return ByteSizeUnit{
		Size: float64(s) / math.Pow(1024, 5),
		Unit: "PiB",
	}
}
func (s ByteSize) TiB() ByteSizeUnit {
	return ByteSizeUnit{
		Size: float64(s) / math.Pow(1024, 4),
		Unit: "TiB",
	}
}
func (s ByteSize) GiB() ByteSizeUnit {
	return ByteSizeUnit{
		Size: float64(s) / math.Pow(1024, 3),
		Unit: "GiB",
	}
}
func (s ByteSize) MiB() ByteSizeUnit {
	return ByteSizeUnit{
		Size: float64(s) / math.Pow(1024, 2),
		Unit: "MiB",
	}
}
func (s ByteSize) KiB() ByteSizeUnit {
	return ByteSizeUnit{
		Size: float64(s) / math.Pow(1024, 1),
		Unit: "KiB",
	}
}

func (s ByteSize) PB() ByteSizeUnit {
	return ByteSizeUnit{
		Size: float64(s) / math.Pow(1000, 5),
		Unit: "PB",
	}
}
func (s ByteSize) TB() ByteSizeUnit {
	return ByteSizeUnit{
		Size: float64(s) / math.Pow(1000, 4),
		Unit: "TB",
	}
}
func (s ByteSize) GB() ByteSizeUnit {
	return ByteSizeUnit{
		Size: float64(s) / math.Pow(1000, 3),
		Unit: "GB",
	}
}
func (s ByteSize) MB() ByteSizeUnit {
	return ByteSizeUnit{
		Size: float64(s) / math.Pow(1000, 2),
		Unit: "MB",
	}
}
func (s ByteSize) KB() ByteSizeUnit {
	return ByteSizeUnit{
		Size: float64(s) / math.Pow(1000, 1),
		Unit: "KB",
	}
}

func (b ByteSizeUnit) String() string {
	return strconv.FormatFloat(b.Size, 'f', 2, 64) + " " + b.Unit
}
