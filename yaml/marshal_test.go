// yaml_marshal_with_comment_table_test.go
package yaml

import (
	"strings"
	"testing"
)

type SimpleStruct struct {
	Name string `yaml:"name" comment:"姓名"`
	Age  int    `yaml:"age"`
}

type Inner struct {
	Field string `yaml:"field" comment:"内部字段"`
}

type Outer struct {
	InnerElem Inner `yaml:"inner"`
}

type Person struct {
	FirstName string `yaml:"first_name" comment:"名"`
	LastName  string `yaml:"last_name" comment:"姓"`
}

func TestMarshalWithComment(t *testing.T) {
	tests := []struct {
		name       string
		input      any
		wantErr    bool
		wantSubstr []string
	}{
		{
			name:    "NonStruct_Int",
			input:   42,
			wantErr: false,
			wantSubstr: []string{
				"42\n",
			},
		},
		{
			name:    "SimpleStruct",
			input:   SimpleStruct{Name: "张三", Age: 30},
			wantErr: false,
			wantSubstr: []string{
				"name: 张三 #姓名",
				"age: 30",
			},
		},
		{
			name:    "NestedStruct",
			input:   Outer{InnerElem: Inner{Field: "value"}},
			wantErr: false,
			wantSubstr: []string{
				"inner:",
				"  field: value #内部字段",
			},
		},
		{
			name:    "PointerStruct",
			input:   &Person{FirstName: "Li", LastName: "Lei"},
			wantErr: false,
			wantSubstr: []string{
				"first_name: Li #名",
				"last_name: Lei #姓",
			},
		},
	}

	for _, tc := range tests {
		tc := tc // capture
		t.Run(tc.name, func(t *testing.T) {
			out, err := MarshalWithComment(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("MarshalWithComment() error = %v, wantErr %v", err, tc.wantErr)
			}
			yml := string(out)
			for _, substr := range tc.wantSubstr {
				if !strings.Contains(yml, substr) {
					t.Errorf("output missing %q; got:\n%s", substr, yml)
				}
			}
		})
	}
}
