package strings

import (
	"fmt"
	"reflect"
	"strings"
)

// JoinToString 数组加分隔符号转string
func JoinToString(elems any, sep string) string {
	if e, ok := elems.([]string); ok {
		return strings.Join(e, sep)
	}
	switch reflect.TypeOf(elems).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(elems)
		switch s.Len() {
		case 0:
			return ""
		case 1:
			return fmt.Sprintf("%v", s.Index(0))
		}
		n := len(sep) * (s.Len() - 1)
		if reflect.TypeOf(s.Index(0)).Kind() == reflect.String {
			for i := 0; i < s.Len(); i++ {
				n += s.Index(i).Len()
			}
		}
		var b strings.Builder
		b.Grow(n)
		b.WriteString(fmt.Sprintf("%v", s.Index(0)))

		for i := 1; i < s.Len(); i++ {
			b.WriteString(sep)
			b.WriteString(fmt.Sprintf("%v", s.Index(i)))
		}

		return b.String()
	}
	return ""

}
