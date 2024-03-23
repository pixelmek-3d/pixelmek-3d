package model

import (
	"fmt"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
)

type WeaponType int

const (
	ENERGY WeaponType = iota
	BALLISTIC
	MISSILE
)

type WeaponClassification int

const (
	WEAPON_CLASS_UNDEFINED WeaponClassification = iota
	ENERGY_LASER
	ENERGY_PPC
	ENERGY_FLAMER
	BALLISTIC_MACHINEGUN
	BALLISTIC_AUTOCANNON
	BALLISTIC_LBX_AC
	BALLISTIC_GAUSS
	MISSILE_LRM
	MISSILE_SRM
)

type WeaponFireMode int

const (
	CHAIN_FIRE WeaponFireMode = iota
	GROUP_FIRE
)

const (
	weaponSummaryPadding = 12
)

type Weapon interface {
	File() string
	Name() string
	ShortName() string
	Summary() string
	Tech() TechBase
	Type() WeaponType
	Classification() WeaponClassification
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
	AmmoPerTon() int
	AmmoBin() *AmmoBin
	SetAmmoBin(ammoBin *AmmoBin)
	ProjectileCount() int
	ProjectileDelay() float64
	ProjectileSpread() float64
	SpawnProjectile(angle, pitch float64, spawnedBy Unit) *Projectile
	SpawnProjectileToward(convergencePoint *geom3d.Vector3, spawnedBy Unit) *Projectile

	Audio() string
	Clone() Weapon
	Parent() Entity
}

func weaponSummary(w Weapon) string {
	s := ""
	pad := weaponSummaryPadding

	pCount := w.ProjectileCount()
	pDamage := w.Damage()
	if pCount > 1 {
		pDamage /= float64(pCount)
		s += fmt.Sprintf("%-*s%0.1fx%d\n", pad, "Damage:", pDamage, pCount)
	} else {
		s += fmt.Sprintf("%-*s%0.1f\n", pad, "Damage:", pDamage)
	}

	s += fmt.Sprintf("%-*s%0.1f\n", pad, "Heat:", w.Heat())
	s += fmt.Sprintf("%-*s%0.0fm\n", pad, "Range:", w.Distance())
	s += fmt.Sprintf("%-*s%0.1fs\n", pad, "Cooldown:", w.MaxCooldown())
	return s
}

// WeaponPosition3D gets the X, Y and Z axis offsets needed for weapon projectile spawned from a 2-D sprite reference
func WeaponPosition3D(e Unit, weaponOffX, weaponOffY float64) *geom3d.Vector3 {
	unitPosition := e.Pos()
	wX, wY, wZ := unitPosition.X, unitPosition.Y, e.PosZ()+weaponOffY

	if weaponOffX == 0 {
		// no X/Y position adjustments needed
		return &geom3d.Vector3{X: wX, Y: wY, Z: wZ}
	}

	// calculate X,Y based on player orientation angle perpendicular to angle of view
	offAngle := e.Heading() + e.TurretAngle() + geom.Pi/2

	// create line segment using offset angle and X offset to determine 3D position offset of X/Y
	offLine := geom.LineFromAngle(0, 0, offAngle, weaponOffX)
	wX, wY = wX+offLine.X2, wY+offLine.Y2

	return &geom3d.Vector3{X: wX, Y: wY, Z: wZ}
}

func HeadingPitchTowardPoint3D(source, target *geom3d.Vector3) (float64, float64) {
	var heading, pitch float64
	convergenceLine3d := &geom3d.Line3d{
		X1: source.X, Y1: source.Y, Z1: source.Z,
		X2: target.X, Y2: target.Y, Z2: target.Z,
	}
	heading, pitch = convergenceLine3d.Heading(), convergenceLine3d.Pitch()
	return heading, pitch
}

func GetGroupsForWeapon(w Weapon, weaponGroups [][]Weapon) []uint {
	groupsForWeapon := make([]uint, 0, len(weaponGroups))
	if w == nil {
		return groupsForWeapon
	}

	for g, weapons := range weaponGroups {
		for _, gWeapon := range weapons {
			if w == gWeapon {
				groupsForWeapon = append(groupsForWeapon, uint(g))
			}
		}
	}

	return groupsForWeapon
}

func IsWeaponInGroup(w Weapon, g uint, weaponGroups [][]Weapon) bool {
	if w == nil || int(g) >= len(weaponGroups) {
		return false
	}

	for _, gWeapon := range weaponGroups[g] {
		if w == gWeapon {
			return true
		}
	}

	return false
}
