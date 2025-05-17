package winrate

import (
	"pokerDB/pkg/rules/evaluation"
)

// cardToEval converts a card represented as 1-52 (suit first) into the card
// representation used by the evaluation package which is ordered by rank.
func cardToEval(c int) evaluation.Card {
	r := uint8((c - 1) % 13) // our numbering uses A=0
	var rank uint8
	if r == 0 {
		rank = 12
	} else {
		rank = r - 1
	}
	suit := uint8((c - 1) / 13)
	return evaluation.NewCardFromId((rank << 2) | suit)
}

// Calculate takes player hands and community cards (board) and returns the
// winning probability for each player given the known cards.
// Each hand should contain exactly two cards represented as integers 1-52.
// The board may contain 0 to 5 cards, also encoded as 1-52.
func Calculate(hands [][]int, board []int) []float64 {
	// prepare deck of remaining cards
	used := make(map[int]bool)
	for _, h := range hands {
		for _, c := range h {
			used[c] = true
		}
	}
	for _, c := range board {
		used[c] = true
	}
	remaining := make([]int, 0, 52-len(used))
	for i := 1; i <= 52; i++ {
		if !used[i] {
			remaining = append(remaining, i)
		}
	}

	need := 5 - len(board)
	wins := make([]float64, len(hands))
	var total int

	// recursive enumeration of remaining board cards
	var choose func(start int, picked []int)
	choose = func(start int, picked []int) {
		if len(picked) == need {
			fullBoard := append(board[:], picked...)
			// evaluate hands
			bestVal := uint16(65535)
			winners := []int{}
			for i, h := range hands {
				all := make([]evaluation.Card, 0, 7)
				all = append(all, cardToEval(h[0]), cardToEval(h[1]))
				for _, bc := range fullBoard {
					all = append(all, cardToEval(bc))
				}
				r := evaluation.EvaluateCards(all...)
				v := r.GetValue()
				if v < bestVal {
					bestVal = v
					winners = []int{i}
				} else if v == bestVal {
					winners = append(winners, i)
				}
			}
			split := 1.0 / float64(len(winners))
			for _, w := range winners {
				wins[w] += split
			}
			total++
			return
		}
		for i := start; i <= len(remaining)-(need-len(picked)); i++ {
			picked = append(picked, remaining[i])
			choose(i+1, picked)
			picked = picked[:len(picked)-1]
		}
	}

	if need == 0 {
		choose(0, nil)
	} else {
		choose(0, []int{})
	}

	results := make([]float64, len(hands))
	if total == 0 {
		return results
	}
	for i, w := range wins {
		results[i] = w / float64(total)
	}
	return results
}

func Preflop(hands [][]int) []float64 {
	return Calculate(hands, nil)
}

func Flop(hands [][]int, flop []int) []float64 {
	return Calculate(hands, flop)
}

func Turn(hands [][]int, board []int) []float64 {
	return Calculate(hands, board)
}

func River(hands [][]int, board []int) []float64 {
	return Calculate(hands, board)
}
