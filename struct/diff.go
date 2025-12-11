package _struct

import (
	"errors"
	"reflect"
)

type DiffData struct {
	Field  reflect.StructField
	ValueA reflect.Value
	ValueB reflect.Value
}

// 检测 structA 与 structB 的区别，返回字段名称及对应的 [A值, B值]
func DiffStruct(structA, structB any, ignoreKey []string) ([]*DiffData, error) {
	// 获取 structA 的值并解引用指针
	structAValue := reflect.ValueOf(structA)
	for structAValue.Kind() == reflect.Pointer {
		structAValue = structAValue.Elem()
	}
	// 获取 structB 的值并解引用指针
	structBValue := reflect.ValueOf(structB)
	for structBValue.Kind() == reflect.Pointer {
		structBValue = structBValue.Elem()
	}
	// 类型必须相同
	if structAValue.Type() != structBValue.Type() {
		return nil, errors.New("A与B不是同一个属性的结构体")
	}
	ty := structAValue.Type()

	// 差异结果: 键为字段名称，值为 [A字段值, B字段值]
	out := make([]*DiffData, 0, ty.NumField())
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
		aVal := structAValue.Field(i)
		bVal := structBValue.Field(i)
		if !aVal.Equal(bVal) {
			out = append(out, &DiffData{
				Field:  field,
				ValueA: aVal,
				ValueB: bVal,
			})
		}
	}
	return out, nil
}
