package model

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type Entity interface {
	Pos() *geom.Vector2
	SetPos(*geom.Vector2)
	PosZ() float64
	SetPosZ(float64)

	Scale() float64
	SetScale(float64)
	Anchor() raycaster.SpriteAnchor
	SetAnchor(raycaster.SpriteAnchor)

	Angle() float64
	SetAngle(float64)
	Pitch() float64
	SetPitch(float64)
	Velocity() float64
	SetVelocity(float64)

	CollisionRadius() float64
	SetCollisionRadius(float64)
	CollisionHeight() float64
	SetCollisionHeight(float64)

	HitPoints() float64
	SetHitPoints(float64)
	DamageHitPoints(float64) float64

	Parent() Entity
	SetParent(Entity)
}

type BasicEntity struct {
	position        *geom.Vector2
	positionZ       float64
	scale           float64
	anchor          raycaster.SpriteAnchor
	angle           float64
	pitch           float64
	velocity        float64
	collisionRadius float64
	collisionHeight float64
	hitPoints       float64
	parent          Entity
}

func (e *BasicEntity) Pos() *geom.Vector2 {
	return e.position
}

func (e *BasicEntity) SetPos(pos *geom.Vector2) {
	e.position = pos
}

func (e *BasicEntity) PosZ() float64 {
	return e.positionZ
}

func (e *BasicEntity) SetPosZ(posZ float64) {
	e.positionZ = posZ
}

func (e *BasicEntity) Scale() float64 {
	return e.scale
}

func (e *BasicEntity) SetScale(scale float64) {
	e.scale = scale
}

func (e *BasicEntity) Anchor() raycaster.SpriteAnchor {
	return e.anchor
}

func (e *BasicEntity) SetAnchor(anchor raycaster.SpriteAnchor) {
	e.anchor = anchor
}

func (e *BasicEntity) Angle() float64 {
	return e.angle
}

func (e *BasicEntity) SetAngle(angle float64) {
	e.angle = angle
}

func (e *BasicEntity) Pitch() float64 {
	return e.pitch
}

func (e *BasicEntity) SetPitch(pitch float64) {
	e.pitch = pitch
}

func (e *BasicEntity) Velocity() float64 {
	return e.velocity
}

func (e *BasicEntity) SetVelocity(velocity float64) {
	e.velocity = velocity
}

func (e *BasicEntity) CollisionRadius() float64 {
	return e.collisionRadius
}

func (e *BasicEntity) SetCollisionRadius(collisionRadius float64) {
	e.collisionRadius = collisionRadius
}

func (e *BasicEntity) CollisionHeight() float64 {
	return e.collisionHeight
}

func (e *BasicEntity) SetCollisionHeight(collisionHeight float64) {
	e.collisionHeight = collisionHeight
}

func (e *BasicEntity) HitPoints() float64 {
	return e.hitPoints
}

func (e *BasicEntity) SetHitPoints(hitPoints float64) {
	e.hitPoints = hitPoints
}

func (e *BasicEntity) DamageHitPoints(damage float64) float64 {
	e.hitPoints -= damage
	return e.hitPoints
}

func (e *BasicEntity) Parent() Entity {
	return e.parent
}

func (e *BasicEntity) SetParent(parent Entity) {
	e.parent = parent
}
