package model

import (
	"math"

	"github.com/harbdog/raycaster-go/geom"
)

func PointInProximity(distance, srcX, srcY, tgtX, tgtY float64) bool {
	distance = math.Ceil(distance)
	return (srcX-distance <= tgtX && tgtX <= srcX+distance &&
		srcY-distance <= tgtY && tgtY <= srcY+distance)
}

func RandFloat64In(lo, hi float64) float64 {
	return lo + (hi-lo)*Randish.Float64()
}

// ClampAngle clamps the given angle in a range of -pi to pi
func ClampAngle(angle float64) float64 {
	for angle > geom.Pi {
		angle = angle - geom.Pi2
	}
	for angle <= -geom.Pi {
		angle = angle + geom.Pi2
	}
	return angle
}

// ClampAngle2Pi clamps the given in a range of 0 to 2*pi
func ClampAngle2Pi(angle float64) float64 {
	for angle >= geom.Pi2 {
		angle = angle - geom.Pi2
	}
	for angle < 0 {
		angle = angle + geom.Pi2
	}
	return angle
}

func AngleDistance(a, b float64) float64 {
	// sourced from https://stackoverflow.com/a/75587465/854696
	return math.Mod(math.Mod(b-a+geom.Pi, geom.Pi2)-geom.Pi2, geom.Pi2) + geom.Pi
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
