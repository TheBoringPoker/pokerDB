package models

import (
	"github.com/google/uuid"
	"testing"
)

func TestGameStartEndErrors(t *testing.T) {
	g := NewGame(uuid.New(), 1)
	if err := g.Start(); err == nil {
		t.Fatal("expected error with too few players")
	}

	g = NewGame(uuid.New(), 2)
	if err := g.Start(); err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}
	if err := g.Start(); err == nil {
		t.Fatal("expected error on second start")
	}
	if err := g.End(); err != nil {
		t.Fatalf("unexpected end error: %v", err)
	}
	if err := g.End(); err == nil {
		t.Fatal("expected error on second end")
	}
}
