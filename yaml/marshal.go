package yaml

import (
	"bytes"
	"reflect"

	"gopkg.in/yaml.v3"
)

// MarshalWithComment 带注释的序列化，自动去除注释前多余空格
func MarshalWithComment(v any) ([]byte, error) {
	var tmp bytes.Buffer
	encoder := yaml.NewEncoder(&tmp)
	encoder.SetIndent(2)

	var node yaml.Node
	// 对 v 进行节点编码
	node.Encode(v)

	// 获取实际类型
	vv := reflect.TypeOf(v)
	for vv.Kind() == reflect.Pointer {
		vv = vv.Elem()
	}

	// 如果不是结构体，则直接编码 v
	if vv.Kind() != reflect.Struct {
		if err := encoder.Encode(v); err != nil {
			return nil, err
		}
		return tmp.Bytes(), nil
	}

	// 结构体：递归添加注释到节点
	rangeStruct(vv, &node)

	// 将带注释的节点编码
	if err := encoder.Encode(node); err != nil {
		return nil, err
	}
	return tmp.Bytes(), nil
}

// rangeStruct 对结构体类型和对应节点进行遍历，添加行注释
func rangeStruct(v reflect.Type, node *yaml.Node) {
	if v.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < v.NumField(); i++ {
		// 处理嵌套指针
		vv := v.Field(i).Type
		for vv.Kind() == reflect.Pointer {
			vv = vv.Elem()
		}

		// 找到对应的节点索引位置
		index := node
		n := (i * 2) + 1
		if vv.Kind() == reflect.Struct {
			if node.Content[n].Tag != "!!null" {
				rangeStruct(vv, node.Content[n])
			}
			if vv.NumField() != 0 {
				n--
			}
		}
		if len(node.Content) != 0 {
			index = node.Content[n]
		}

		// 根据 tag 添加注释
		if c := v.Field(i).Tag.Get("comment"); c != "" {
			index.LineComment = c
		}
	}
}
