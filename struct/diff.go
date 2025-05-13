package _struct

import (
	"errors"
	"fmt"
	"reflect"
)

// 检测 structA 与 structB 的区别，返回字段名称及对应的 [A值, B值]
func DiffStruct(structA, structB any, ignoreKey []string) (map[string][2]string, error) {
	// 获取 structA 的值并解引用指针
	structAValue := reflect.ValueOf(structA)
	for structAValue.Kind() == reflect.Ptr {
		structAValue = structAValue.Elem()
	}
	// 获取 structB 的值并解引用指针
	structBValue := reflect.ValueOf(structB)
	for structBValue.Kind() == reflect.Ptr {
		structBValue = structBValue.Elem()
	}
	// 类型必须相同
	if structAValue.Type() != structBValue.Type() {
		return nil, errors.New("A与B不是同一个属性的结构体")
	}
	ty := structAValue.Type()

	// 差异结果: 键为字段名称，值为 [A字段值, B字段值]
	out := make(map[string][2]string, ty.NumField())
	// 构建忽略字段集合
	ignoreMap := make(map[string]struct{}, len(ignoreKey))
	for _, key := range ignoreKey {
		ignoreMap[key] = struct{}{}
	}

	// 遍历字段并比较
	for i := 0; i < ty.NumField(); i++ {
		field := ty.Field(i)
		// 跳过忽略字段及未导出字段
		if _, skip := ignoreMap[field.Name]; skip || field.PkgPath != "" {
			continue
		}
		// 获取字段值
		aVal := structAValue.Field(i).Interface()
		bVal := structBValue.Field(i).Interface()
		strA := fmt.Sprintf("%v", aVal)
		strB := fmt.Sprintf("%v", bVal)
		if strA != strB {
			out[field.Name] = [2]string{strA, strB}
		}
	}
	return out, nil
}
