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

// MaxSeats defines the number of seats available at a table.
// Seat numbers must be between 1 and MaxSeats.
const MaxSeats = 9

type Game struct {
	ID              uuid.UUID           `json:"id" gorm:"primary_key;type:uuid"`
	TableID         uuid.UUID           `json:"table_id" gorm:"type:uuid"`
	CardSequence    IntSlice            `json:"card_sequence" gorm:"type:json"`
	StartedTime     time.Time           `json:"started_time" gorm:"type:timestamp"`
	EndedTime       time.Time           `json:"ended_time" gorm:"type:timestamp"`
	PersonCount     int                 `json:"person_count" gorm:"type:integer"`
	Ante            int64               `json:"ante" gorm:"type:bigint"`
	SmallBlind      int64               `json:"small_blind" gorm:"type:bigint"`
	BigBlind        int64               `json:"big_blind" gorm:"type:bigint"`
	AllowRunItTwice bool                `json:"allow_run_it_twice" gorm:"type:boolean"`
	AllowStraddle   bool                `json:"allow_straddle" gorm:"type:boolean"`
	MinBuyIn        int64               `json:"min_buy_in" gorm:"type:bigint"`
	MaxBuyIn        int64               `json:"max_buy_in" gorm:"type:bigint"`
	BuyIns          BuyInList           `json:"buy_ins" gorm:"type:json"`
	ActionLog       ActionLog           `json:"action_log" gorm:"type:json"`
	Ledgers         []Ledger            `json:"ledgers"`
	NextCardIndex   int                 `json:"-" gorm:"-"`
	CurrentRound    int                 `json:"current_round" gorm:"-"`
	CurrentDealer   int                 `json:"current_dealer" gorm:"-"`
	Stacks          map[uuid.UUID]int64 `json:"-" gorm:"-"`
	Seats           map[uuid.UUID]int   `json:"-" gorm:"-"`
	NextSeats       map[uuid.UUID]int   `json:"-" gorm:"-"`
	currentBets     map[uuid.UUID]int64 `json:"-" gorm:"-"`
	inRound         bool                `json:"-" gorm:"-"`
	mu              sync.RWMutex        `json:"-" gorm:"-"`
}

func NewGame(tableID uuid.UUID, personCount int) *Game {
	return &Game{
		ID:           uuid.New(),
		TableID:      tableID,
		CardSequence: IntSlice(constants.CardSequence),
		PersonCount:  personCount,
		ActionLog:    ActionLog{},
		BuyIns:       BuyInList{},
		Stacks:       make(map[uuid.UUID]int64),
		Seats:        make(map[uuid.UUID]int),
		NextSeats:    make(map[uuid.UUID]int),
		currentBets:  make(map[uuid.UUID]int64),
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

	if g.PersonCount > MaxSeats {
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
	g.CurrentRound = 1
	g.CurrentDealer = 0
	g.inRound = true
	g.currentBets = make(map[uuid.UUID]int64)
	g.Stacks = make(map[uuid.UUID]int64)
	for _, b := range g.BuyIns {
		g.Stacks[b.PlayerID] = b.Amount
	}
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
	g.inRound = false
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

// endNoLock finalizes the game without locking. The caller must hold the mutex.
func (g *Game) endNoLock() {
	if !g.EndedTime.IsZero() {
		return
	}
	g.EndedTime = time.Now()
	g.inRound = false
	pairs := make([]string, len(g.Ledgers))
	for i, l := range g.Ledgers {
		id := shortID(l.PlayerID)
		pairs[i] = fmt.Sprintf("%s=%d", id, l.Balance)
	}
	endEntry := fmt.Sprintf("E:%s,%d", strings.Join(pairs, ":"), g.EndedTime.Unix())
	g.ActionLog = append(g.ActionLog, endEntry)
}

// EndRound finishes the current round and rotates the dealer position.
func (g *Game) EndRound() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.inRound {
		return errors.New("round not active")
	}
	g.inRound = false
	g.currentBets = make(map[uuid.UUID]int64)
	g.CurrentDealer = (g.CurrentDealer + 1) % g.PersonCount
	return nil
}

// StartRound begins a new round after the previous one has ended.
func (g *Game) StartRound() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.StartedTime.IsZero() {
		return errors.New("game not started")
	}
	if g.inRound {
		return errors.New("round already active")
	}
	g.CurrentRound++
	g.inRound = true
	g.currentBets = make(map[uuid.UUID]int64)
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
	if !g.EndedTime.IsZero() {
		return errors.New("game already ended")
	}
	if amount < g.MinBuyIn || (g.MaxBuyIn > 0 && amount > g.MaxBuyIn) {
		return fmt.Errorf("buy-in must be between %d and %d", g.MinBuyIn, g.MaxBuyIn)
	}
	id := playerID.String()
	if len(id) > 8 {
		id = id[:8]
	}
	g.BuyIns = append(g.BuyIns, BuyIn{PlayerID: playerID, Amount: amount})
	g.Stacks[playerID] += amount
	entry := fmt.Sprintf("%s%s%d,%d", id, ActionBuyIn, amount, time.Now().Unix())
	g.ActionLog = append(g.ActionLog, entry)
	return nil
}

// Join adds a player to the table. The join action is recorded and
// the player count is updated. Players may join at any time.
func (g *Game) Join(playerID uuid.UUID) error {
	g.mu.Lock()
	if _, ok := g.Seats[playerID]; ok {
		g.mu.Unlock()
		return fmt.Errorf("player %s already joined", playerID)
	}
	g.Seats[playerID] = -1
	g.PersonCount = len(g.Seats)
	entry := fmt.Sprintf("%s%s0,%d", shortID(playerID), ActionJoin, time.Now().Unix())
	g.ActionLog = append(g.ActionLog, entry)
	g.mu.Unlock()
	return nil
}

// Quit removes a player from the table and records the action. If no
// players remain the game is automatically ended.
func (g *Game) Quit(playerID uuid.UUID) error {
	g.mu.Lock()
	if _, ok := g.Seats[playerID]; !ok {
		g.mu.Unlock()
		return fmt.Errorf("unknown player %s", playerID)
	}
	delete(g.Seats, playerID)
	delete(g.Stacks, playerID)
	delete(g.currentBets, playerID)
	g.PersonCount = len(g.Seats)
	entry := fmt.Sprintf("%s%s0,%d", shortID(playerID), ActionQuit, time.Now().Unix())
	g.ActionLog = append(g.ActionLog, entry)
	shouldEnd := g.PersonCount == 0 && g.EndedTime.IsZero()
	g.mu.Unlock()
	if shouldEnd {
		if g.Started() {
			return g.End()
		}
		g.mu.Lock()
		if g.EndedTime.IsZero() {
			g.endNoLock()
		}
		g.mu.Unlock()
	}
	return nil
}

// ChooseSeat records a seat selection for the next game. The change
// does not affect the current game but is logged for historical
// purposes.
func (g *Game) ChooseSeat(playerID uuid.UUID, seat int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.Seats[playerID]; !ok {
		return fmt.Errorf("unknown player %s", playerID)
	}
	if seat < 1 || seat > MaxSeats {
		return fmt.Errorf("invalid seat %d", seat)
	}
	for pid, s := range g.NextSeats {
		if pid != playerID && s == seat {
			return fmt.Errorf("seat %d already taken", seat)
		}
	}
	g.NextSeats[playerID] = seat
	entry := fmt.Sprintf("%s%s%d,%d", shortID(playerID), ActionSeat, seat, time.Now().Unix())
	g.ActionLog = append(g.ActionLog, entry)
	return nil
}

// AddAction appends a new action to the game.
func (g *Game) AddAction(playerID uuid.UUID, code string, amount int64) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.inRound {
		return errors.New("no active round")
	}
	stack, ok := g.Stacks[playerID]
	if !ok {
		return fmt.Errorf("unknown player %s", playerID)
	}
	need := amount - g.currentBets[playerID]
	if need < 0 {
		need = 0
	}
	if need > stack {
		return fmt.Errorf("insufficient chips")
	}
	g.Stacks[playerID] = stack - need
	g.currentBets[playerID] = amount

	id := playerID.String()
	if len(id) > 8 {
		id = id[:8]
	}
	entry := fmt.Sprintf("%s%s%d,%d", id, code, amount, time.Now().Unix())
	g.ActionLog = append(g.ActionLog, entry)
	return nil
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

func shortID(id uuid.UUID) string {
	s := id.String()
	if len(s) > 8 {
		return s[:8]
	}
	return s
}
