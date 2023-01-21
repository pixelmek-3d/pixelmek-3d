package model

import (
	"math"
	"math/rand"
)

func PointInProximity(distance, srcX, srcY, tgtX, tgtY float64) bool {
	distance = math.Ceil(distance)
	return (srcX-distance <= tgtX && tgtX <= srcX+distance &&
		srcY-distance <= tgtY && tgtY <= srcY+distance)
}

func RandFloat64In(lo, hi float64) float64 {
	return lo + (hi-lo)*rand.Float64()
}
