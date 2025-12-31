package model

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
)

type Rect struct {
	X1, Y1, X2, Y2 float64
}

// NewRect returns a new rectangle with the given coordinates. The resulting
// rectangle has minimum and maximum coordinates swapped if necessary so that
// it is well-formed.
func NewRect(x1, y1, x2, y2 float64) Rect {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	return Rect{X1: x1, Y1: y1, X2: x2, Y2: y2}
}

func (r Rect) ContainsPoint(x, y float64) bool {
	return r.X1 <= x && x <= r.X2 && r.Y1 <= y && y <= r.Y2
}

func (r Rect) Dx() float64 {
	return r.X2 - r.X1
}

func (r Rect) Dy() float64 {
	return r.Y2 - r.Y1
}

func LineOpposite(l geom.Line) geom.Line {
	return geom.Line{X1: l.X2, Y1: l.Y2, X2: l.X1, Y2: l.Y1}
}

// PointInProximity is a fast but inaccurate distance check between two positions
func PointInProximity(distance, srcX, srcY, tgtX, tgtY float64) bool {
	distance = math.Ceil(distance)
	return math.Abs(tgtX-srcX) <= distance && math.Abs(tgtY-srcY) <= distance
}

// PointInDistance is a distance check between two positions
func PointInDistance(distance, srcX, srcY, tgtX, tgtY float64) bool {
	line := geom.Line{
		X1: srcX, Y1: srcY,
		X2: tgtX, Y2: tgtY,
	}
	return line.Distance() <= distance
}

// PointInLine returns true if a point is on a line segment
func PointInLine(p geom.Vector2, l geom.Line, epsilon float64) bool {
	// based on https://stackoverflow.com/a/328122/854696
	crossProduct := (p.Y-l.Y1)*(l.X2-l.X1) - (p.X-l.X1)*(l.Y2-l.Y1)
	if math.Abs(crossProduct) > epsilon {
		return false
	}

	dotProduct := (p.X-l.X1)*(l.X2-l.X1) + (p.Y-l.Y1)*(l.Y2-l.Y1)
	if dotProduct < 0 {
		return false
	}

	squaredLengthBA := (l.X2-l.X1)*(l.X2-l.X1) + (l.Y2-l.Y1)*(l.Y2-l.Y1)
	if dotProduct > squaredLengthBA {
		return false
	}
	return true
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

// AngleToCardinal converts model radian angle based on X/Y plane
// to cardinal compass in degrees (where North is 0 and goes clockwise)
func AngleToCardinal(angle float64) float64 {
	return geom.Degrees(ClampAngle2Pi(geom.HalfPi - angle))
}

// CardinalToAngle converts cardinal compass in degrees
// to model radian angle based on X/Y plane (where East is 0 and goes counter-clockwise)
func CardinalToAngle(compassDegrees float64) float64 {
	return ClampAngle2Pi(geom.Radians(-compassDegrees) + geom.HalfPi)
}

func Hypotenuse(a, b float64) float64 {
	return math.Sqrt(a*a + b*b)
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

	// determine approximate lead distance needed for weapon projectile
	if w != nil {
		// approximate position of target based on its current heading and speed for projectile flight time
		tDist := tLine.Distance()
		tProjectile := w.Projectile()
		tDelta := tDist / tProjectile.MaxVelocity()
		tLine = geom3d.Line3dFromAngle(t.Pos().X, t.Pos().Y, t.PosZ()+zTargetOffset, t.Heading(), 0, tDelta*t.Velocity())
	}

	return &geom3d.Vector3{X: tLine.X2, Y: tLine.Y2, Z: tLine.Z2}
}
