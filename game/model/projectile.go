package model

import (
	"math/rand"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/jinzhu/copier"
)

type Projectile struct {
	Resource        *ModelProjectileResource
	position        *geom.Vector2
	positionZ       float64
	anchor          raycaster.SpriteAnchor
	angle           float64
	pitch           float64
	velocity        float64
	velocityZ       float64
	maxVelocity     float64
	acceleration    float64
	collisionRadius float64
	collisionHeight float64
	lifespan        float64
	extremeLifespan float64
	inExtremeRange  bool
	damage          float64
	weapon          Weapon
	lockOnOffset    *geom3d.Vector3
	team            int
	rng             *rand.Rand
	parent          Entity
}

const projectileHitPointsIgnored float64 = 0.12345

func NewProjectile(
	r *ModelProjectileResource, damage, velocity, optimalLifespan, extremeLifespan,
	collisionRadius, collisionHeight float64,
) *Projectile {
	p := &Projectile{
		Resource:        r,
		anchor:          raycaster.AnchorCenter,
		damage:          damage,
		velocity:        velocity,
		maxVelocity:     velocity,
		acceleration:    velocity,
		lifespan:        optimalLifespan,
		extremeLifespan: extremeLifespan,
		inExtremeRange:  false,
		collisionRadius: collisionRadius,
		collisionHeight: collisionHeight,
	}
	return p
}

func (e *Projectile) LockOnOffset() *geom3d.Vector3 {
	if e.rng == nil {
		// using its own rand generator to avoid mutex lock in async updates
		e.rng = NewRNG()
	}
	if e.lockOnOffset == nil {
		missileWeapon, isMissile := e.weapon.(*MissileWeapon)
		if isMissile && missileWeapon.IsLockOn() {
			groupRadius := missileWeapon.LockOnGroupRadius()
			randRadius := RandFloat64In(-groupRadius, groupRadius, e.rng)
			randHeading := RandFloat64In(-geom.Pi, geom.Pi, e.rng)
			randPitch := RandFloat64In(-geom.Pi, geom.Pi, e.rng)

			randLine := geom3d.Line3dFromAngle(0, 0, 0, randHeading, randPitch, randRadius)
			e.lockOnOffset = &geom3d.Vector3{X: randLine.X2, Y: randLine.Y2, Z: randLine.Z2}
		} else {
			e.lockOnOffset = &geom3d.Vector3{}
		}
	}
	return e.lockOnOffset
}

func (e *Projectile) InExtremeRange() bool {
	return e.inExtremeRange
}

func (e *Projectile) Clone() Entity {
	eClone := &Projectile{}
	copier.Copy(eClone, e)
	return eClone
}

func (e *Projectile) Team() int {
	return e.team
}

func (e *Projectile) SetTeam(team int) {
	e.team = team
}

func (e *Projectile) Damage() float64 {
	actualDamage := e.damage
	if e.inExtremeRange && e.lifespan >= 0 {
		// linear damage dropoff based on lifespan/extremeLifespan
		actualDamage = (e.lifespan / e.extremeLifespan) * actualDamage

		if actualDamage < 0 {
			actualDamage = 0
		}
	}

	return actualDamage
}

func (e *Projectile) Lifespan() float64 {
	return e.lifespan
}

func (e *Projectile) DecreaseLifespan(decreaseBy float64) float64 {
	if e.lifespan > 0 && decreaseBy > 0 {
		e.lifespan -= decreaseBy

		if e.lifespan <= 0 && e.extremeLifespan > 0 && !e.inExtremeRange {
			// enter extended range mode
			e.lifespan += e.extremeLifespan
			e.inExtremeRange = true
		}

		if e.lifespan < 0 {
			e.lifespan = 0
		}
	}

	return e.lifespan
}

func (e *Projectile) Destroy() {
	e.lifespan = -1
}

func (e *Projectile) Pos() *geom.Vector2 {
	return e.position
}

func (e *Projectile) SetPos(pos *geom.Vector2) {
	e.position = pos
}

func (e *Projectile) PosZ() float64 {
	return e.positionZ
}

func (e *Projectile) SetPosZ(posZ float64) {
	e.positionZ = posZ
}

func (e *Projectile) Anchor() raycaster.SpriteAnchor {
	return e.anchor
}

func (e *Projectile) SetAnchor(anchor raycaster.SpriteAnchor) {
	e.anchor = anchor
}

func (e *Projectile) Heading() float64 {
	return e.angle
}

func (e *Projectile) SetHeading(angle float64) {
	e.angle = angle
}

func (e *Projectile) Pitch() float64 {
	return e.pitch
}

func (e *Projectile) SetPitch(pitch float64) {
	e.pitch = pitch
}

func (e *Projectile) Velocity() float64 {
	return e.velocity
}

func (e *Projectile) SetVelocity(velocity float64) {
	e.velocity = velocity
}

func (e *Projectile) MaxVelocity() float64 {
	return e.maxVelocity
}

func (e *Projectile) SetMaxVelocity(maxVelocity float64) {
	e.maxVelocity = maxVelocity
}

func (e *Projectile) Acceleration() float64 {
	return e.acceleration
}

func (e *Projectile) SetAcceleration(acceleration float64) {
	e.acceleration = acceleration
}

func (e *Projectile) VelocityZ() float64 {
	return e.velocityZ
}

func (e *Projectile) SetVelocityZ(velocityZ float64) {
	e.velocityZ = velocityZ
}

func (e *Projectile) CollisionRadius() float64 {
	return e.collisionRadius
}

func (e *Projectile) SetCollisionRadius(collisionRadius float64) {
	e.collisionRadius = collisionRadius
}

func (e *Projectile) CollisionHeight() float64 {
	return e.collisionHeight
}

func (e *Projectile) SetCollisionHeight(collisionHeight float64) {
	e.collisionHeight = collisionHeight
}

func (e *Projectile) CockpitOffset() *geom.Vector2 {
	return &geom.Vector2{}
}

func (e *Projectile) ApplyDamage(damage float64) {
	// projectileHitPointsIgnored
}

func (e *Projectile) ArmorPoints() float64 {
	return projectileHitPointsIgnored
}

func (e *Projectile) SetArmorPoints(armor float64) {
	// projectileHitPointsIgnored
}

func (e *Projectile) MaxArmorPoints() float64 {
	return projectileHitPointsIgnored
}

func (e *Projectile) StructurePoints() float64 {
	return projectileHitPointsIgnored
}

func (e *Projectile) SetStructurePoints(structure float64) {
	// projectileHitPointsIgnored
}

func (e *Projectile) MaxStructurePoints() float64 {
	return projectileHitPointsIgnored
}

func (e *Projectile) IsDestroyed() bool {
	// projectileHitPointsIgnored
	return false
}

func (e *Projectile) Weapon() Weapon {
	return e.weapon
}

func (e *Projectile) SetWeapon(weapon Weapon) {
	e.weapon = weapon
}

func (e *Projectile) Parent() Entity {
	return e.parent
}

func (e *Projectile) SetParent(parent Entity) {
	e.parent = parent
}
