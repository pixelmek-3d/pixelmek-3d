package model

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type VTOL struct {
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

func NewVTOL(collisionRadius, collisionHeight, hitPoints float64) *VTOL {
	m := &VTOL{
		anchor:          raycaster.AnchorCenter,
		collisionRadius: collisionRadius,
		collisionHeight: collisionHeight,
		hitPoints:       hitPoints,
		maxHitPoints:    hitPoints,
	}
	return m
}

func (e *VTOL) Pos() *geom.Vector2 {
	return e.position
}

func (e *VTOL) SetPos(pos *geom.Vector2) {
	e.position = pos
}

func (e *VTOL) PosZ() float64 {
	return e.positionZ
}

func (e *VTOL) SetPosZ(posZ float64) {
	e.positionZ = posZ
}

func (e *VTOL) Anchor() raycaster.SpriteAnchor {
	return e.anchor
}

func (e *VTOL) SetAnchor(anchor raycaster.SpriteAnchor) {
	e.anchor = anchor
}

func (e *VTOL) Angle() float64 {
	return e.angle
}

func (e *VTOL) SetAngle(angle float64) {
	e.angle = angle
}

func (e *VTOL) Pitch() float64 {
	return e.pitch
}

func (e *VTOL) SetPitch(pitch float64) {
	e.pitch = pitch
}

func (e *VTOL) Velocity() float64 {
	return e.velocity
}

func (e *VTOL) SetVelocity(velocity float64) {
	e.velocity = velocity
}

func (e *VTOL) CollisionRadius() float64 {
	return e.collisionRadius
}

func (e *VTOL) SetCollisionRadius(collisionRadius float64) {
	e.collisionRadius = collisionRadius
}

func (e *VTOL) CollisionHeight() float64 {
	return e.collisionHeight
}

func (e *VTOL) SetCollisionHeight(collisionHeight float64) {
	e.collisionHeight = collisionHeight
}

func (e *VTOL) HitPoints() float64 {
	return e.hitPoints
}

func (e *VTOL) SetHitPoints(hitPoints float64) {
	e.hitPoints = hitPoints
}

func (e *VTOL) DamageHitPoints(damage float64) float64 {
	e.hitPoints -= damage
	return e.hitPoints
}

func (e *VTOL) MaxHitPoints() float64 {
	return e.maxHitPoints
}

func (e *VTOL) Parent() Entity {
	return e.parent
}

func (e *VTOL) SetParent(parent Entity) {
	e.parent = parent
}
