package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Compact action codes used in action log entries.
const (
	ActionRaise    = "R" // player raises
	ActionFold     = "F" // player folds
	ActionCheck    = "C" // player checks
	ActionAllIn    = "A" // player goes all in
	ActionStraddle = "S" // player posts a straddle
	ActionRunTwice = "T" // players choose run it twice/once
)

// ActionWords maps short action codes to fully spelled words used
// when formatting a game's action log.
var ActionWords = map[string]string{
	ActionRaise:    "raise",
	ActionFold:     "fold",
	ActionCheck:    "check",
	ActionAllIn:    "all-in",
	ActionStraddle: "straddle",
	ActionRunTwice: "run-twice",
}

// ActionToWord returns a human readable word for the given action
// code. Unknown codes are returned unchanged.
func ActionToWord(code string) string {
	if w, ok := ActionWords[code]; ok {
		return w
	}
	return code
}

// ActionLog stores encoded action strings. It is persisted as JSON
// in the database so that different SQL backends can store it uniformly.
type ActionLog []string

// Value implements the driver.Valuer interface so ActionLog can be stored by GORM.
func (a ActionLog) Value() (driver.Value, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements the sql.Scanner interface for loading ActionLog from the DB.
func (a *ActionLog) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	case string:
		return json.Unmarshal([]byte(v), a)
	default:
		return fmt.Errorf("cannot scan %T", value)
	}
}
