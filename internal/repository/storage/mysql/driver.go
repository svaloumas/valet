package mysql

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// MapStringInterface is used to insert and fetch task params type of
// map[string]interface{} as JSON to and from MySQL.
type MapStringInterface map[string]interface{}

// https://golang.org/pkg/database/sql/driver/#Valuer
func (ss MapStringInterface) Value() (driver.Value, error) {
	value, err := json.Marshal(ss)
	return string(value), err
}

// https://golang.org/pkg/database/sql/#Scanner
func (ss *MapStringInterface) Scan(src interface{}) error {
	val, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan field value")
	}

	var jsonMapStringInterface map[string]interface{}
	err := json.Unmarshal(val, &jsonMapStringInterface)
	if err != nil {
		return err
	}

	*ss = jsonMapStringInterface

	return nil
}
