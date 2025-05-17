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
	if err := g.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	g.AddAction(uuid.New(), ActionCheck, 0)
	g.AddAction(uuid.New(), ActionRaise, 200)
	if err := g.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	lines := g.ActionStrings()
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines got %d", len(lines))
	}
	if !containsWord(lines[1], "check") {
		t.Errorf("expected check action, got %s", lines[1])
	}
	if !containsWord(lines[2], "raise") {
		t.Errorf("expected raise action, got %s", lines[2])
	}
}

func containsWord(s, word string) bool {
	return strings.Contains(strings.ToLower(s), word)
}
