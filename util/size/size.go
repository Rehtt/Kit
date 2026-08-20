// Package size 用于解析和格式化数据大小，并在常用的字节、位单位之间换算
package size

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
)

// ByteSize 表示以字节为单位的数据大小
type ByteSize uint64

var (
	// ErrInvalidFormat 表示输入不符合数据大小的书写格式
	ErrInvalidFormat = errors.New("invalid size format")
	// ErrInvalidUnit 表示单位无法识别
	ErrInvalidUnit = errors.New("invalid size unit")
	// ErrFractionalByte 表示换算结果包含不足一个字节的小数部分
	ErrFractionalByte = errors.New("size is not a whole number of bytes")
	// ErrOverflow 表示换算后的字节数超出 uint64 的取值范围
	ErrOverflow = errors.New("size overflows uint64")
	// ErrInvalidPrecision 表示小数位数小于 0
	ErrInvalidPrecision = errors.New("invalid size precision")
)

var maxUint64Int = new(big.Int).SetUint64(math.MaxUint64)

// Parse 将形如 "1.5 MiB"、"8Kb" 的字符串换算为字节数
// 数值和单位之间可以留空格；不写单位时按字节处理。单位前缀及其中的 i
// 不区分大小写，但末尾的 B 和 b 分别表示字节和位，大小写含义不同
func Parse(input string) (ByteSize, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return 0, formatError(input)
	}

	numberEnd, fractionDigits, ok := scanDecimal(s)
	if !ok {
		return 0, formatError(input)
	}

	unitText := strings.TrimSpace(s[numberEnd:])
	unit := Byte
	if unitText != "" {
		// 单位只能以字母开头；这样可将重复小数点等问题归为数值格式错误
		first := unitText[0]
		if (first < 'A' || first > 'Z') && (first < 'a' || first > 'z') {
			return 0, formatError(input)
		}
		var err error
		unit, err = ParseUnit(unitText)
		if err != nil {
			return 0, fmt.Errorf("%w: %q", err, input)
		}
	}

	number := decimalInteger(s[:numberEnd])
	denominator := pow10(fractionDigits)
	return bytesFromRatio(number, denominator, unit)
}

// ParseFromString 将字符串表示的数据大小换算为字节数
//
// Deprecated: 改用 Parse
func ParseFromString(str string) (ByteSize, error) {
	return Parse(str)
}

// In 将 s 换算成指定单位。返回值为 float64，数值较大时可能损失精度
func (s ByteSize) In(unit Unit) (float64, error) {
	info, ok := unit.info()
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrInvalidUnit, unit.invalidString())
	}
	return float64(s) * float64(info.divisor) / float64(info.factor), nil
}

// Format 按指定单位和小数位数格式化 s。结果采用四舍五入，数值和单位之间
// 不留空格，例如 ByteSize(1536).Format(Kibibyte, 2) 返回 "1.50KiB"
func (s ByteSize) Format(unit Unit, precision int) (string, error) {
	if precision < 0 {
		return "", fmt.Errorf("%w: %d", ErrInvalidPrecision, precision)
	}
	info, ok := unit.info()
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrInvalidUnit, unit.invalidString())
	}

	numerator := new(big.Int).SetUint64(uint64(s))
	numerator.Mul(numerator, new(big.Int).SetUint64(info.divisor))
	if precision != 0 {
		numerator.Mul(numerator, pow10(precision))
	}
	denominator := new(big.Int).SetUint64(info.factor)

	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, denominator, remainder)
	if remainder.Lsh(remainder, 1).Cmp(denominator) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}

	return fixedDecimal(quotient.String(), precision) + info.symbol, nil
}

// String 自动选用不大于 s 的最大二进制字节单位，最多保留两位小数，
// 并省略小数末尾无意义的 0。例如 1536 字节会格式化为 "1.5KiB"
func (s ByteSize) String() string {
	unit := Byte
	for _, candidate := range [...]Unit{Pebibyte, Tebibyte, Gibibyte, Mebibyte, Kibibyte} {
		info, _ := candidate.info()
		if uint64(s) >= info.factor {
			unit = candidate
			break
		}
	}

	formatted, _ := s.Format(unit, 2)
	suffix := unit.String()
	number := strings.TrimSuffix(formatted, suffix)
	number = strings.TrimRight(strings.TrimRight(number, "0"), ".")
	return number + suffix
}

func scanDecimal(s string) (end, fractionDigits int, ok bool) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	digits := i
	if i < len(s) && s[i] == '.' {
		i++
		fractionStart := i
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		fractionDigits = i - fractionStart
		digits += fractionDigits
	}
	return i, fractionDigits, digits != 0
}

func decimalInteger(number string) *big.Int {
	digits := strings.Replace(number, ".", "", 1)
	value, _ := new(big.Int).SetString(digits, 10)
	return value
}

func pow10(exponent int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exponent)), nil)
}

func bytesFromRatio(numerator, denominator *big.Int, unit Unit) (ByteSize, error) {
	info, ok := unit.info()
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrInvalidUnit, unit.invalidString())
	}

	n := new(big.Int).Mul(new(big.Int).Set(numerator), new(big.Int).SetUint64(info.factor))
	d := new(big.Int).Mul(new(big.Int).Set(denominator), new(big.Int).SetUint64(info.divisor))
	bytes, remainder := new(big.Int), new(big.Int)
	bytes.QuoRem(n, d, remainder)
	if remainder.Sign() != 0 {
		return 0, ErrFractionalByte
	}
	if bytes.Cmp(maxUint64Int) > 0 {
		return 0, ErrOverflow
	}
	return ByteSize(bytes.Uint64()), nil
}

func fixedDecimal(digits string, precision int) string {
	if precision == 0 {
		return digits
	}
	if len(digits) <= precision {
		digits = strings.Repeat("0", precision-len(digits)+1) + digits
	}
	point := len(digits) - precision
	return digits[:point] + "." + digits[point:]
}

func formatError(input string) error {
	return fmt.Errorf("%w: %q", ErrInvalidFormat, input)
}
