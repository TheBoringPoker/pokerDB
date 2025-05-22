package models

import (
	"sync"
	"testing"

	"github.com/google/uuid"
)

// TestGameConcurrentInvalidOps exercises Game methods concurrently.
// It runs multiple goroutines performing invalid and valid operations to
// verify thread safety and proper error handling.
func TestGameConcurrentInvalidOps(t *testing.T) {
	g := NewGame(uuid.New(), 3)
	g.MinBuyIn = 100
	g.MaxBuyIn = 1000

	p1 := uuid.New()
	p2 := uuid.New()
	p3 := uuid.New()
	missing := uuid.New()

	var wg sync.WaitGroup

	// Run a batch of invalid operations concurrently before the game starts.
	invalidFuncs := []func(){
		func() {
			if err := g.AddAction(p1, ActionRaise, 100); err == nil {
				t.Errorf("AddAction before start should fail")
			}
		},
		func() {
			if err := g.ChooseSeat(p1, 10); err == nil {
				t.Errorf("ChooseSeat with invalid seat should fail")
			}
		},
		func() {
			if err := g.Quit(missing); err == nil {
				t.Errorf("Quit unknown player should fail")
			}
		},
		func() {
			if err := g.End(); err == nil {
				t.Errorf("End before start should fail")
			}
		},
		func() {
			if err := g.BuyIn(p1, 50); err == nil {
				t.Errorf("BuyIn below minimum should fail")
			}
		},
		func() {
			if err := g.Start(); err == nil {
				t.Errorf("Start without buy-ins should fail")
			}
		},
	}

	for _, f := range invalidFuncs {
		wg.Add(1)
		go func(fn func()) {
			defer wg.Done()
			fn()
		}(f)
	}
	wg.Wait()

	// Set up a valid game sequentially.
	if err := g.BuyIn(p1, 200); err != nil {
		t.Fatalf("buyin p1: %v", err)
	}
	if err := g.BuyIn(p2, 200); err != nil {
		t.Fatalf("buyin p2: %v", err)
	}
	if err := g.BuyIn(p3, 200); err != nil {
		t.Fatalf("buyin p3: %v", err)
	}
	if err := g.Start(); err != nil {
		t.Fatalf("start game: %v", err)
	}

	// Concurrent actions while the game is running.
	actionFuncs := []struct {
		pid       uuid.UUID
		code      string
		amount    int64
		expectErr bool
	}{
		{p1, ActionCheck, 0, false},
		{p2, ActionRaise, 200, false},
		{p3, ActionFold, 0, false},
		{missing, ActionRaise, 100, true},
	}

	wg.Add(len(actionFuncs))
	for _, a := range actionFuncs {
		go func(ac struct {
			pid       uuid.UUID
			code      string
			amount    int64
			expectErr bool
		}) {
			defer wg.Done()
			err := g.AddAction(ac.pid, ac.code, ac.amount)
			if ac.expectErr && err == nil {
				t.Errorf("expected error for player %s", ac.pid)
			}
			if !ac.expectErr && err != nil {
				t.Errorf("unexpected error for player %s: %v", ac.pid, err)
			}
		}(a)
	}
	wg.Wait()

	if err := g.End(); err != nil {
		t.Fatalf("end game: %v", err)
	}
}
