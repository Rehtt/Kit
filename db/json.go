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
	return json.Unmarshal(value.([]byte), &j.data)
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
