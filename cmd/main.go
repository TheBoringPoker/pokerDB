package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"pokerDB/pkg/models"
	"time"
)

func main() {
	// ...
	game := models.Game{
		ID:           uuid.UUID{},
		TableID:      uuid.UUID{},
		CardSequence: nil,
		StartedTime:  time.Time{},
		EndedTime:    time.Time{},
		PersonCount:  3,
	}
	jsonData, _ := json.Marshal(game)
	println(string(jsonData))

	err := game.Start()
	if err != nil {
		logrus.Debug("Game start error: ", err)
		return
	}
	jsonData, _ = json.Marshal(game)
	println(string(jsonData))

}
