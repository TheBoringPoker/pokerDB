package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// IntSlice is a []int that implements Scanner and Valuer so it can be
// stored uniformly across SQL backends. Values are persisted as JSON.
type IntSlice []int

// Value implements driver.Valuer.
func (s IntSlice) Value() (driver.Value, error) {
	b, err := json.Marshal([]int(s))
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements sql.Scanner.
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
