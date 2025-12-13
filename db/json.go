package db

import (
	"database/sql/driver"
	"encoding/json"
)

// json类型
type JSON[T any] struct {
	data T
}

func (JSON[T]) GormDataType() string {
	return "json"
}

func (j *JSON[T]) Scan(value any) error {
	if value == nil {
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		b, err := json.Marshal(value)
		if err != nil {
			return err
		}
		data = b
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, &j.data)
}

func (j JSON[T]) Value() (value driver.Value, err error) {
	return json.Marshal(j.data)
}

func (j JSON[T]) Get() T {
	return j.data
}

func (j *JSON[T]) Set(data T) {
	j.data = data
}

func (j JSON[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.data)
}

func (j *JSON[T]) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &j.data)
}

func (j JSON[T]) Convert() any {
	return j.data
}

func (j *JSON[T]) UnmarshalValue(value any) error {
	return j.Scan(value)
}

func NewJSON[T any](data T) JSON[T] {
	return JSON[T]{data}
}
