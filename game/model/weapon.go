package model

import (
	"math"

	"github.com/harbdog/raycaster-go/geom"
)

type WeaponType int

const (
	ENERGY WeaponType = iota
	BALLISTIC
	MISSILE
)

type Weapon interface {
	Name() string
	ShortName() string
	Tech() TechBase
	Type() WeaponType
	Tonnage() float64
	Damage() float64
	Heat() float64
	Distance() float64
	Velocity() float64
	Cooldown() float64
	MaxCooldown() float64
	DecreaseCooldown(float64)
	TriggerCooldown()

	Offset() *geom.Vector2
	ProjectileCount() int
	ProjectileDelay() float64
	SpawnProjectile(angle, pitch float64, spawnedBy Entity) *Projectile

	Clone() Weapon
	Parent() Entity
}

// WeaponPosition3D gets the X, Y and Z axis offsets needed for weapon projectile spawned from a 2-D sprite reference
func WeaponPosition3D(e Entity, weaponOffX, weaponOffY float64) (float64, float64, float64) {
	unitPosition := e.Pos()
	wX, wY, wZ := unitPosition.X, unitPosition.Y, e.PosZ()+weaponOffY

	if weaponOffX == 0 {
		// no X/Y position adjustments needed
		return wX, wY, wZ
	}

	// calculate X,Y based on player orientation angle perpendicular to angle of view
	offAngle := e.Angle() + math.Pi/2

	// create line segment using offset angle and X offset to determine 3D position offset of X/Y
	offLine := geom.LineFromAngle(0, 0, offAngle, weaponOffX)
	wX, wY = wX+offLine.X2, wY+offLine.Y2

	return wX, wY, wZ
}
