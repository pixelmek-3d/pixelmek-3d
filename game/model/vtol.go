package model

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type VTOL struct {
	Resource        *ModelVTOLResource
	position        *geom.Vector2
	positionZ       float64
	anchor          raycaster.SpriteAnchor
	angle           float64
	pitch           float64
	velocity        float64
	collisionRadius float64
	collisionHeight float64
	armor           float64
	structure       float64
	parent          Entity
}

func NewVTOL(r *ModelVTOLResource, collisionRadius, collisionHeight float64) *VTOL {
	m := &VTOL{
		Resource:        r,
		anchor:          raycaster.AnchorCenter,
		collisionRadius: collisionRadius,
		collisionHeight: collisionHeight,
		armor:           r.Armor,
		structure:       r.Structure,
	}
	return m
}

func (e *VTOL) Armament() []Weapon {
	return nil
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

func (e *VTOL) ApplyDamage(damage float64) {
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

func (e *VTOL) ArmorPoints() float64 {
	return e.armor
}

func (e *VTOL) SetArmorPoints(armor float64) {
	e.armor = armor
}

func (e *VTOL) MaxArmorPoints() float64 {
	return e.Resource.Armor
}

func (e *VTOL) StructurePoints() float64 {
	return e.structure
}

func (e *VTOL) SetStructurePoints(structure float64) {
	e.structure = structure
}

func (e *VTOL) MaxStructurePoints() float64 {
	return e.Resource.Structure
}

func (e *VTOL) Parent() Entity {
	return e.parent
}

func (e *VTOL) SetParent(parent Entity) {
	e.parent = parent
}
