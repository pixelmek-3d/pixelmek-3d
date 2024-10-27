package model

import (
	"math"
	"math/rand"
	"time"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
)

func NewRNG() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
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

func PointInProximity(distance, srcX, srcY, tgtX, tgtY float64) bool {
	distance = math.Ceil(distance)
	return (srcX-distance <= tgtX && tgtX <= srcX+distance &&
		srcY-distance <= tgtY && tgtY <= srcY+distance)
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
	if u == nil || t == nil {
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

// TargetLeadPosition returns the approximate lead position from current unit to target with given weapon
// Return nil if unit does not have a target
func TargetLeadPosition(u, t Unit, w Weapon) *geom3d.Vector3 {
	if u == nil || t == nil {
		return nil
	}

	var zTargetOffset float64
	switch t.Anchor() {
	case raycaster.AnchorBottom:
		zTargetOffset = t.CollisionHeight() / 2
	case raycaster.AnchorTop:
		zTargetOffset = -t.CollisionHeight() / 2
	}

	// calculate distance from unit to target
	tLine := geom3d.Line3d{
		X1: u.Pos().X, Y1: u.Pos().Y, Z1: u.PosZ() + u.CockpitOffset().Y,
		X2: t.Pos().X, Y2: t.Pos().Y, Z2: t.PosZ() + zTargetOffset,
	}
	tDist := tLine.Distance()

	// determine approximate lead distance needed for weapon projectile
	if w != nil {
		// approximate position of target based on its current heading and speed for projectile flight time
		tProjectile := w.Projectile()
		tDelta := tDist / tProjectile.MaxVelocity()
		tLine = geom3d.Line3dFromAngle(t.Pos().X, t.Pos().Y, t.PosZ()+zTargetOffset, t.Heading(), 0, tDelta*t.Velocity())
	}

	return &geom3d.Vector3{X: tLine.X2, Y: tLine.Y2, Z: tLine.Z2}
}
