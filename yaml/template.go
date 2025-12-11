package yaml

import (
	"reflect"

	"github.com/Rehtt/Kit/value"
)

func GenYamlTemplate(v any) ([]byte, error) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Pointer {
		val = reflect.New(val.Type())
		val.Elem().Set(reflect.ValueOf(v))
	}
	value.InitValue(val.Interface())
	return MarshalWithComment(val.Interface())
}
