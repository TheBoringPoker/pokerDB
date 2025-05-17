package models

import (
	"github.com/google/uuid"
	"time"
)

type Table struct {
	ID          uuid.UUID `json:"id" gorm:"primary_key;type:uuid"`
	Games       []Game
	StartedTime time.Time `json:"started_time" gorm:"type:timestamp"`
	EndedTime   time.Time `json:"ended_time" gorm:"type:timestamp"`
}
