package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// IntSlice is a []int that can be stored as JSON in SQL databases.
type IntSlice []int

// Value implements driver.Valuer so IntSlice can be persisted by GORM.
func (s IntSlice) Value() (driver.Value, error) {
	b, err := json.Marshal([]int(s))
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements sql.Scanner for IntSlice.
func (s *IntSlice) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return fmt.Errorf("cannot scan %T", value)
	}
}
