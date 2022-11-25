package model

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type VTOL struct {
	Resource        *ModelVTOLResource
	position        *geom.Vector2
	positionZ       float64
	anchor          raycaster.SpriteAnchor
	angle           float64
	pitch           float64
	velocity        float64
	targetVelocity  float64
	maxVelocity     float64
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

func NewVTOL(r *ModelVTOLResource, collisionRadius, collisionHeight float64, cockpitOffset *geom.Vector2) *VTOL {
	m := &VTOL{
		Resource:        r,
		anchor:          raycaster.AnchorCenter,
		collisionRadius: collisionRadius,
		collisionHeight: collisionHeight,
		cockpitOffset:   cockpitOffset,
		armor:           r.Armor,
		structure:       r.Structure,
		heatSinks:       r.HeatSinks.Quantity,
		heatSinkType:    r.HeatSinks.Type,
		armament:        make([]Weapon, 0),
		maxVelocity:     r.Speed * KPH_TO_VELOCITY,
	}
	return m
}

func (e *VTOL) CloneUnit() Unit {
	eClone := &VTOL{}
	copier.Copy(eClone, e)

	// weapons needs to be cloned since copier does not handle them automatically
	eClone.armament = make([]Weapon, 0, len(e.armament))
	for _, weapon := range e.armament {
		eClone.AddArmament(weapon.Clone())
	}

	return eClone
}

func (e *VTOL) Clone() Entity {
	return e.CloneUnit()
}

func (e *VTOL) Name() string {
	return e.Resource.Name
}

func (e *VTOL) Variant() string {
	return e.Resource.Variant
}

func (e *VTOL) HasTurret() bool {
	return false
}

func (e *VTOL) SetHasTurret(bool) {}

func (e *VTOL) TurretAngle() float64 {
	return 0
}

func (e *VTOL) SetTurretAngle(angle float64) {
	// VTOL have no turret, just set heading
	e.SetHeading(angle)
}

func (e *VTOL) AddArmament(w Weapon) {
	e.armament = append(e.armament, w)
}

func (e *VTOL) Armament() []Weapon {
	return e.armament
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

func (e *VTOL) Heading() float64 {
	return e.angle
}

func (e *VTOL) SetHeading(angle float64) {
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

func (e *VTOL) MaxVelocity() float64 {
	return e.maxVelocity
}

func (e *VTOL) TargetVelocity() float64 {
	return e.targetVelocity
}

func (e *VTOL) SetTargetVelocity(velocity float64) {
	e.targetVelocity = velocity
}

func (e *VTOL) Update() bool {
	if e.velocity == 0 && e.targetVelocity == 0 { // TODO: update for heading/targetHeading
		// no position update needed
		return false
	}

	if e.velocity != e.targetVelocity {
		// TODO: move velocity toward target by amount allowed by calculated acceleration
		var deltaV, newV float64
		if e.targetVelocity > e.velocity {
			deltaV = 0.0002 // FIXME: testing
		} else {
			deltaV = -0.0002 // FIXME: testing
		}

		newV = e.velocity + deltaV
		if (deltaV > 0 && e.targetVelocity >= 0 && newV > e.targetVelocity) ||
			(deltaV < 0 && e.targetVelocity <= 0 && newV < e.targetVelocity) {
			// bound velocity changes to target velocity
			newV = e.targetVelocity
		}

		e.velocity = newV
	}

	// position update needed
	return true
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

func (e *VTOL) CockpitOffset() *geom.Vector2 {
	return e.cockpitOffset
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

func (e *VTOL) SetAsPlayer(isPlayer bool) {
	e.isPlayer = isPlayer
}

func (e *VTOL) IsPlayer() bool {
	return e.isPlayer
}
