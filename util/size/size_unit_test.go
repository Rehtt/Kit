package size

import (
	"errors"
	"math"
	"testing"
)

func TestUnits(t *testing.T) {
	tests := []struct {
		unit   Unit
		symbol string
		value  uint64
		bytes  ByteSize
	}{
		{Byte, "B", 1, 1},
		{Kilobyte, "KB", 1, 1_000},
		{Megabyte, "MB", 1, 1_000_000},
		{Gigabyte, "GB", 1, 1_000_000_000},
		{Terabyte, "TB", 1, 1_000_000_000_000},
		{Petabyte, "PB", 1, 1_000_000_000_000_000},
		{Kibibyte, "KiB", 1, 1 << 10},
		{Mebibyte, "MiB", 1, 1 << 20},
		{Gibibyte, "GiB", 1, 1 << 30},
		{Tebibyte, "TiB", 1, 1 << 40},
		{Pebibyte, "PiB", 1, 1 << 50},
		{Bit, "b", 8, 1},
		{Kilobit, "Kb", 1, 125},
		{Megabit, "Mb", 1, 125_000},
		{Gigabit, "Gb", 1, 125_000_000},
		{Terabit, "Tb", 1, 125_000_000_000},
		{Petabit, "Pb", 1, 125_000_000_000_000},
		{Kibibit, "Kib", 1, 1 << 7},
		{Mebibit, "Mib", 1, 1 << 17},
		{Gibibit, "Gib", 1, 1 << 27},
		{Tebibit, "Tib", 1, 1 << 37},
		{Pebibit, "Pib", 1, 1 << 47},
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			if got := tt.unit.String(); got != tt.symbol {
				t.Errorf("Unit.String() = %q, want %q", got, tt.symbol)
			}
			parsed, err := ParseUnit(tt.symbol)
			if err != nil || parsed != tt.unit {
				t.Errorf("ParseUnit(%q) = (%v, %v), want (%v, nil)", tt.symbol, parsed, err, tt.unit)
			}
			got, err := New(tt.value, tt.unit)
			if err != nil || got != tt.bytes {
				t.Errorf("New(%d, %s) = (%d, %v), want (%d, nil)", tt.value, tt.unit, got, err, tt.bytes)
			}
			in, err := tt.bytes.In(tt.unit)
			if err != nil || in != float64(tt.value) {
				t.Errorf("ByteSize(%d).In(%s) = (%v, %v), want (%v, nil)", tt.bytes, tt.unit, in, err, float64(tt.value))
			}
		})
	}
}

func TestParseUnitAliasesAndErrors(t *testing.T) {
	aliases := map[string]Unit{
		" kB ": Kilobyte,
		"KIB":  Kibibyte,
		"mIb":  Mebibit,
		"pB":   Petabyte,
		"pb":   Petabit,
	}
	for input, want := range aliases {
		got, err := ParseUnit(input)
		if err != nil || got != want {
			t.Errorf("ParseUnit(%q) = (%v, %v), want (%v, nil)", input, got, err, want)
		}
	}
	for _, input := range []string{"", "K", "Ki", "BB", "XB", "KB/s", "KB " + "junk"} {
		if _, err := ParseUnit(input); !errors.Is(err, ErrInvalidUnit) {
			t.Errorf("ParseUnit(%q) error = %v, want ErrInvalidUnit", input, err)
		}
	}
	if got := Unit(255).String(); got != "Unit(255)" {
		t.Errorf("Unit(255).String() = %q, want Unit(255)", got)
	}
}

func TestNewErrors(t *testing.T) {
	if _, err := New(1, Bit); !errors.Is(err, ErrFractionalByte) {
		t.Errorf("New(1, Bit) error = %v, want ErrFractionalByte", err)
	}
	if _, err := New(math.MaxUint64, Kilobyte); !errors.Is(err, ErrOverflow) {
		t.Errorf("New(MaxUint64, Kilobyte) error = %v, want ErrOverflow", err)
	}
	if _, err := New(1, Unit(255)); !errors.Is(err, ErrInvalidUnit) {
		t.Errorf("New(1, invalid) error = %v, want ErrInvalidUnit", err)
	}
}

func TestShortcutConstructors(t *testing.T) {
	tests := []struct {
		name  string
		build func(uint64) ByteSize
		want  ByteSize
	}{
		{"B", B, 2},
		{"KB", KB, 2_000},
		{"MB", MB, 2_000_000},
		{"GB", GB, 2_000_000_000},
		{"TB", TB, 2_000_000_000_000},
		{"PB", PB, 2_000_000_000_000_000},
		{"KiB", KiB, 2 << 10},
		{"MiB", MiB, 2 << 20},
		{"GiB", GiB, 2 << 30},
		{"TiB", TiB, 2 << 40},
		{"PiB", PiB, 2 << 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.build(2); got != tt.want {
				t.Errorf("%s(2) = %d, want %d", tt.name, got, tt.want)
			}
			if got := tt.build(0); got != 0 {
				t.Errorf("%s(0) = %d, want 0", tt.name, got)
			}
		})
	}
}

func TestShortcutOverflowPanics(t *testing.T) {
	for _, build := range []func(uint64) ByteSize{KB, MB, GB, TB, PB, KiB, MiB, GiB, TiB, PiB} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("overflowing shortcut did not panic")
				}
			}()
			_ = build(math.MaxUint64)
		}()
	}
}

func TestLegacyByteSizeUnit(t *testing.T) {
	if got := ByteSize(1_536).KiB(); got != (ByteSizeUnit{Size: 1.5, Unit: "KiB"}) {
		t.Errorf("legacy KiB conversion = %#v", got)
	}
	if got := (ByteSizeUnit{Size: 1.5, Unit: "KiB"}).String(); got != "1.50 KiB" {
		t.Errorf("ByteSizeUnit.String() = %q, want %q", got, "1.50 KiB")
	}
	if got := (ByteSizeUnit{Size: 1.234, Unit: "KiB"}).ToByteSize(); got != 1_263 {
		t.Errorf("precise legacy ToByteSize() = %d, want 1263", got)
	}
	if got := (ByteSizeUnit{Size: 8, Unit: "Kb"}).ToByteSize(); got != 1_000 {
		t.Errorf("bit legacy ToByteSize() = %d, want 1000", got)
	}
	if got := (ByteSizeUnit{Size: 1, Unit: "invalid"}).ToByteSize(); got != 0 {
		t.Errorf("invalid legacy ToByteSize() = %d, want 0", got)
	}
}
