package size

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Unit 表示数据大小的计量单位，包括十进制、二进制的字节和位单位
type Unit uint8

const (
	Byte Unit = iota
	Kilobyte
	Megabyte
	Gigabyte
	Terabyte
	Petabyte
	Kibibyte
	Mebibyte
	Gibibyte
	Tebibyte
	Pebibyte
	Bit
	Kilobit
	Megabit
	Gigabit
	Terabit
	Petabit
	Kibibit
	Mebibit
	Gibibit
	Tebibit
	Pebibit
)

type unitInfo struct {
	symbol  string
	factor  uint64 // 每单位对应 factor/divisor 字节
	divisor uint64
}

var unitTable = [...]unitInfo{
	{symbol: "B", factor: 1, divisor: 1},
	{symbol: "KB", factor: 1_000, divisor: 1},
	{symbol: "MB", factor: 1_000_000, divisor: 1},
	{symbol: "GB", factor: 1_000_000_000, divisor: 1},
	{symbol: "TB", factor: 1_000_000_000_000, divisor: 1},
	{symbol: "PB", factor: 1_000_000_000_000_000, divisor: 1},
	{symbol: "KiB", factor: 1 << 10, divisor: 1},
	{symbol: "MiB", factor: 1 << 20, divisor: 1},
	{symbol: "GiB", factor: 1 << 30, divisor: 1},
	{symbol: "TiB", factor: 1 << 40, divisor: 1},
	{symbol: "PiB", factor: 1 << 50, divisor: 1},
	{symbol: "b", factor: 1, divisor: 8},
	{symbol: "Kb", factor: 1_000, divisor: 8},
	{symbol: "Mb", factor: 1_000_000, divisor: 8},
	{symbol: "Gb", factor: 1_000_000_000, divisor: 8},
	{symbol: "Tb", factor: 1_000_000_000_000, divisor: 8},
	{symbol: "Pb", factor: 1_000_000_000_000_000, divisor: 8},
	{symbol: "Kib", factor: 1 << 10, divisor: 8},
	{symbol: "Mib", factor: 1 << 20, divisor: 8},
	{symbol: "Gib", factor: 1 << 30, divisor: 8},
	{symbol: "Tib", factor: 1 << 40, divisor: 8},
	{symbol: "Pib", factor: 1 << 50, divisor: 8},
}

// ByteUnit 是旧版 API 使用的单位前缀表
//
// Deprecated: 改用 Unit 常量；解析单位字符串时使用 ParseUnit
var ByteUnit = []string{"", "K", "M", "G", "T", "P"}

// String 返回单位的规范写法，例如 Kibibyte 返回 "KiB"，Kilobit 返回 "Kb"
func (u Unit) String() string {
	if info, ok := u.info(); ok {
		return info.symbol
	}
	return u.invalidString()
}

func (u Unit) invalidString() string {
	return "Unit(" + strconv.FormatUint(uint64(u), 10) + ")"
}

func (u Unit) info() (unitInfo, bool) {
	if int(u) >= len(unitTable) {
		return unitInfo{}, false
	}
	return unitTable[u], true
}

// ParseUnit 解析单位字符串。K、M、G、T、P 以及二进制单位中的 i 均不区分
// 大小写；末尾的大写 B 表示字节，小写 b 表示位
func ParseUnit(input string) (Unit, error) {
	s := strings.TrimSpace(input)
	for i, info := range unitTable {
		canonical := info.symbol
		if len(s) != len(canonical) || len(s) == 0 {
			continue
		}
		last := len(s) - 1
		if s[last] == canonical[last] && strings.EqualFold(s[:last], canonical[:last]) {
			return Unit(i), nil
		}
	}
	return 0, unitError(input)
}

func unitError(input string) error {
	return &unitParseError{input: input}
}

type unitParseError struct {
	input string
}

func (e *unitParseError) Error() string {
	return ErrInvalidUnit.Error() + ": " + strconv.Quote(e.input)
}

func (e *unitParseError) Unwrap() error { return ErrInvalidUnit }

// New 将指定单位的整数值换算为字节数。结果不是整数个字节、超出 uint64
// 范围或 unit 无效时返回相应错误
func New(value uint64, unit Unit) (ByteSize, error) {
	return bytesFromRatio(new(big.Int).SetUint64(value), big.NewInt(1), unit)
}

func mustNew(value uint64, unit Unit) ByteSize {
	result, err := New(value, unit)
	if err != nil {
		panic(err)
	}
	return result
}

func B(value uint64) ByteSize   { return mustNew(value, Byte) }
func KB(value uint64) ByteSize  { return mustNew(value, Kilobyte) }
func MB(value uint64) ByteSize  { return mustNew(value, Megabyte) }
func GB(value uint64) ByteSize  { return mustNew(value, Gigabyte) }
func TB(value uint64) ByteSize  { return mustNew(value, Terabyte) }
func PB(value uint64) ByteSize  { return mustNew(value, Petabyte) }
func KiB(value uint64) ByteSize { return mustNew(value, Kibibyte) }
func MiB(value uint64) ByteSize { return mustNew(value, Mebibyte) }
func GiB(value uint64) ByteSize { return mustNew(value, Gibibyte) }
func TiB(value uint64) ByteSize { return mustNew(value, Tebibyte) }
func PiB(value uint64) ByteSize { return mustNew(value, Pebibyte) }

// ByteSizeUnit 是旧版 API 中“数值 + 单位”的浮点表示
//
// Deprecated: 单位换算改用 ByteSize.In，输出字符串改用 ByteSize.Format
type ByteSizeUnit struct {
	Size float64
	Unit string
}

// String 按旧版格式输出：固定保留两位小数，并在数值和单位之间留一个空格
func (b ByteSizeUnit) String() string {
	return strconv.FormatFloat(b.Size, 'f', 2, 64) + " " + b.Unit
}

// ToByteSize 将 b 换算为字节数，不足一个字节的部分直接舍去
//
// Deprecated: 整数值改用 New；带小数的文本输入改用 Parse
func (b ByteSizeUnit) ToByteSize() ByteSize {
	unit, err := ParseUnit(b.Unit)
	if err != nil || math.IsNaN(b.Size) || math.IsInf(b.Size, 0) || b.Size <= 0 {
		return 0
	}
	info, _ := unit.info()
	value := b.Size * float64(info.factor) / float64(info.divisor)
	if value >= math.Exp2(64) {
		return ByteSize(math.MaxUint64)
	}
	return ByteSize(value)
}

func legacyUnit(s ByteSize, unit Unit) ByteSizeUnit {
	value, _ := s.In(unit)
	return ByteSizeUnit{Size: value, Unit: unit.String()}
}

// Deprecated: 改用 s.In(Kilobyte)
func (s ByteSize) KB() ByteSizeUnit { return legacyUnit(s, Kilobyte) }

// Deprecated: 改用 s.In(Megabyte)
func (s ByteSize) MB() ByteSizeUnit { return legacyUnit(s, Megabyte) }

// Deprecated: 改用 s.In(Gigabyte)
func (s ByteSize) GB() ByteSizeUnit { return legacyUnit(s, Gigabyte) }

// Deprecated: 改用 s.In(Terabyte)
func (s ByteSize) TB() ByteSizeUnit { return legacyUnit(s, Terabyte) }

// Deprecated: 改用 s.In(Petabyte)
func (s ByteSize) PB() ByteSizeUnit { return legacyUnit(s, Petabyte) }

// Deprecated: 改用 s.In(Kibibyte)
func (s ByteSize) KiB() ByteSizeUnit { return legacyUnit(s, Kibibyte) }

// Deprecated: 改用 s.In(Mebibyte)
func (s ByteSize) MiB() ByteSizeUnit { return legacyUnit(s, Mebibyte) }

// Deprecated: 改用 s.In(Gibibyte)
func (s ByteSize) GiB() ByteSizeUnit { return legacyUnit(s, Gibibyte) }

// Deprecated: 改用 s.In(Tebibyte)
func (s ByteSize) TiB() ByteSizeUnit { return legacyUnit(s, Tebibyte) }

// Deprecated: 改用 s.In(Pebibyte)
func (s ByteSize) PiB() ByteSizeUnit { return legacyUnit(s, Pebibyte) }
