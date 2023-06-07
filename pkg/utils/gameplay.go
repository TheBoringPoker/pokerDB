package utils

import (
	"pokerDB/pkg/models"
)

type Kind int

const (
	Spade Kind = iota
	Heart
	Diamond
	Club
)

type Card struct {
	Kind Kind
	Num  int
}

func NewCard(num int) *Card {
	return &Card{
		Kind: Kind(num / 13),
		Num:  num % 13,
	}
}

type GameResolver struct {
	Game *models.Game `json:"game"`
}
