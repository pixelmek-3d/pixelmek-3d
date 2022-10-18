package model

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
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
	collisionRadius float64
	collisionHeight float64
	lifespan        float64
	lifespanDropoff float64
	damage          float64
	parent          Entity
}

const projectileHitPointsIgnored float64 = 0.12345

func NewProjectile(r *ModelProjectileResource, damage, velocity, lifespan, collisionRadius, collisionHeight float64, parent Entity) *Projectile {
	p := &Projectile{
		Resource:        r,
		anchor:          raycaster.AnchorCenter,
		damage:          damage,
		velocity:        velocity,
		lifespan:        lifespan,
		lifespanDropoff: lifespan / 2,
		collisionRadius: collisionRadius,
		collisionHeight: collisionHeight,
		parent:          parent,
	}
	return p
}

func (e *Projectile) Clone() Entity {
	eClone := &Projectile{}
	copier.Copy(eClone, e)
	return eClone
}

func (e *Projectile) Name() string {
	return "projectile"
}

func (e *Projectile) Variant() string {
	return "projectile"
}

func (e *Projectile) AddArmament(Weapon) {}

func (e *Projectile) Armament() []Weapon {
	return nil
}

func (e *Projectile) Damage() float64 {
	actualDamage := e.damage
	if e.lifespan >= 0 && e.lifespan < e.lifespanDropoff {
		// linear damage dropoff based on lifespan/lifespanDropoff
		actualDamage = (e.lifespan / e.lifespanDropoff) * actualDamage

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

func (e *Projectile) Angle() float64 {
	return e.angle
}

func (e *Projectile) SetAngle(angle float64) {
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

func (e *Projectile) Parent() Entity {
	return e.parent
}

func (e *Projectile) SetParent(parent Entity) {
	e.parent = parent
}
