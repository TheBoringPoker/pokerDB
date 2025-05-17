package utils

var suits = []string{"♠", "♥", "♦", "♣"}
var ranks = []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}

// CardToString converts a card value from 1-52 into a human readable
// representation like "A♠" or "10♥". If the value is outside this range,
// "??" is returned.
func CardToString(card int) string {
	if card < 1 || card > 52 {
		return "??"
	}
	suit := suits[(card-1)/13]
	rank := ranks[(card-1)%13]
	return rank + suit
}
