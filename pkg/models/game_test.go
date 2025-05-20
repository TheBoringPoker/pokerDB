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
	g.MinBuyIn = 100
	g.MaxBuyIn = 1000
	p1 := uuid.New()
	p2 := uuid.New()
	if err := g.BuyIn(p1, 200); err != nil {
		t.Fatalf("buyin1: %v", err)
	}
	if err := g.BuyIn(p2, 200); err != nil {
		t.Fatalf("buyin2: %v", err)
	}
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

func TestGameBuyInLogic(t *testing.T) {
	g := NewGame(uuid.New(), 2)
	g.MinBuyIn = 100
	g.MaxBuyIn = 1000

	if err := g.Start(); err == nil {
		t.Fatal("start should fail without buy-ins")
	}

	p1 := uuid.New()
	p2 := uuid.New()
	if err := g.BuyIn(p1, 50); err == nil {
		t.Fatal("expected buy-in below min to fail")
	}
	if err := g.BuyIn(p1, 200); err != nil {
		t.Fatalf("buy-in p1: %v", err)
	}
	if err := g.BuyIn(p2, 300); err != nil {
		t.Fatalf("buy-in p2: %v", err)
	}
	if err := g.Start(); err != nil {
		t.Fatalf("start with buy-ins: %v", err)
	}
	if err := g.BuyIn(uuid.New(), 200); err == nil {
		t.Fatal("expected buy-in after start to fail")
	}
}
