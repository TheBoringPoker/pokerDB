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
	if err := g.BuyIn(uuid.New(), 200); err != nil {
		t.Fatalf("buy-in mid-round: %v", err)
	}
	if err := g.EndRound(); err != nil {
		t.Fatalf("end round: %v", err)
	}
	if err := g.BuyIn(uuid.New(), 200); err != nil {
		t.Fatalf("buy-in between rounds: %v", err)
	}
}

func TestAddActionInsufficientChips(t *testing.T) {
	g := NewGame(uuid.New(), 2)
	g.MinBuyIn = 100
	g.MaxBuyIn = 1000
	p1 := uuid.New()
	p2 := uuid.New()
	if err := g.BuyIn(p1, 100); err != nil {
		t.Fatalf("buyin1: %v", err)
	}
	if err := g.BuyIn(p2, 100); err != nil {
		t.Fatalf("buyin2: %v", err)
	}
	if err := g.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := g.AddAction(p1, ActionRaise, 200); err == nil {
		t.Fatal("expected error on insufficient chips")
	}
}

func TestJoinQuitSeat(t *testing.T) {
	g := NewGame(uuid.New(), 0)
	p1 := uuid.New()
	if err := g.Join(p1); err != nil {
		t.Fatalf("join: %v", err)
	}
	if err := g.ChooseSeat(p1, 3); err != nil {
		t.Fatalf("seat: %v", err)
	}
	if g.NextSeats[p1] != 3 {
		t.Fatalf("expected seat 3 got %d", g.NextSeats[p1])
	}
	if err := g.Quit(p1); err != nil {
		t.Fatalf("quit: %v", err)
	}
	if !g.Ended() {
		t.Fatalf("game should auto end when last player quits")
	}
	if len(g.ActionLog) == 0 {
		t.Fatalf("expected actions recorded")
	}
}

func TestSeatValidation(t *testing.T) {
	g := NewGame(uuid.New(), 0)
	p1 := uuid.New()
	p2 := uuid.New()
	if err := g.Join(p1); err != nil {
		t.Fatalf("join1: %v", err)
	}
	if err := g.Join(p2); err != nil {
		t.Fatalf("join2: %v", err)
	}
	if err := g.ChooseSeat(p1, 10); err == nil {
		t.Fatal("expected seat out of range")
	}
	if err := g.ChooseSeat(p1, 2); err != nil {
		t.Fatalf("seat1: %v", err)
	}
	if err := g.ChooseSeat(p2, 2); err == nil {
		t.Fatal("expected seat conflict")
	}
}
