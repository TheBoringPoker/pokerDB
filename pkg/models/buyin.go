package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// BuyIn records the starting chip amount a player brings to a game.
type BuyIn struct {
	PlayerID uuid.UUID `json:"player_id"`
	Amount   int64     `json:"amount"`
}

// BuyInList is a JSON serializable slice of BuyIns.
type BuyInList []BuyIn

// Value implements driver.Valuer so BuyInList can be persisted by GORM.
func (b BuyInList) Value() (driver.Value, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// Scan implements sql.Scanner for BuyInList.
func (b *BuyInList) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, b)
	case string:
		return json.Unmarshal([]byte(v), b)
	default:
		return fmt.Errorf("cannot scan %T", value)
	}
}
