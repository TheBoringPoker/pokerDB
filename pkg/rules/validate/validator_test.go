package validate

import (
	"testing"

	"github.com/google/uuid"
	"pokerDB/pkg/models"
)

func TestValidateGameValid(t *testing.T) {
	g := models.NewGame(uuid.New(), 2)
	g.SmallBlind = 50
	g.BigBlind = 100
	g.MinBuyIn = 100
	g.MaxBuyIn = 1000
	p1 := uuid.New()
	p2 := uuid.New()
	if err := g.BuyIn(p1, 500); err != nil {
		t.Fatalf("buyin p1: %v", err)
	}
	if err := g.BuyIn(p2, 500); err != nil {
		t.Fatalf("buyin p2: %v", err)
	}
	if err := g.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := g.AddAction(p1, models.ActionCheck, 100); err != nil {
		t.Fatalf("action1: %v", err)
	}
	if err := g.AddAction(p2, models.ActionRaise, 300); err != nil {
		t.Fatalf("action2: %v", err)
	}
	if err := g.AddAction(p1, models.ActionFold, 0); err != nil {
		t.Fatalf("action3: %v", err)
	}
	if err := g.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	stacks := map[string]int64{
		p1.String()[:8]: 1000,
		p2.String()[:8]: 1000,
	}
	if err := Validate(g, stacks); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateGameInvalidRaise(t *testing.T) {
	g := models.NewGame(uuid.New(), 2)
	g.SmallBlind = 50
	g.BigBlind = 100
	g.MinBuyIn = 100
	g.MaxBuyIn = 1000
	p1 := uuid.New()
	p2 := uuid.New()
	if err := g.BuyIn(p1, 500); err != nil {
		t.Fatalf("buyin p1: %v", err)
	}
	if err := g.BuyIn(p2, 500); err != nil {
		t.Fatalf("buyin p2: %v", err)
	}
	if err := g.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := g.AddAction(p1, models.ActionRaise, 200); err != nil {
		t.Fatalf("action1: %v", err)
	}
	if err := g.AddAction(p2, models.ActionRaise, 250); err != nil {
		t.Fatalf("action2: %v", err)
	}
	if err := g.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	if err := Validate(g, nil); err == nil {
		t.Fatalf("expected error")
	}
}

func TestValidateGameInsufficientCall(t *testing.T) {
	g := models.NewGame(uuid.New(), 2)
	g.SmallBlind = 50
	g.BigBlind = 100
	g.MinBuyIn = 100
	g.MaxBuyIn = 1000
	p1 := uuid.New()
	p2 := uuid.New()
	if err := g.BuyIn(p1, 500); err != nil {
		t.Fatalf("buyin p1: %v", err)
	}
	if err := g.BuyIn(p2, 500); err != nil {
		t.Fatalf("buyin p2: %v", err)
	}
	if err := g.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := g.AddAction(p1, models.ActionCheck, 100); err != nil {
		t.Fatalf("action1: %v", err)
	}
	if err := g.AddAction(p2, models.ActionRaise, 300); err != nil {
		t.Fatalf("action2: %v", err)
	}
	if err := g.AddAction(p1, models.ActionCheck, 300); err != nil {
		t.Fatalf("action3: %v", err)
	}
	if err := g.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	stacks := map[string]int64{
		p1.String()[:8]: 200,
		p2.String()[:8]: 1000,
	}
	if err := Validate(g, stacks); err == nil {
		t.Fatalf("expected error")
	}
}
