package yaml

import (
	"errors"
	"gopkg.in/yaml.v3"
	"reflect"
)

// MarshalWithComment 带注释的序列化
func MarshalWithComment(v interface{}) ([]byte, error) {
	var node yaml.Node
	node.Encode(v)
	vv := reflect.TypeOf(v)
	for vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}
	if vv.Kind() != reflect.Struct {
		return nil, errors.New("必须要Struct")
	}
	rangeStruct(vv, &node)
	return yaml.Marshal(node)
}
func rangeStruct(v reflect.Type, node *yaml.Node) {
	if v.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < v.NumField(); i++ {
		vv := v.Field(i).Type

		for vv.Kind() == reflect.Ptr {
			vv = vv.Elem()
		}

		var index = node
		var n = (i * 2) + 1
		if vv.Kind() == reflect.Struct {
			if node.Content[n].Tag != "!!null" {
				rangeStruct(vv, node.Content[n])
			}
			if vv.NumField() != 0 {
				n -= 1
			}
		}
		if len(node.Content) != 0 {
			index = node.Content[n]
		}
		if c := v.Field(i).Tag.Get("comment"); c != "" {
			index.LineComment = c
		}
	}
}
