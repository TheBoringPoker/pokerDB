package winrate

import (
	"math"
	"testing"
)

func nearlyEqual(a, b, eps float64) bool {
	return math.Abs(a-b) < eps
}

func TestPreflopAAvsKK(t *testing.T) {
	// A♠ A♥ vs K♠ K♥
	hands := [][]int{{1, 14}, {13, 26}}
	rates := Preflop(hands)
	if len(rates) != 2 {
		t.Fatalf("expected 2 results, got %d", len(rates))
	}
	if !(nearlyEqual(rates[0], 0.82, 0.02) && nearlyEqual(rates[1], 0.18, 0.02)) {
		t.Errorf("unexpected win rates: %v", rates)
	}
}

func TestPreflopAKvsQQ(t *testing.T) {
	// A♠ K♠ vs Q♥ Q♦
	hands := [][]int{{1, 13}, {25, 38}}
	rates := Preflop(hands)
	if len(rates) != 2 {
		t.Fatalf("expected 2 results, got %d", len(rates))
	}
	if !(nearlyEqual(rates[0], 0.46, 0.02) && nearlyEqual(rates[1], 0.54, 0.02)) {
		t.Errorf("unexpected win rates: %v", rates)
	}
}

func TestFlopAAvsKK(t *testing.T) {
	hands := [][]int{{1, 14}, {13, 26}}
	flop := []int{41, 33, 22} // 2♣ 7♦ 9♥
	rates := Flop(hands, flop)
	if len(rates) != 2 {
		t.Fatalf("expected 2 results, got %d", len(rates))
	}
	if !(nearlyEqual(rates[0], 0.91, 0.02) && nearlyEqual(rates[1], 0.09, 0.02)) {
		t.Errorf("unexpected win rates: %v", rates)
	}
}
