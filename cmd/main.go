// Command line demonstration program for the pokerDB toolkit.
// Generated with OpenAI Codex; functionality is not guaranteed.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"pokerDB/pkg/models"
	"pokerDB/pkg/utils"
	"time"
)

func main() {
	game := models.Game{
		ID:           uuid.UUID{},
		TableID:      uuid.UUID{},
		CardSequence: nil,
		StartedTime:  time.Time{},
		EndedTime:    time.Time{},
		PersonCount:  3,
	}

	jsonData, _ := json.Marshal(game)
	fmt.Println(string(jsonData))

	if err := game.Start(); err != nil {
		logrus.Debug("Game start error: ", err)
		return
	}

	fmt.Println("Game started at", game.StartedTime.Format(time.RFC3339))

	hands := game.DealHands()
	for i, h := range hands {
		fmt.Printf("Player %d: %s %s\n", i+1, utils.CardToString(h[0]), utils.CardToString(h[1]))
	}

	flop := game.Deal(3)
	fmt.Printf("Flop: %s %s %s\n", utils.CardToString(flop[0]), utils.CardToString(flop[1]), utils.CardToString(flop[2]))

	turn := game.Deal(1)
	fmt.Printf("Turn: %s\n", utils.CardToString(turn[0]))

	river := game.Deal(1)
	fmt.Printf("River: %s\n", utils.CardToString(river[0]))

	if err := game.End(); err != nil {
		logrus.Debug("Game end error: ", err)
		return
	}

	fmt.Println("Game ended at", game.EndedTime.Format(time.RFC3339))

}
