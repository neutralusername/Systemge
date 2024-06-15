package Utilities

import "math"

// result is 1 if A wins, 0.5 if draw, 0 if B wins.
// kValue is the weight of the match -> higher kValue means greater impact on the rating.
func CalcElo(ratingA float64, ratingB float64, kValue int, result float64) (float64, float64) {
	expectedA := 1 / (1 + math.Pow(10, (ratingB-ratingA)/400))
	expectedB := 1 / (1 + math.Pow(10, (ratingA-ratingB)/400))
	newRatingA := ratingA + float64(kValue)*(result-expectedA)
	newRatingB := ratingB + float64(kValue)*(1-result-expectedB)
	return newRatingA, newRatingB
}
