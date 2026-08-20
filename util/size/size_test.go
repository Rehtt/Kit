package size

import (
	"errors"
	"math"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  ByteSize
	}{
		{name: "zero", input: "0", want: 0},
		{name: "bytes without unit", input: "42", want: 42},
		{name: "integral decimal without unit", input: "1.0", want: 1},
		{name: "surrounding whitespace", input: " \t1 B\n", want: 1},
		{name: "decimal byte unit", input: "1.5 MB", want: 1_500_000},
		{name: "binary byte unit", input: "1.5MiB", want: 1_572_864},
		{name: "case insensitive decimal prefix", input: "2kB", want: 2_000},
		{name: "case insensitive binary marker", input: "2kIB", want: 2_048},
		{name: "leading decimal point", input: ".008 Kb", want: 1},
		{name: "exact decimal fraction", input: "0.001KB", want: 1},
		{name: "decimal bits", input: "8Kb", want: 1_000},
		{name: "binary bits", input: "8Kib", want: 1_024},
		{name: "single bits", input: "8b", want: 1},
		{name: "maximum", input: "18446744073709551615B", want: ByteSize(math.MaxUint64)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("Parse(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  error
	}{
		{name: "empty", input: "", want: ErrInvalidFormat},
		{name: "whitespace", input: " \t", want: ErrInvalidFormat},
		{name: "negative", input: "-1B", want: ErrInvalidFormat},
		{name: "decimal point", input: ".", want: ErrInvalidFormat},
		{name: "double decimal point", input: "1..0B", want: ErrInvalidFormat},
		{name: "incomplete prefix", input: "1K", want: ErrInvalidUnit},
		{name: "incomplete binary unit", input: "1Ki", want: ErrInvalidUnit},
		{name: "trailing characters", input: "1KB/s", want: ErrInvalidUnit},
		{name: "unknown unit", input: "1XB", want: ErrInvalidUnit},
		{name: "lowercase means bits", input: "1b", want: ErrFractionalByte},
		{name: "fractional bytes", input: "1.5", want: ErrFractionalByte},
		{name: "fractional unit result", input: "0.1B", want: ErrFractionalByte},
		{name: "overflow", input: "18446744073709551616B", want: ErrOverflow},
		{name: "scaled overflow", input: "18446744073709552KB", want: ErrOverflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Parse(%q) = (%d, %v), want error %v", tt.input, got, err, tt.want)
			}
		})
	}
}

func TestParseFromStringCompatibility(t *testing.T) {
	got, err := ParseFromString("1.5 MiB")
	if err != nil || got != 1_572_864 {
		t.Fatalf("ParseFromString() = (%d, %v), want (1572864, nil)", got, err)
	}
	if _, err := ParseFromString(""); !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("ParseFromString(\"\") error = %v, want ErrInvalidFormat", err)
	}
}

func TestByteSizeString(t *testing.T) {
	tests := []struct {
		size ByteSize
		want string
	}{
		{size: 0, want: "0B"},
		{size: 1_023, want: "1023B"},
		{size: 1_024, want: "1KiB"},
		{size: 1_280, want: "1.25KiB"},
		{size: 1_536, want: "1.5KiB"},
		{size: 2_043, want: "2KiB"},
		{size: 1_024*1_024 - 1, want: "1024KiB"},
		{size: 1_024 * 1_024, want: "1MiB"},
		{size: 1_024 * 1_024 * 1_024, want: "1GiB"},
		{size: 1_024 * 1_024 * 1_024 * 1_024, want: "1TiB"},
		{size: 1_024 * 1_024 * 1_024 * 1_024 * 1_024, want: "1PiB"},
	}

	for _, tt := range tests {
		if got := tt.size.String(); got != tt.want {
			t.Errorf("ByteSize(%d).String() = %q, want %q", tt.size, got, tt.want)
		}
	}
}

func TestByteSizeInAndFormat(t *testing.T) {
	got, err := ByteSize(1_536).In(Kibibyte)
	if err != nil || got != 1.5 {
		t.Fatalf("In(Kibibyte) = (%v, %v), want (1.5, nil)", got, err)
	}
	got, err = ByteSize(1).In(Bit)
	if err != nil || got != 8 {
		t.Fatalf("In(Bit) = (%v, %v), want (8, nil)", got, err)
	}

	formats := []struct {
		size      ByteSize
		unit      Unit
		precision int
		want      string
	}{
		{size: 1_536, unit: Kibibyte, precision: 2, want: "1.50KiB"},
		{size: 1_536, unit: Kibibyte, precision: 0, want: "2KiB"},
		{size: 1_280, unit: Kibibyte, precision: 1, want: "1.3KiB"},
		{size: 1, unit: Kilobit, precision: 3, want: "0.008Kb"},
		{size: math.MaxUint64, unit: Byte, precision: 0, want: "18446744073709551615B"},
	}
	for _, tt := range formats {
		got, err := tt.size.Format(tt.unit, tt.precision)
		if err != nil || got != tt.want {
			t.Errorf("ByteSize(%d).Format(%s, %d) = (%q, %v), want (%q, nil)", tt.size, tt.unit, tt.precision, got, err, tt.want)
		}
	}

	invalid := Unit(255)
	if _, err := ByteSize(1).In(invalid); !errors.Is(err, ErrInvalidUnit) {
		t.Errorf("In(invalid) error = %v, want ErrInvalidUnit", err)
	}
	if _, err := ByteSize(1).Format(invalid, 2); !errors.Is(err, ErrInvalidUnit) {
		t.Errorf("Format(invalid) error = %v, want ErrInvalidUnit", err)
	}
	if _, err := ByteSize(1).Format(Byte, -1); !errors.Is(err, ErrInvalidPrecision) {
		t.Errorf("Format(Byte, -1) error = %v, want ErrInvalidPrecision", err)
	}
}

func FuzzParseNeverPanics(f *testing.F) {
	for _, seed := range []string{"", "1", "1.5MiB", "1b", "1Ki", "1KBjunk", "18446744073709551616B", "\xff"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		_, _ = Parse(input)
	})
}
