package model

import "math/rand"

func RandFloat64In(lo, hi float64) float64 {
	return lo + (hi-lo)*rand.Float64()
}
