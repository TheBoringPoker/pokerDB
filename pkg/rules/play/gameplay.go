package play

import (
	"pokerDB/pkg/models"
)

type GameResolver struct {
	Game *models.Game `json:"game"`
}
