package _struct

import (
	"reflect"
	"strings"
	"testing"
)

type MyStruct struct {
	IntField    int
	UintField   uint
	FloatField  float64
	BoolField   bool
	StringField string
	skipField   int
}

func TestParseArrayStringStruct(t *testing.T) {
	tests := []struct {
		name    string
		rows    [][]string
		data    any
		p       ParseFuncs
		want    any
		wantErr bool
		errMsg  string
	}{
		{
			name: "slice of structs default parsing",
			rows: [][]string{{"1", "2", "3.14", "true", "hello", "ignored"}},
			data: &[]MyStruct{},
			p:    nil,
			want: []MyStruct{{1, 2, 3.14, true, "hello", 0}},
		},
		{
			name: "slice of ptr to structs default parsing",
			rows: [][]string{{"4", "5", "6.28", "false", "world"}},
			data: &[]*MyStruct{},
			p:    nil,
			want: []*MyStruct{{IntField: 4, UintField: 5, FloatField: 6.28, BoolField: false, StringField: "world"}},
		},
		{
			name: "fewer columns",
			rows: [][]string{{"7", ""}},
			data: &[]MyStruct{},
			p:    nil,
			want: []MyStruct{{IntField: 7, UintField: 0, FloatField: 0, BoolField: false, StringField: ""}},
		},
		{
			name: "custom parser override",
			rows: [][]string{{"0", "0", "0", "false", "abc"}},
			data: &[]MyStruct{},
			p: ParseFuncs{
				"StringField": func(txt string) any { return strings.ToUpper(txt) },
			},
			want: []MyStruct{{0, 0, 0, false, "ABC", 0}},
		},
		{
			name:    "custom parser wrong return type",
			rows:    [][]string{{"9"}},
			data:    &[]MyStruct{},
			p:       ParseFuncs{"IntField": func(txt string) any { return "notint" }},
			wantErr: true,
			errMsg:  "func return type must be int",
		},
		{
			name:    "non-pointer data",
			rows:    [][]string{{}},
			data:    []MyStruct{},
			p:       nil,
			wantErr: true,
			errMsg:  "data must be &[]*struct/&[]struct",
		},
		{
			name:    "pointer to non-slice",
			rows:    [][]string{{}},
			data:    &MyStruct{},
			p:       nil,
			wantErr: true,
			errMsg:  "data must be &[]*struct/&[]struct",
		},
		{
			name:    "pointer to slice of non-struct",
			rows:    [][]string{{}},
			data:    &[]int{},
			p:       nil,
			wantErr: true,
			errMsg:  "data must be &[]*struct/&[]struct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseArrayStringStruct(tt.rows, tt.data, tt.p)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if err.Error() != tt.errMsg {
					t.Errorf("unexpected error: got %q want %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := reflect.ValueOf(tt.data).Elem().Interface()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
