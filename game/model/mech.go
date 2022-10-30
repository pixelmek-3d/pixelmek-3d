package model

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type Mech struct {
	Resource        *ModelMechResource
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
	heatSinks       int
	heatSinkType    ModelHeatSinkType
	armament        []Weapon
	parent          Entity
	isPlayer        bool
}

func NewMech(r *ModelMechResource, collisionRadius, collisionHeight float64, cockpitOffset *geom.Vector2) *Mech {
	m := &Mech{
		Resource:        r,
		anchor:          raycaster.AnchorBottom,
		collisionRadius: collisionRadius,
		collisionHeight: collisionHeight,
		cockpitOffset:   cockpitOffset,
		armor:           r.Armor,
		structure:       r.Structure,
		heatSinks:       r.HeatSinks.Quantity,
		heatSinkType:    r.HeatSinks.Type,
		armament:        make([]Weapon, 0),
	}
	return m
}

func (e *Mech) CloneUnit() Unit {
	eClone := &Mech{}
	copier.Copy(eClone, e)

	// weapons needs to be cloned since copier does not handle them automatically
	eClone.armament = make([]Weapon, 0, len(e.armament))
	for _, weapon := range e.armament {
		eClone.AddArmament(weapon.Clone())
	}

	return eClone
}

func (e *Mech) Clone() Entity {
	return e.CloneUnit()
}

func (e *Mech) Name() string {
	return e.Resource.Name
}

func (e *Mech) Variant() string {
	return e.Resource.Variant
}

func (e *Mech) AddArmament(w Weapon) {
	e.armament = append(e.armament, w)
}

func (e *Mech) Armament() []Weapon {
	return e.armament
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

func (e *Mech) Heading() float64 {
	return e.angle
}

func (e *Mech) SetHeading(angle float64) {
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

func (e *Mech) CockpitOffset() *geom.Vector2 {
	return e.cockpitOffset
}

func (e *Mech) ApplyDamage(damage float64) {
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

func (e *Mech) ArmorPoints() float64 {
	return e.armor
}

func (e *Mech) SetArmorPoints(armor float64) {
	e.armor = armor
}

func (e *Mech) MaxArmorPoints() float64 {
	return e.Resource.Armor
}

func (e *Mech) StructurePoints() float64 {
	return e.structure
}

func (e *Mech) SetStructurePoints(structure float64) {
	e.structure = structure
}

func (e *Mech) MaxStructurePoints() float64 {
	return e.Resource.Structure
}

func (e *Mech) Parent() Entity {
	return e.parent
}

func (e *Mech) SetParent(parent Entity) {
	e.parent = parent
}

func (e *Mech) SetAsPlayer(isPlayer bool) {
	e.isPlayer = isPlayer
}
func (e *Mech) IsPlayer() bool {
	return e.isPlayer
}
