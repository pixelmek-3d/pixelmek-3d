package model

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
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

	Clone() Entity
	Parent() Entity
	SetParent(Entity)
}

type Unit interface {
	Entity
	Name() string
	Variant() string

	Heat() float64
	HeatDissipation() float64
	TriggerWeapon(Weapon) bool

	Target() Entity
	SetTarget(Entity)

	TurnRate() float64
	SetTargetRelativeHeading(float64)
	MaxVelocity() float64
	TargetVelocity() float64
	SetTargetVelocity(float64)
	TargetVelocityZ() float64
	SetTargetVelocityZ(float64)
	Update() bool

	HasTurret() bool
	SetHasTurret(bool)
	TurretAngle() float64
	SetTurretAngle(float64)

	CockpitOffset() *geom.Vector2
	Armament() []Weapon
	AddArmament(Weapon)

	SetAsPlayer(bool)
	IsPlayer() bool

	CloneUnit() Unit
}

type UnitModel struct {
	position         *geom.Vector2
	positionZ        float64
	anchor           raycaster.SpriteAnchor
	angle            float64
	targetRelHeading float64
	maxTurnRate      float64
	pitch            float64
	hasTurret        bool
	turretAngle      float64
	velocity         float64
	velocityZ        float64
	targetVelocity   float64
	targetVelocityZ  float64
	maxVelocity      float64
	collisionRadius  float64
	collisionHeight  float64
	cockpitOffset    *geom.Vector2
	armor            float64
	structure        float64
	heat             float64
	heatDissipation  float64
	heatSinks        int
	heatSinkType     HeatSinkType
	armament         []Weapon
	target           Entity
	parent           Entity
	isPlayer         bool
}

func EntityUnit(entity Entity) Unit {
	if unit, ok := entity.(Unit); ok {
		return unit
	}
	return nil
}

func IsEntityUnit(entity Entity) bool {
	_, ok := entity.(Unit)
	return ok
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

func (e *BasicEntity) Parent() Entity {
	return e.parent
}

func (e *BasicEntity) SetParent(parent Entity) {
	e.parent = parent
}
