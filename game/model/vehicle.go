package model

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type Vehicle struct {
	Resource        *ModelVehicleResource
	position        *geom.Vector2
	positionZ       float64
	anchor          raycaster.SpriteAnchor
	angle           float64
	pitch           float64
	hasTurret       bool
	turretAngle     float64
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

func NewVehicle(r *ModelVehicleResource, collisionRadius, collisionHeight float64, cockpitOffset *geom.Vector2) *Vehicle {
	m := &Vehicle{
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
		hasTurret:       true,
	}
	return m
}

func (e *Vehicle) CloneUnit() Unit {
	eClone := &Vehicle{}
	copier.Copy(eClone, e)

	// weapons needs to be cloned since copier does not handle them automatically
	eClone.armament = make([]Weapon, 0, len(e.armament))
	for _, weapon := range e.armament {
		eClone.AddArmament(weapon.Clone())
	}

	return eClone
}

func (e *Vehicle) Clone() Entity {
	return e.CloneUnit()
}

func (e *Vehicle) Name() string {
	return e.Resource.Name
}

func (e *Vehicle) Variant() string {
	return e.Resource.Variant
}

func (e *Vehicle) HasTurret() bool {
	return e.hasTurret
}

func (e *Vehicle) SetHasTurret(hasTurret bool) {
	e.hasTurret = hasTurret
}

func (e *Vehicle) TurretAngle() float64 {
	if e.hasTurret {
		return e.turretAngle
	}
	return e.Heading()
}

func (e *Vehicle) SetTurretAngle(angle float64) {
	if e.hasTurret {
		e.turretAngle = angle
	} else {
		e.SetHeading(angle)
	}
}

func (e *Vehicle) AddArmament(w Weapon) {
	e.armament = append(e.armament, w)
}

func (e *Vehicle) Armament() []Weapon {
	return e.armament
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

func (e *Vehicle) Heading() float64 {
	return e.angle
}

func (e *Vehicle) SetHeading(angle float64) {
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

func (e *Vehicle) CockpitOffset() *geom.Vector2 {
	return e.cockpitOffset
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
	return e.Resource.Armor
}

func (e *Vehicle) StructurePoints() float64 {
	return e.structure
}

func (e *Vehicle) SetStructurePoints(structure float64) {
	e.structure = structure
}

func (e *Vehicle) MaxStructurePoints() float64 {
	return e.Resource.Structure
}

func (e *Vehicle) Parent() Entity {
	return e.parent
}

func (e *Vehicle) SetParent(parent Entity) {
	e.parent = parent
}

func (e *Vehicle) SetAsPlayer(isPlayer bool) {
	e.isPlayer = isPlayer
}

func (e *Vehicle) IsPlayer() bool {
	return e.isPlayer
}
