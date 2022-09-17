package model

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type Infantry struct {
	position        *geom.Vector2
	positionZ       float64
	anchor          raycaster.SpriteAnchor
	angle           float64
	pitch           float64
	velocity        float64
	collisionRadius float64
	collisionHeight float64
	hitPoints       float64
	maxHitPoints    float64
	parent          Entity
}

func NewInfantry(collisionRadius, collisionHeight, hitPoints float64) *Infantry {
	m := &Infantry{
		anchor:          raycaster.AnchorBottom,
		collisionRadius: collisionRadius,
		collisionHeight: collisionHeight,
		hitPoints:       hitPoints,
		maxHitPoints:    hitPoints,
	}
	return m
}

func (e *Infantry) Pos() *geom.Vector2 {
	return e.position
}

func (e *Infantry) SetPos(pos *geom.Vector2) {
	e.position = pos
}

func (e *Infantry) PosZ() float64 {
	return e.positionZ
}

func (e *Infantry) SetPosZ(posZ float64) {
	e.positionZ = posZ
}

func (e *Infantry) Anchor() raycaster.SpriteAnchor {
	return e.anchor
}

func (e *Infantry) SetAnchor(anchor raycaster.SpriteAnchor) {
	e.anchor = anchor
}

func (e *Infantry) Angle() float64 {
	return e.angle
}

func (e *Infantry) SetAngle(angle float64) {
	e.angle = angle
}

func (e *Infantry) Pitch() float64 {
	return e.pitch
}

func (e *Infantry) SetPitch(pitch float64) {
	e.pitch = pitch
}

func (e *Infantry) Velocity() float64 {
	return e.velocity
}

func (e *Infantry) SetVelocity(velocity float64) {
	e.velocity = velocity
}

func (e *Infantry) CollisionRadius() float64 {
	return e.collisionRadius
}

func (e *Infantry) SetCollisionRadius(collisionRadius float64) {
	e.collisionRadius = collisionRadius
}

func (e *Infantry) CollisionHeight() float64 {
	return e.collisionHeight
}

func (e *Infantry) SetCollisionHeight(collisionHeight float64) {
	e.collisionHeight = collisionHeight
}

func (e *Infantry) HitPoints() float64 {
	return e.hitPoints
}

func (e *Infantry) SetHitPoints(hitPoints float64) {
	e.hitPoints = hitPoints
}

func (e *Infantry) DamageHitPoints(damage float64) float64 {
	e.hitPoints -= damage
	return e.hitPoints
}

func (e *Infantry) MaxHitPoints() float64 {
	return e.maxHitPoints
}

func (e *Infantry) Parent() Entity {
	return e.parent
}

func (e *Infantry) SetParent(parent Entity) {
	e.parent = parent
}
