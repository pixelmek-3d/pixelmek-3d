package model

import (
	"math"
	"math/rand"

	"github.com/harbdog/raycaster-go/geom"
)

func PointInProximity(distance, srcX, srcY, tgtX, tgtY float64) bool {
	distance = math.Ceil(distance)
	return (srcX-distance <= tgtX && tgtX <= srcX+distance &&
		srcY-distance <= tgtY && tgtY <= srcY+distance)
}

func RandFloat64In(lo, hi float64) float64 {
	return lo + (hi-lo)*rand.Float64()
}

func IsBetweenDegrees(start, end, mid float64) bool {
	if end-start < 0.0 {
		end = end - start + 360.0
	} else {
		end = end - start
	}

	if (mid - start) < 0.0 {
		mid = mid - start + 360.0
	} else {
		mid = mid - start
	}

	return mid < end
}

func IsBetweenRadians(start, end, mid float64) bool {
	if end-start < 0.0 {
		end = end - start + geom.Pi2
	} else {
		end = end - start
	}

	if (mid - start) < 0.0 {
		mid = mid - start + geom.Pi2
	} else {
		mid = mid - start
	}

	return mid < end
}
