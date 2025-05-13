package _struct

import (
	"errors"
	"reflect"
	"strconv"
)

type (
	FieldName  string
	ParseFuncs map[FieldName]func(txt string) (any, error)
)

// ParseArrayStringStruct 解析字符串数组到结构体切片
func ParseArrayStringStruct(rows [][]string, data any, p ParseFuncs) error {
	rdata := reflect.ValueOf(data)
	if rdata.Kind() != reflect.Ptr {
		return errors.New("data must be &[]*struct/&[]struct")
	}
	rdata = rdata.Elem()
	if rdata.Type().Kind() != reflect.Slice {
		return errors.New("data must be &[]*struct/&[]struct")
	}
	dataElemType := rdata.Type().Elem()
	var isPrt bool
	if dataElemType.Kind() == reflect.Ptr {
		dataElemType = dataElemType.Elem()
		isPrt = true
	}
	if dataElemType.Kind() != reflect.Struct {
		return errors.New("data must be &[]*struct/&[]struct")
	}

	tmp := reflect.MakeSlice(reflect.SliceOf(rdata.Type().Elem()), 0, len(rows))
	for _, r := range rows {
		var elemPtr, elemVal reflect.Value
		if isPrt {
			elemPtr = reflect.New(dataElemType)
			elemVal = elemPtr.Elem()
		} else {
			elemVal = reflect.New(dataElemType).Elem()
		}
		for i, txt := range r {
			if i >= dataElemType.NumField() {
				break
			}
			f := elemVal.Field(i)
			if !f.CanSet() {
				continue
			}

			if p != nil && p[FieldName(dataElemType.Field(i).Name)] != nil {
				o, err := p[FieldName(dataElemType.Field(i).Name)](txt)
				if err != nil {
					return err
				}
				out := reflect.ValueOf(o)
				if out.Kind() != f.Kind() {
					return errors.New("func return type must be " + f.Kind().String())
				}
				f.Set(out)
				continue
			}

			switch f.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if n, err := strconv.ParseInt(txt, 10, 64); err == nil {
					f.SetInt(n)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if n, err := strconv.ParseUint(txt, 10, 64); err == nil {
					f.SetUint(n)
				}
			case reflect.Float32, reflect.Float64:
				if f64, err := strconv.ParseFloat(txt, 64); err == nil {
					f.SetFloat(f64)
				}
			case reflect.Bool:
				if b, err := strconv.ParseBool(txt); err == nil {
					f.SetBool(b)
				}
			case reflect.String:
				f.SetString(txt)
			}
		}
		if isPrt {
			tmp = reflect.Append(tmp, elemPtr)
		} else {
			tmp = reflect.Append(tmp, elemVal)
		}
	}
	rdata.Set(tmp)
	return nil
}
