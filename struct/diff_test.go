package _struct

import (
	"errors"
	"reflect"
	"testing"
)

// 测试用例辅助类型
type Person struct {
	Name   string
	Age    int
	hidden string // 不可导出字段，应被跳过
}

func TestDiffStruct(t *testing.T) {
	tests := []struct {
		name      string
		a, b      any
		ignore    []string
		wantDiff  map[string][2]string
		wantError error
	}{
		{
			name:     "完全相同的结构体",
			a:        Person{"Alice", 30, "x"},
			b:        Person{"Alice", 30, "y"},
			ignore:   nil,
			wantDiff: map[string][2]string{},
		},
		{
			name: "字段不同",
			a:    Person{"Bob", 25, ""},
			b:    Person{"Bob", 26, ""},
			wantDiff: map[string][2]string{
				"Age": {"25", "26"},
			},
		},
		{
			name:   "忽略某个字段",
			a:      Person{"Carol", 40, ""},
			b:      Person{"Carol", 41, ""},
			ignore: []string{"Age"},
			// 虽然 Age 不同，但被忽略，故无差异
			wantDiff: map[string][2]string{},
		},
		{
			name:      "类型不匹配",
			a:         Person{"D", 50, ""},
			b:         struct{ Name string }{"D"},
			wantError: errors.New("A与B不是同一个属性的结构体"),
		},
		{
			name: "指针输入也能处理",
			a:    &Person{"Eve", 20, ""},
			b:    &Person{"Eve", 21, ""},
			wantDiff: map[string][2]string{
				"Age": {"20", "21"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diff, err := DiffStruct(tc.a, tc.b, tc.ignore)
			if tc.wantError != nil {
				if err == nil || err.Error() != tc.wantError.Error() {
					t.Fatalf("预期错误 %v，实际 %v", tc.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("意外错误：%v", err)
			}
			if !reflect.DeepEqual(diff, tc.wantDiff) {
				t.Errorf("DiffStruct() = %v, want %v", diff, tc.wantDiff)
			}
		})
	}
}
