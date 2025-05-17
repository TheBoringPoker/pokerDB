package models

import "github.com/google/uuid"

// Ledger stores the final balance of a player after a game.
type Ledger struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key"`
	GameID   uuid.UUID `gorm:"type:uuid;index"`
	PlayerID uuid.UUID `gorm:"type:uuid"`
	Balance  int64
}
