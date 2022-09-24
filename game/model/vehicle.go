package model

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type Vehicle struct {
	position                *geom.Vector2
	positionZ               float64
	anchor                  raycaster.SpriteAnchor
	angle                   float64
	pitch                   float64
	velocity                float64
	collisionRadius         float64
	collisionHeight         float64
	armor, maxArmor         float64
	structure, maxStructure float64
	parent                  Entity
}

func NewVehicle(r *ModelVehicleResource, collisionRadius, collisionHeight float64) *Vehicle {
	m := &Vehicle{
		anchor:          raycaster.AnchorBottom,
		collisionRadius: collisionRadius,
		collisionHeight: collisionHeight,
		armor:           r.Armor,
		maxArmor:        r.Armor,
		structure:       r.Structure,
		maxStructure:    r.Structure,
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

func (e *Vehicle) ApplyDamage(damage float64) {
	if e.armor > 0 {
		e.armor -= damage
		if e.armor < 0 {
			// apply remainder of armor damage on structure
			e.structure -= math.Abs(e.armor)
			e.armor = 0
		}
	} else {
		e.structure -= damage
	}
}

func (e *Vehicle) ArmorPoints() float64 {
	return e.armor
}

func (e *Vehicle) SetArmorPoints(armor float64) {
	e.armor = armor
}

func (e *Vehicle) MaxArmorPoints() float64 {
	return e.maxArmor
}

func (e *Vehicle) StructurePoints() float64 {
	return e.structure
}

func (e *Vehicle) SetStructurePoints(structure float64) {
	e.structure = structure
}

func (e *Vehicle) MaxStructurePoints() float64 {
	return e.maxStructure
}

func (e *Vehicle) Parent() Entity {
	return e.parent
}

func (e *Vehicle) SetParent(parent Entity) {
	e.parent = parent
}
