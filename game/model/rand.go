package model

import (
	"math/rand"

	"github.com/harbdog/raycaster-go/geom"
)

type Rand struct {
	*rand.Rand
}

func NewRNG() *Rand {
	return &Rand{Rand: rand.New(rand.NewSource(rand.Int63()))}
}

func (rng *Rand) RandRelativeLocation(x, y, minDist, maxDist, xMax, yMax int) (int, int) {
	randAngle := rng.RandFloat64In(0, geom.Pi2)
	randDist := rng.RandFloat64In(float64(minDist), float64(maxDist))
	randLine := geom.LineFromAngle(float64(x), float64(y), randAngle, randDist)
	return geom.ClampInt(int(randLine.X2), 0, xMax), geom.ClampInt(int(randLine.Y2), 0, yMax)
}

func (rng *Rand) RandFloat64In(lo, hi float64) float64 {
	return RandFloat64In(lo, hi, rng.Rand)
}

func RandFloat64In(lo, hi float64, rng *rand.Rand) float64 {
	var randFloat float64
	if rng == nil {
		randFloat = rand.Float64()
	} else {
		randFloat = rng.Float64()
	}
	return lo + (hi-lo)*randFloat
}
