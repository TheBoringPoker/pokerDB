// Package models defines the data structures for pokerDB.
// Generated with OpenAI Codex; provided without warranty.
package models

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"math/rand"
	"pokerDB/pkg/constants"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Game struct {
	ID              uuid.UUID    `json:"id" gorm:"primary_key;type:uuid"`
	TableID         uuid.UUID    `json:"table_id" gorm:"type:uuid"`
	CardSequence    IntSlice     `json:"card_sequence" gorm:"type:json"`
	StartedTime     time.Time    `json:"started_time" gorm:"type:timestamp"`
	EndedTime       time.Time    `json:"ended_time" gorm:"type:timestamp"`
	PersonCount     int          `json:"person_count" gorm:"type:integer"`
	Ante            int64        `json:"ante" gorm:"type:bigint"`
	SmallBlind      int64        `json:"small_blind" gorm:"type:bigint"`
	BigBlind        int64        `json:"big_blind" gorm:"type:bigint"`
	AllowRunItTwice bool         `json:"allow_run_it_twice" gorm:"type:boolean"`
	AllowStraddle   bool         `json:"allow_straddle" gorm:"type:boolean"`
	MinBuyIn        int64        `json:"min_buy_in" gorm:"type:bigint"`
	MaxBuyIn        int64        `json:"max_buy_in" gorm:"type:bigint"`
	BuyIns          BuyInList    `json:"buy_ins" gorm:"type:json"`
	ActionLog       ActionLog    `json:"action_log" gorm:"type:json"`
	Ledgers         []Ledger     `json:"ledgers"`
	NextCardIndex   int          `json:"-" gorm:"-"`
	mu              sync.RWMutex `json:"-" gorm:"-"`
}

func NewGame(tableID uuid.UUID, personCount int) *Game {
	return &Game{
		ID:           uuid.New(),
		TableID:      tableID,
		CardSequence: IntSlice(constants.CardSequence),
		PersonCount:  personCount,
		ActionLog:    ActionLog{},
		BuyIns:       BuyInList{},
	}
}

func (g *Game) Started() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return !g.StartedTime.IsZero()
}

func (g *Game) Ended() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return !g.EndedTime.IsZero()
}

func (g *Game) Start() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	// check timestamps directly while holding the lock to avoid
	// re-entrantly acquiring the mutex via Started() or Ended()
	if !g.StartedTime.IsZero() {
		logrus.Warn("Game already started")
		return errors.New("game already started")
	}

	if !g.EndedTime.IsZero() {
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

	if len(g.BuyIns) != g.PersonCount {
		logrus.Warn("buy-ins not complete")
		return errors.New("all players must buy in before start")
	}

	for _, b := range g.BuyIns {
		if b.Amount < g.MinBuyIn || (g.MaxBuyIn > 0 && b.Amount > g.MaxBuyIn) {
			return fmt.Errorf("invalid buy-in for player %s", b.PlayerID)
		}
	}

	g.StartedTime = time.Now()
	shuffled := make([]int, len(constants.CardSequence))
	copy(shuffled, constants.CardSequence)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	logrus.Info("Shuffled Sequences: ", shuffled)
	g.CardSequence = IntSlice(shuffled)
	g.NextCardIndex = 0
	startEntry := fmt.Sprintf("G:%d:%d:%d:%d:%d,%d", g.SmallBlind, g.BigBlind, g.Ante, boolToInt(g.AllowRunItTwice), boolToInt(g.AllowStraddle), g.StartedTime.Unix())
	g.ActionLog = append(g.ActionLog, startEntry)
	return nil
}

func (g *Game) End() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	// check timestamps directly to avoid re-entrant mutex locking
	if g.StartedTime.IsZero() {
		logrus.Warn("Game not started")
		return errors.New("game not started")
	}
	if !g.EndedTime.IsZero() {
		logrus.Warn("Game already ended")
		return errors.New("game already ended")
	}
	g.EndedTime = time.Now()
	pairs := make([]string, len(g.Ledgers))
	for i, l := range g.Ledgers {
		id := l.PlayerID.String()
		if len(id) > 8 {
			id = id[:8]
		}
		pairs[i] = fmt.Sprintf("%s=%d", id, l.Balance)
	}
	endEntry := fmt.Sprintf("E:%s,%d", strings.Join(pairs, ":"), g.EndedTime.Unix())
	g.ActionLog = append(g.ActionLog, endEntry)
	return nil
}

// Deal removes the first count cards from the game's card sequence and returns
// them. If there are not enough cards remaining, an empty slice is returned.
// Deal returns the next 'count' cards from the sequence without modifying the
// underlying deck. The index is advanced so subsequent calls return the next
// cards in order. If there are not enough cards remaining, an empty slice is
// returned.
func (g *Game) dealNoLock(count int) []int {
	if g.NextCardIndex+count > len(g.CardSequence) {
		return []int{}
	}
	seq := g.CardSequence[g.NextCardIndex : g.NextCardIndex+count]
	cards := make([]int, len(seq))
	for i, v := range seq {
		cards[i] = v
	}
	g.NextCardIndex += count
	return cards
}

func (g *Game) Deal(count int) []int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.dealNoLock(count)
}

// DealHands deals two cards to each player and returns a slice of hands where
// each hand contains two cards for one player.
// DealHands returns a slice of player hands without altering the card
// sequence, relying on Deal to advance the index.
func (g *Game) DealHands() [][]int {
	g.mu.Lock()
	defer g.mu.Unlock()
	hands := make([][]int, g.PersonCount)
	for i := 0; i < g.PersonCount; i++ {
		hands[i] = g.dealNoLock(2)
	}
	return hands
}

// BuyIn records a player's initial chip stack before the game starts.
func (g *Game) BuyIn(playerID uuid.UUID, amount int64) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.StartedTime.IsZero() {
		return errors.New("cannot buy-in after game start")
	}
	if amount < g.MinBuyIn || (g.MaxBuyIn > 0 && amount > g.MaxBuyIn) {
		return fmt.Errorf("buy-in must be between %d and %d", g.MinBuyIn, g.MaxBuyIn)
	}
	id := playerID.String()
	if len(id) > 8 {
		id = id[:8]
	}
	for i := range g.BuyIns {
		if g.BuyIns[i].PlayerID == playerID {
			g.BuyIns[i].Amount = amount
			goto log
		}
	}
	g.BuyIns = append(g.BuyIns, BuyIn{PlayerID: playerID, Amount: amount})
log:
	entry := fmt.Sprintf("%s%s%d,%d", id, ActionBuyIn, amount, time.Now().Unix())
	g.ActionLog = append(g.ActionLog, entry)
	return nil
}

// AddAction appends a new action to the game.
func (g *Game) AddAction(playerID uuid.UUID, code string, amount int64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	id := playerID.String()
	if len(id) > 8 {
		id = id[:8]
	}
	entry := fmt.Sprintf("%s%s%d,%d", id, code, amount, time.Now().Unix())
	g.ActionLog = append(g.ActionLog, entry)
}

// ActionStrings returns human readable lines describing actions in order.
func (g *Game) ActionStrings() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	lines := make([]string, len(g.ActionLog))
	for i, raw := range g.ActionLog {
		parts := strings.SplitN(raw, ",", 2)
		if len(parts) != 2 {
			lines[i] = raw
			continue
		}
		ts, _ := strconv.ParseInt(parts[1], 10, 64)
		body := parts[0]
		if strings.HasPrefix(body, "G:") {
			fields := strings.Split(body[2:], ":")
			if len(fields) >= 5 {
				lines[i] = fmt.Sprintf("start sb=%s bb=%s ante=%s runTwice=%s straddle=%s at %s", fields[0], fields[1], fields[2], fields[3], fields[4], time.Unix(ts, 0).Format(time.RFC3339))
				continue
			}
		} else if strings.HasPrefix(body, "E:") {
			fields := strings.Split(body[2:], ":")
			lines[i] = fmt.Sprintf("result %v at %s", fields, time.Unix(ts, 0).Format(time.RFC3339))
			continue
		} else if len(body) >= 9 {
			pid := body[:8]
			code := string(body[8])
			amt := body[9:]
			word := ActionToWord(code)
			lines[i] = fmt.Sprintf("%s %s %s at %s", pid, word, amt, time.Unix(ts, 0).Format(time.RFC3339))
			continue
		}
		lines[i] = raw
	}
	return lines
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
