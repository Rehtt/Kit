package _struct

import (
	"errors"
	"fmt"
	"reflect"
)

// 检测 structA 与 structB 的区别
func DiffStruct(structA, structB interface{}, ignoreKey []string) (map[*reflect.StructField][2]string, error) {
	getValue := func(a interface{}) (reflect.Value, reflect.Type) {
		ref := reflect.ValueOf(a)
		ty := reflect.TypeOf(a)
		for ref.Kind() == reflect.Ptr {
			ref = ref.Elem()
			ty = ty.Elem()
		}
		return ref, ty
	}

	structAValue, ty := getValue(structA)
	structBValue, _ := getValue(structB)
	if structAValue.Type() != structBValue.Type() {
		return nil, errors.New("A与B不是同一个属性的结构体")
	}

	out := make(map[*reflect.StructField][2]string)

	ignoreKeyMap := make(map[string]struct{})
	for _, key := range ignoreKey {
		ignoreKeyMap[key] = struct{}{}
	}

	for i := 0; i < structAValue.NumField(); i++ {
		key := ty.Field(i).Name
		if _, ok := ignoreKeyMap[key]; ok {
			continue
		}
		va, _ := getValue(structAValue.Field(i).Interface())
		vb, _ := getValue(structBValue.Field(i).Interface())
		var valueA, valueB string
		// if va.IsValid() && !va.IsNil() {
		valueA = fmt.Sprintf("%v", va)
		// }
		// if vb.IsValid() && !vb.IsNil() {
		valueB = fmt.Sprintf("%v", vb)
		// }
		if valueA != valueB {
			field := ty.Field(i)
			out[&field] = [2]string{
				valueA,
				valueB,
			}
		}
	}
	return out, nil
}
