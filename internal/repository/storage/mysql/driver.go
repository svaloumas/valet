package mysql

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// MapInterface is used to insert and fetch JSON to and from MySQ MySQL.
type MapInterface map[string]interface{}

// https://golang.org/pkg/database/sql/driver/#Valuer
func (ss MapInterface) Value() (driver.Value, error) {
	value, err := json.Marshal(ss)
	return string(value), err
}

// https://golang.org/pkg/database/sql/#Scanner
func (ss *MapInterface) Scan(src interface{}) error {
	val, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan field value")
	}

	var jsonMapInterface map[string]interface{}
	err := json.Unmarshal(val, &jsonMapInterface)
	if err != nil {
		return err
	}

	*ss = jsonMapInterface

	return nil
}
