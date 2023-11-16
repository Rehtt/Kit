package yaml

import (
	"github.com/Rehtt/Kit/value"
	"reflect"
)

func GenYamlTemplate(v any) ([]byte, error) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		val = reflect.New(val.Type())
		val.Elem().Set(reflect.ValueOf(v))
	}
	value.InitValue(val.Interface())
	return MarshalWithComment(val.Interface())
}
