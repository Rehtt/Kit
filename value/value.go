package value

import (
	"reflect"
)

func InitValue(v interface{}) {
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			fieldValue := val.Field(i)
			var ty = fieldValue.Type()

			// 防止死循环
			if ty.Kind() == reflect.Ptr && ty.Elem().Name() == val.Type().Name() {
				continue
			}
			InitValue(fieldValue.Addr().Interface())
		}
	}

	//todo test3{test3: []test3 } 死循环
	if val.Kind() == reflect.Slice {
		if val.Len() == 0 {
			l := 1
			val.Set(reflect.MakeSlice(val.Type(), l, l))
		}
		for i := 0; i < val.Len(); i++ {
			fieldValue := val.Index(i)
			InitValue(fieldValue.Addr().Interface())
		}
	}

}
