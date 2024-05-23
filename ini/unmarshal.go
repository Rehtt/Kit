// 解析ini文件，支持将重复的key转为数组

package ini

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
)

type Decoder struct {
	r          io.Reader
	rawData    map[string]any
	decodeDone uint32
	m          sync.Mutex
}

func Unmarshal(data []byte, v any) error {
	return NewDecoder(bytes.NewReader(data)).Decode(v)
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) decodeData() error {
	// like sync.Once
	if atomic.LoadUint32(&d.decodeDone) != 0 {
		return nil
	}
	d.m.Lock()
	defer d.m.Unlock()
	if d.decodeDone != 0 {
		return nil
	}
	var err error
	defer func() {
		if err == nil {
			atomic.StoreUint32(&d.decodeDone, 1)
		}
	}()

	scann := bufio.NewScanner(d.r)
	d.rawData = make(map[string]any)
	var tmpGroupName string
	for scann.Scan() {
		txt := strings.Split(scann.Text(), ";")[0]
		txt = strings.TrimSpace(txt)
		if txt == "" {
			continue
		}
		if strings.HasPrefix(txt, "[") && strings.HasSuffix(txt, "]") {
			tmpGroupName = strings.TrimPrefix(txt, "[")
			tmpGroupName = strings.TrimSuffix(tmpGroupName, "]")
			continue
		}
		kv := strings.SplitN(txt, "=", 2)
		if len(kv) != 2 {
			err = fmt.Errorf("%s Unmarshal error", txt)
		}
		if tmpGroupName == "" {
			d.rawData[kv[0]] = kv[1]
			continue
		}

		m, ok := d.rawData[tmpGroupName].(map[string]any)
		if !ok {
			m = make(map[string]any)
		}
		m[kv[0]] = kv[1]
		d.rawData[tmpGroupName] = m
	}
	return err
}

type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "ini: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Pointer {
		return "ini: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "ini: Unmarshal(nil " + e.Type.String() + ")"
}

func (d *Decoder) Decode(v any) error {
	err := d.decodeData()
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	rt := reflect.TypeOf(v)
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	switch rv.Type().Kind() {
	case reflect.Map:
		if rv.IsNil() {
			return &InvalidUnmarshalError{reflect.TypeOf(v)}
		}
		err = d.decodeToMap(d.rawData, rv)
	case reflect.Struct:
		err = d.decodeToStruct(d.rawData, rt, rv)
	}
	return err
}

func (d *Decoder) Show() {
	fmt.Println(d.rawData)
}

func (d *Decoder) decodeToStruct(rawData map[string]any, rt reflect.Type, rv reflect.Value) error {
	var err error
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		raw, ok := rawData[field.Name]
		if !ok {
			raw = rawData[field.Tag.Get("ini")]
		}
		if raw == nil {
			continue
		}

		if r, ok := raw.(string); ok {
			switch field.Type.Kind() {
			case reflect.String:
				rv.Field(i).SetString(r)
			}
		} else if r, ok := raw.(map[string]any); ok {
			switch field.Type.Kind() {
			case reflect.Struct:
				err = d.decodeToStruct(r, field.Type, rv.Field(i))
			case reflect.Map:
				err = d.decodeToMap(r, rv.Field(i))
			}
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (d *Decoder) decodeToMap(rawData map[string]any, rv reflect.Value) error {
	for key, value := range rawData {
		keyV := reflect.ValueOf(key)
		valueV := reflect.ValueOf(value)

		tmp := valueV
		if valueV.Kind() == reflect.Map {
			// 新创建一个对象
			tmp = reflect.MakeMap(reflect.TypeOf(map[string]string{}))
			for subKey, subValue := range value.(map[string]any) {
				tmp.SetMapIndex(reflect.ValueOf(subKey), reflect.ValueOf(subValue))
			}
		}
		rv.SetMapIndex(keyV, tmp)
	}
	return nil
}
