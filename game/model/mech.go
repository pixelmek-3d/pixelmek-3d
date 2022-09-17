package model

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type Mech struct {
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

func NewMech(collisionRadius, collisionHeight, hitPoints float64) *Mech {
	m := &Mech{
		anchor:          raycaster.AnchorBottom,
		collisionRadius: collisionRadius,
		collisionHeight: collisionHeight,
		hitPoints:       hitPoints,
		maxHitPoints:    hitPoints,
	}
	return m
}

func (e *Mech) Pos() *geom.Vector2 {
	return e.position
}

func (e *Mech) SetPos(pos *geom.Vector2) {
	e.position = pos
}

func (e *Mech) PosZ() float64 {
	return e.positionZ
}

func (e *Mech) SetPosZ(posZ float64) {
	e.positionZ = posZ
}

func (e *Mech) Anchor() raycaster.SpriteAnchor {
	return e.anchor
}

func (e *Mech) SetAnchor(anchor raycaster.SpriteAnchor) {
	e.anchor = anchor
}

func (e *Mech) Angle() float64 {
	return e.angle
}

func (e *Mech) SetAngle(angle float64) {
	e.angle = angle
}

func (e *Mech) Pitch() float64 {
	return e.pitch
}

func (e *Mech) SetPitch(pitch float64) {
	e.pitch = pitch
}

func (e *Mech) Velocity() float64 {
	return e.velocity
}

func (e *Mech) SetVelocity(velocity float64) {
	e.velocity = velocity
}

func (e *Mech) CollisionRadius() float64 {
	return e.collisionRadius
}

func (e *Mech) SetCollisionRadius(collisionRadius float64) {
	e.collisionRadius = collisionRadius
}

func (e *Mech) CollisionHeight() float64 {
	return e.collisionHeight
}

func (e *Mech) SetCollisionHeight(collisionHeight float64) {
	e.collisionHeight = collisionHeight
}

func (e *Mech) HitPoints() float64 {
	return e.hitPoints
}

func (e *Mech) SetHitPoints(hitPoints float64) {
	e.hitPoints = hitPoints
}

func (e *Mech) DamageHitPoints(damage float64) float64 {
	e.hitPoints -= damage
	return e.hitPoints
}

func (e *Mech) MaxHitPoints() float64 {
	return e.maxHitPoints
}

func (e *Mech) Parent() Entity {
	return e.parent
}

func (e *Mech) SetParent(parent Entity) {
	e.parent = parent
}
