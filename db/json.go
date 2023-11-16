package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// json类型

type JSON json.RawMessage

func (JSON) GormDataType() string {
	return "json"
}

func (j *JSON) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	j1 := json.RawMessage{}
	err := json.Unmarshal(bytes, &j1)
	*j = JSON(j1)
	return err
}

func (j JSON) Value() (value driver.Value, err error) {
	return json.RawMessage(j).MarshalJSON()
}
