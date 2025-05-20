package models

import (
	"github.com/google/uuid"
	"strings"
	"testing"
)

func TestActionStringsReadable(t *testing.T) {
	g := NewGame(uuid.New(), 2)
	g.SmallBlind = 50
	g.BigBlind = 100
	p1 := uuid.New()
	p2 := uuid.New()
	g.MinBuyIn = 100
	g.MaxBuyIn = 1000
	if err := g.BuyIn(p1, 200); err != nil {
		t.Fatalf("buyin1: %v", err)
	}
	if err := g.BuyIn(p2, 200); err != nil {
		t.Fatalf("buyin2: %v", err)
	}
	if err := g.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	g.AddAction(p1, ActionCheck, 0)
	g.AddAction(p2, ActionRaise, 200)
	if err := g.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	lines := g.ActionStrings()
	if len(lines) < 5 {
		t.Fatalf("expected at least 5 lines got %d", len(lines))
	}
	if !containsWord(lines[3], "check") {
		t.Errorf("expected check action, got %s", lines[3])
	}
	if !containsWord(lines[4], "raise") {
		t.Errorf("expected raise action, got %s", lines[4])
	}
}

func containsWord(s, word string) bool {
	return strings.Contains(strings.ToLower(s), word)
}
