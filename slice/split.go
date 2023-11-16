package slice

import (
	"reflect"
)

// Split 将一个切片划分为多个大小为len的切片
func Split(data any, len int) (out []any) {
	if len < 1 {
		return
	}
	ty := reflect.ValueOf(data)
	if ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
	}
	l := len
	for i := 0; i < ty.Len(); {
		if i+len >= ty.Len() {
			len = ty.Len() - i
		}
		var da reflect.Value
		if reflect.TypeOf(data).Kind() == reflect.Ptr {
			da = reflect.New(ty.Slice(i, i+len).Type())
			da.Elem().Set(ty.Slice(i, i+len))
		} else {
			da = reflect.ValueOf(ty.Slice(i, i+len).Interface())
		}
		i += len
		if i%l == 0 || l != len {
			out = append(out, da.Interface())
		}
	}
	return
}
