package _struct

import (
	"fmt"
	"reflect"
)

func GetTag(s any, key string) (map[string]any, error) {
	ty := reflect.TypeOf(s)
	if ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
	}
	if ty.Kind() != reflect.Struct {
		return nil, fmt.Errorf("必须传入Struct")
	}
	out := make(map[string]any, ty.NumField())
	for i := 0; i < ty.NumField(); i++ {
		f := ty.Field(i)
		out[f.Name] = f.Tag.Get(key)
	}
	return out, nil
}
