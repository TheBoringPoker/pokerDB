package models

import (
	"errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"math/rand"
	"pokerDB/pkg/constants"
	"time"
)

type Game struct {
	ID            uuid.UUID `json:"id" gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	TableID       uuid.UUID `json:"table_id" gorm:"type:uuid"`
	CardSequence  []int     `json:"card_sequence" gorm:"type:integer[]"`
	StartedTime   time.Time `json:"started_time" gorm:"type:timestamp"`
	EndedTime     time.Time `json:"ended_time" gorm:"type:timestamp"`
	PersonCount   int       `json:"person_count" gorm:"type:integer"`
	NextCardIndex int       `json:"-" gorm:"-"`
}

func NewGame(tableID uuid.UUID, personCount int) *Game {
	return &Game{
		TableID:      tableID,
		CardSequence: constants.CardSequence,
		PersonCount:  personCount,
	}
}

func (g *Game) Started() bool {
	return !g.StartedTime.IsZero()
}

func (g *Game) Ended() bool {
	return !g.EndedTime.IsZero()
}

func (g *Game) Start() error {
	if g.Started() {
		logrus.Warn("Game already started")
		return errors.New("game already started")
	}

	if g.Ended() {
		logrus.Warn("Game already ended")
		return errors.New("game already ended")
	}

	if g.PersonCount < 2 {
		logrus.Warn("Not enough players")
		return errors.New("not enough players")
	}

	if g.PersonCount > 10 {
		logrus.Warn("Too many players")
		return errors.New("too many players")
	}

	g.StartedTime = time.Now()
	shuffled := make([]int, len(constants.CardSequence))
	copy(shuffled, constants.CardSequence)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	logrus.Info("Shuffled Sequences: ", shuffled)
	g.CardSequence = shuffled
	g.NextCardIndex = 0
	return nil
}

func (g *Game) End() error {
	if !g.Started() {
		logrus.Warn("Game not started")
		return errors.New("game not started")
	}
	if g.Ended() {
		logrus.Warn("Game already ended")
		return errors.New("game already ended")
	}
	g.EndedTime = time.Now()
	return nil
}

// Deal removes the first count cards from the game's card sequence and returns
// them. If there are not enough cards remaining, an empty slice is returned.
// Deal returns the next 'count' cards from the sequence without modifying the
// underlying deck. The index is advanced so subsequent calls return the next
// cards in order. If there are not enough cards remaining, an empty slice is
// returned.
func (g *Game) Deal(count int) []int {
	if g.NextCardIndex+count > len(g.CardSequence) {
		return []int{}
	}
	cards := g.CardSequence[g.NextCardIndex : g.NextCardIndex+count]
	g.NextCardIndex += count
	return cards
}

// DealHands deals two cards to each player and returns a slice of hands where
// each hand contains two cards for one player.
// DealHands returns a slice of player hands without altering the card
// sequence, relying on Deal to advance the index.
func (g *Game) DealHands() [][]int {
	hands := make([][]int, g.PersonCount)
	for i := 0; i < g.PersonCount; i++ {
		hands[i] = g.Deal(2)
	}
	return hands
}
