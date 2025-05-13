package _struct

import (
	"reflect"
	"testing"
)

// Sample 用于测试的结构体
type Sample struct {
	A int    `json:"a" xml:"A"`
	B string `xml:"b"`
	C bool
}

func TestGetTag(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		key     string
		want    map[string]any
		wantErr bool
	}{
		{
			name:    "非结构体类型",
			input:   123,
			key:     "json",
			wantErr: true,
		},
		{
			name:  "结构体值类型，json tag",
			input: Sample{A: 1, B: "x", C: true},
			key:   "json",
			want: map[string]any{
				"A": "a",
				"B": "",
				"C": "",
			},
		},
		{
			name:  "结构体值类型，xml tag",
			input: Sample{A: 2, B: "y", C: false},
			key:   "xml",
			want: map[string]any{
				"A": "A",
				"B": "b",
				"C": "",
			},
		},
		{
			name:  "结构体指针，json tag",
			input: &Sample{A: 3, B: "z", C: false},
			key:   "json",
			want: map[string]any{
				"A": "a",
				"B": "",
				"C": "",
			},
		},
		{
			name:  "nil 结构体指针",
			input: (*Sample)(nil),
			key:   "xml",
			want: map[string]any{
				"A": "A",
				"B": "b",
				"C": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTag(tt.input, tt.key)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTag(%#v, %q) = %#v; want %#v", tt.input, tt.key, got, tt.want)
			}
		})
	}
}
