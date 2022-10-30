package model

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type Infantry struct {
	Resource        *ModelInfantryResource
	position        *geom.Vector2
	positionZ       float64
	anchor          raycaster.SpriteAnchor
	angle           float64
	pitch           float64
	velocity        float64
	collisionRadius float64
	collisionHeight float64
	cockpitOffset   *geom.Vector2
	armor           float64
	structure       float64
	armament        []Weapon
	parent          Entity
}

func NewInfantry(r *ModelInfantryResource, collisionRadius, collisionHeight float64, cockpitOffset *geom.Vector2) *Infantry {
	m := &Infantry{
		Resource:        r,
		anchor:          raycaster.AnchorBottom,
		collisionRadius: collisionRadius,
		collisionHeight: collisionHeight,
		cockpitOffset:   cockpitOffset,
		armor:           r.Armor,
		structure:       r.Structure,
		armament:        make([]Weapon, 0),
	}
	return m
}

func (e *Infantry) Clone() Entity {
	eClone := &Infantry{}
	copier.Copy(eClone, e)

	// weapons needs to be cloned since copier does not handle them automatically
	eClone.armament = make([]Weapon, 0, len(e.armament))
	for _, weapon := range e.armament {
		eClone.AddArmament(weapon.Clone())
	}

	return eClone
}

func (e *Infantry) Name() string {
	return e.Resource.Name
}

func (e *Infantry) Variant() string {
	return e.Resource.Variant
}

func (e *Infantry) AddArmament(w Weapon) {
	e.armament = append(e.armament, w)
}

func (e *Infantry) Armament() []Weapon {
	return e.armament
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

func (e *Infantry) Heading() float64 {
	return e.angle
}

func (e *Infantry) SetHeading(angle float64) {
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

func (e *Infantry) CockpitOffset() *geom.Vector2 {
	return e.cockpitOffset
}

func (e *Infantry) ApplyDamage(damage float64) {
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

func (e *Infantry) ArmorPoints() float64 {
	return e.armor
}

func (e *Infantry) SetArmorPoints(armor float64) {
	e.armor = armor
}

func (e *Infantry) MaxArmorPoints() float64 {
	return e.Resource.Armor
}

func (e *Infantry) StructurePoints() float64 {
	return e.structure
}

func (e *Infantry) SetStructurePoints(structure float64) {
	e.structure = structure
}

func (e *Infantry) MaxStructurePoints() float64 {
	return e.Resource.Structure
}

func (e *Infantry) Parent() Entity {
	return e.parent
}

func (e *Infantry) SetParent(parent Entity) {
	e.parent = parent
}

func (e *Infantry) SetAsPlayer(bool) {}
func (e *Infantry) IsPlayer() bool {
	return false
}
