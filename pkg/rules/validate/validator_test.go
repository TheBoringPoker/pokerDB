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
	if err := g.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	p1 := uuid.New()
	p2 := uuid.New()
	g.AddAction(p1, models.ActionCheck, 0)
	g.AddAction(p2, models.ActionRaise, 300)
	g.AddAction(p1, models.ActionFold, 0)
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
	if err := g.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	p1 := uuid.New()
	p2 := uuid.New()
	g.AddAction(p1, models.ActionRaise, 200)
	g.AddAction(p2, models.ActionRaise, 250)
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
	if err := g.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	p1 := uuid.New()
	p2 := uuid.New()
	g.AddAction(p1, models.ActionCheck, 0)
	g.AddAction(p2, models.ActionRaise, 300)
	g.AddAction(p1, models.ActionCheck, 300)
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
