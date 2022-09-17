package model

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type Vehicle struct {
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

func NewVehicle(collisionRadius, collisionHeight, hitPoints float64) *Vehicle {
	m := &Vehicle{
		anchor:          raycaster.AnchorBottom,
		collisionRadius: collisionRadius,
		collisionHeight: collisionHeight,
		hitPoints:       hitPoints,
		maxHitPoints:    hitPoints,
	}
	return m
}

func (e *Vehicle) Pos() *geom.Vector2 {
	return e.position
}

func (e *Vehicle) SetPos(pos *geom.Vector2) {
	e.position = pos
}

func (e *Vehicle) PosZ() float64 {
	return e.positionZ
}

func (e *Vehicle) SetPosZ(posZ float64) {
	e.positionZ = posZ
}

func (e *Vehicle) Anchor() raycaster.SpriteAnchor {
	return e.anchor
}

func (e *Vehicle) SetAnchor(anchor raycaster.SpriteAnchor) {
	e.anchor = anchor
}

func (e *Vehicle) Angle() float64 {
	return e.angle
}

func (e *Vehicle) SetAngle(angle float64) {
	e.angle = angle
}

func (e *Vehicle) Pitch() float64 {
	return e.pitch
}

func (e *Vehicle) SetPitch(pitch float64) {
	e.pitch = pitch
}

func (e *Vehicle) Velocity() float64 {
	return e.velocity
}

func (e *Vehicle) SetVelocity(velocity float64) {
	e.velocity = velocity
}

func (e *Vehicle) CollisionRadius() float64 {
	return e.collisionRadius
}

func (e *Vehicle) SetCollisionRadius(collisionRadius float64) {
	e.collisionRadius = collisionRadius
}

func (e *Vehicle) CollisionHeight() float64 {
	return e.collisionHeight
}

func (e *Vehicle) SetCollisionHeight(collisionHeight float64) {
	e.collisionHeight = collisionHeight
}

func (e *Vehicle) HitPoints() float64 {
	return e.hitPoints
}

func (e *Vehicle) SetHitPoints(hitPoints float64) {
	e.hitPoints = hitPoints
}

func (e *Vehicle) DamageHitPoints(damage float64) float64 {
	e.hitPoints -= damage
	return e.hitPoints
}

func (e *Vehicle) MaxHitPoints() float64 {
	return e.maxHitPoints
}

func (e *Vehicle) Parent() Entity {
	return e.parent
}

func (e *Vehicle) SetParent(parent Entity) {
	e.parent = parent
}
