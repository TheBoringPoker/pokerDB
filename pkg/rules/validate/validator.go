package validate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"pokerDB/pkg/models"
)

// ValidationError describes a problem found while checking a game's action log.
type ValidationError struct {
	Index int    // zero-based index of the entry in ActionLog
	Entry string // raw action log entry
	Err   error  // underlying error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("entry %d %q: %v", e.Index, e.Entry, e.Err)
}

func shortID(id uuid.UUID) string {
	s := id.String()
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

// Validate checks the recorded actions of a game. The optional stacks map
// contains the starting chip count for each player, keyed by the truncated
// player ID used in the action log. When provided, chip amounts are verified
// against raises and calls.
func Validate(g *models.Game, stacks map[string]int64) error {
	if len(g.ActionLog) < 2 {
		return fmt.Errorf("action log too short")
	}

	startIdx := -1
	for i, entry := range g.ActionLog {
		if strings.HasPrefix(entry, "G:") {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		return fmt.Errorf("missing start entry")
	}
	if !strings.HasPrefix(g.ActionLog[len(g.ActionLog)-1], "E:") {
		return fmt.Errorf("missing end entry")
	}

	// track current highest bet and minimum raise size
	currentBet := g.BigBlind
	lastRaiseDelta := g.BigBlind

	playerBets := make(map[string]int64)
	if stacks == nil {
		stacks = make(map[string]int64)
	}

	for i, entry := range g.ActionLog[startIdx+1 : len(g.ActionLog)-1] {
		idx := i + startIdx + 1
		parts := strings.SplitN(entry, ",", 2)
		if len(parts) != 2 {
			return &ValidationError{Index: idx, Entry: entry, Err: fmt.Errorf("malformed entry")}
		}
		body := parts[0]
		if len(body) < 9 {
			return &ValidationError{Index: idx, Entry: entry, Err: fmt.Errorf("invalid body")}
		}
		pid := body[:8]
		code := string(body[8])
		amtStr := body[9:]
		amount, err := strconv.ParseInt(amtStr, 10, 64)
		if err != nil {
			return &ValidationError{Index: idx, Entry: entry, Err: fmt.Errorf("bad amount")}
		}

		switch code {
		case models.ActionRaise:
			if amount <= currentBet {
				return &ValidationError{Index: idx, Entry: entry, Err: fmt.Errorf("raise below current bet")}
			}
			delta := amount - currentBet
			if delta < lastRaiseDelta {
				return &ValidationError{Index: idx, Entry: entry, Err: fmt.Errorf("raise too small")}
			}
			need := amount - playerBets[pid]
			if s, ok := stacks[pid]; ok && need > s {
				return &ValidationError{Index: idx, Entry: entry, Err: fmt.Errorf("insufficient chips to raise")}
			}
			if s, ok := stacks[pid]; ok {
				stacks[pid] = s - need
			}
			playerBets[pid] = amount
			currentBet = amount
			lastRaiseDelta = delta

		case models.ActionCheck:
			if currentBet > playerBets[pid] {
				// call
				if amount != currentBet {
					return &ValidationError{Index: idx, Entry: entry, Err: fmt.Errorf("call amount mismatch")}
				}
				need := currentBet - playerBets[pid]
				if s, ok := stacks[pid]; ok && need > s {
					return &ValidationError{Index: idx, Entry: entry, Err: fmt.Errorf("insufficient chips to call; must all-in")}
				}
				if s, ok := stacks[pid]; ok {
					stacks[pid] = s - need
				}
				playerBets[pid] = currentBet
			} else {
				if amount != 0 {
					return &ValidationError{Index: idx, Entry: entry, Err: fmt.Errorf("check amount must be 0")}
				}
			}

		case models.ActionAllIn:
			need := amount - playerBets[pid]
			if need < 0 {
				need = 0
			}
			if s, ok := stacks[pid]; ok {
				if need != s {
					return &ValidationError{Index: idx, Entry: entry, Err: fmt.Errorf("all-in amount must equal remaining stack")}
				}
				stacks[pid] = 0
			}
			playerBets[pid] += need
			if amount > currentBet {
				delta := amount - currentBet
				if delta < lastRaiseDelta {
					return &ValidationError{Index: idx, Entry: entry, Err: fmt.Errorf("all-in raise too small")}
				}
				currentBet = amount
				lastRaiseDelta = delta
			}

		case models.ActionFold:
			if amount != 0 {
				return &ValidationError{Index: idx, Entry: entry, Err: fmt.Errorf("fold amount must be 0")}
			}
			playerBets[pid] = 0

		case models.ActionJoin, models.ActionQuit, models.ActionSeat:
			// joining, quitting and seat selections do not impact validation

		default:
			return &ValidationError{Index: idx, Entry: entry, Err: fmt.Errorf("unknown action %s", code)}
		}
	}
	return nil
}
