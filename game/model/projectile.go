package model

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type Projectile struct {
	position        *geom.Vector2
	positionZ       float64
	anchor          raycaster.SpriteAnchor
	angle           float64
	pitch           float64
	velocity        float64
	collisionRadius float64
	collisionHeight float64
	lifespan        float64
	damage          float64
	parent          Entity
}

const projectileHitPointsIgnored float64 = 0.123

func NewProjectile(damage, lifespan, collisionRadius, collisionHeight float64) *Projectile {
	p := &Projectile{
		anchor:          raycaster.AnchorCenter,
		collisionRadius: collisionRadius,
		collisionHeight: collisionHeight,
		lifespan:        lifespan,
		damage:          damage,
	}
	return p
}

func (e *Projectile) Damage() float64 {
	return e.damage
}

func (e *Projectile) Lifespan() float64 {
	return e.lifespan
}

func (e *Projectile) DecreaseLifespan(decreaseBy float64) float64 {
	if decreaseBy > 0 {
		e.lifespan -= decreaseBy
	}
	return e.lifespan
}

func (e *Projectile) ZeroLifespan() {
	e.lifespan = 0
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

func (e *Projectile) HitPoints() float64 {
	// projectiles are only tested for damage against non-projectiles
	return projectileHitPointsIgnored
}

func (e *Projectile) SetHitPoints(hitPointsIgnored float64) {
	// projectiles are only tested for damage against non-projectiles
}

func (e *Projectile) DamageHitPoints(damageIgnored float64) float64 {
	// projectiles are only tested for damage against non-projectiles
	return projectileHitPointsIgnored
}

func (e *Projectile) MaxHitPoints() float64 {
	// projectiles are only tested for damage against non-projectiles
	return projectileHitPointsIgnored
}

func (e *Projectile) Parent() Entity {
	return e.parent
}

func (e *Projectile) SetParent(parent Entity) {
	e.parent = parent
}
