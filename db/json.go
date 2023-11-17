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
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	}
	return json.Unmarshal(data, &j.data)
}

func (j JSON[T]) Value() (value driver.Value, err error) {
	src, err := json.Marshal(j.data)
	return src, err
}

func (j JSON[T]) Get() T {
	return j.data
}

func (j *JSON[T]) Set(data T) {
	j.data = data
}
