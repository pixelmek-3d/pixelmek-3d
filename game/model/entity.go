package model

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/jinzhu/copier"
)

type Entity interface {
	Pos() *geom.Vector2
	SetPos(*geom.Vector2)
	PosZ() float64
	SetPosZ(float64)

	Anchor() raycaster.SpriteAnchor
	SetAnchor(raycaster.SpriteAnchor)

	Heading() float64
	SetHeading(float64)
	Pitch() float64
	SetPitch(float64)
	Velocity() float64
	SetVelocity(float64)
	VelocityZ() float64
	SetVelocityZ(float64)

	CollisionRadius() float64
	SetCollisionRadius(float64)
	CollisionHeight() float64
	SetCollisionHeight(float64)

	ApplyDamage(float64)
	ArmorPoints() float64
	SetArmorPoints(float64)
	MaxArmorPoints() float64
	StructurePoints() float64
	SetStructurePoints(float64)
	MaxStructurePoints() float64
	IsDestroyed() bool
	Team() int
	SetTeam(int)

	Clone() Entity
	Parent() Entity
	SetParent(Entity)
}

func EntityDistance(e1, e2 Entity) float64 {
	pos1, pos2 := e1.Pos(), e2.Pos()
	line := geom3d.Line3d{
		X1: pos1.X, Y1: pos1.Y, Z1: e1.PosZ(),
		X2: pos2.X, Y2: pos2.Y, Z2: e2.PosZ(),
	}
	return line.Distance()
}

func EntityDistance2D(e1, e2 Entity) float64 {
	pos1, pos2 := e1.Pos(), e2.Pos()
	line := geom.Line{
		X1: pos1.X, Y1: pos1.Y,
		X2: pos2.X, Y2: pos2.Y,
	}
	return line.Distance()
}

type BasicEntity struct {
	position                *geom.Vector2
	positionZ               float64
	anchor                  raycaster.SpriteAnchor
	angle                   float64
	pitch                   float64
	velocity                float64
	velocityZ               float64
	collisionRadius         float64
	collisionHeight         float64
	armor, maxArmor         float64
	structure, maxStructure float64
	team                    int
	parent                  Entity
}

func BasicCollisionEntity(x, y, z float64, anchor raycaster.SpriteAnchor, collisionRadius, collisionHeight, hitPoints float64) *BasicEntity {
	e := &BasicEntity{
		position:        &geom.Vector2{X: x, Y: y},
		positionZ:       z,
		anchor:          anchor,
		collisionRadius: collisionRadius,
		collisionHeight: collisionHeight,
		armor:           0,
		maxArmor:        0,
		structure:       hitPoints,
		maxStructure:    hitPoints,
	}
	return e
}

func BasicVisualEntity(x, y, z float64, anchor raycaster.SpriteAnchor) *BasicEntity {
	e := &BasicEntity{
		position:        &geom.Vector2{X: x, Y: y},
		positionZ:       z,
		anchor:          anchor,
		collisionRadius: 0,
		collisionHeight: 0,
	}
	return e
}

func (e *BasicEntity) Clone() Entity {
	eClone := &BasicEntity{}
	copier.Copy(eClone, e)
	return eClone
}

func (e *BasicEntity) Team() int {
	return e.team
}

func (e *BasicEntity) SetTeam(team int) {
	e.team = team
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

func (e *BasicEntity) Anchor() raycaster.SpriteAnchor {
	return e.anchor
}

func (e *BasicEntity) SetAnchor(anchor raycaster.SpriteAnchor) {
	e.anchor = anchor
}

func (e *BasicEntity) Heading() float64 {
	return e.angle
}

func (e *BasicEntity) SetHeading(angle float64) {
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

func (e *BasicEntity) VelocityZ() float64 {
	return e.velocityZ
}

func (e *BasicEntity) SetVelocityZ(velocityZ float64) {
	e.velocityZ = velocityZ
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

func (e *BasicEntity) ApplyDamage(damage float64) {
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

func (e *BasicEntity) ArmorPoints() float64 {
	return e.armor
}

func (e *BasicEntity) SetArmorPoints(armor float64) {
	e.armor = armor
}

func (e *BasicEntity) MaxArmorPoints() float64 {
	return e.maxArmor
}

func (e *BasicEntity) StructurePoints() float64 {
	return e.structure
}

func (e *BasicEntity) SetStructurePoints(structure float64) {
	e.structure = structure
}

func (e *BasicEntity) MaxStructurePoints() float64 {
	return e.maxStructure
}

func (e *BasicEntity) IsDestroyed() bool {
	return e.structure <= 0
}

func (e *BasicEntity) Parent() Entity {
	return e.parent
}

func (e *BasicEntity) SetParent(parent Entity) {
	e.parent = parent
}
