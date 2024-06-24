package model

import (
	"math"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
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

// ConvergencePoint returns the convergence point from current unit angle/pitch to unit target
// Returns nil if unit does not have a target.
func ConvergencePoint(u Unit, t Entity) *geom3d.Vector3 {
	if t == nil {
		return nil
	}

	uX, uY, uZ := u.Pos().X, u.Pos().Y, u.PosZ()+u.CockpitOffset().Y
	tX, tY, tZ := t.Pos().X, t.Pos().Y, t.PosZ()
	targetDist := (&geom3d.Line3d{
		X1: uX, Y1: uY, Z1: uZ,
		X2: tX, Y2: tY, Z2: tZ,
	}).Distance()

	convergenceLine := geom3d.Line3dFromAngle(uX, uY, uZ, u.TurretAngle(), u.Pitch(), targetDist)
	convergencePoint := &geom3d.Vector3{X: convergenceLine.X2, Y: convergenceLine.Y2, Z: convergenceLine.Z2}

	return convergencePoint
}
