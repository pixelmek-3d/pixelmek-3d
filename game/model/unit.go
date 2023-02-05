package model

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

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
	heading          float64
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

func (e *UnitModel) Pitch() float64 {
	return e.pitch
}

func (e *UnitModel) Heat() float64 {
	return e.heat
}

func (e *UnitModel) HeatDissipation() float64 {
	return e.heatDissipation
}

func (e *UnitModel) TriggerWeapon(w Weapon) bool {
	if w.Cooldown() > 0 {
		return false
	}

	w.TriggerCooldown()
	e.heat += w.Heat()
	return true
}

func (e *UnitModel) Target() Entity {
	return e.target
}

func (e *UnitModel) SetTarget(t Entity) {
	e.target = t
}

func (e *UnitModel) HasTurret() bool {
	return e.hasTurret
}

func (e *UnitModel) SetHasTurret(hasTurret bool) {
	e.hasTurret = hasTurret
}

func (e *UnitModel) TurretAngle() float64 {
	if e.hasTurret {
		return e.turretAngle
	}
	return 0
}

func (e *UnitModel) SetTurretAngle(angle float64) {
	if e.hasTurret {
		e.turretAngle = angle
	} else {
		e.SetHeading(angle)
	}
}

func (e *UnitModel) Armament() []Weapon {
	return e.armament
}

func (e *UnitModel) AddArmament(w Weapon) {
	e.armament = append(e.armament, w)
}

func (e *UnitModel) Pos() *geom.Vector2 {
	return e.position
}

func (e *UnitModel) SetPos(pos *geom.Vector2) {
	e.position = pos
}

func (e *UnitModel) PosZ() float64 {
	return e.positionZ
}

func (e *UnitModel) SetPosZ(posZ float64) {
	e.positionZ = posZ
}

func (e *UnitModel) Anchor() raycaster.SpriteAnchor {
	return e.anchor
}

func (e *UnitModel) SetAnchor(anchor raycaster.SpriteAnchor) {
	e.anchor = anchor
}

func (e *UnitModel) Heading() float64 {
	return e.heading
}

func (e *UnitModel) SetHeading(angle float64) {
	e.heading = angle
}

func (e *UnitModel) SetPitch(pitch float64) {
	e.pitch = pitch
}

func (e *UnitModel) TurnRate() float64 {
	if e.velocity == 0 {
		return e.maxTurnRate
	}

	// dynamic turn rate is half of the max turn rate when at max velocity
	vTurnRatio := 0.5 + 0.5*(e.maxVelocity-math.Abs(e.velocity))/e.maxVelocity
	return e.maxTurnRate * vTurnRatio
}

func (e *UnitModel) SetTargetRelativeHeading(rHeading float64) {
	e.targetRelHeading = rHeading
}

func (e *UnitModel) Velocity() float64 {
	return e.velocity
}

func (e *UnitModel) SetVelocity(velocity float64) {
	e.velocity = velocity
}

func (e *UnitModel) VelocityZ() float64 {
	return e.velocityZ
}

func (e *UnitModel) SetVelocityZ(velocityZ float64) {
	e.velocityZ = velocityZ
}

func (e *UnitModel) MaxVelocity() float64 {
	return e.maxVelocity
}

func (e *UnitModel) TargetVelocity() float64 {
	return e.targetVelocity
}

func (e *UnitModel) SetTargetVelocity(tVelocity float64) {
	maxV := e.MaxVelocity()
	if tVelocity > maxV {
		tVelocity = maxV
	} else if tVelocity < -maxV/2 {
		tVelocity = -maxV / 2
	}
	e.targetVelocity = tVelocity
}

func (e *UnitModel) TargetVelocityZ() float64 {
	return e.targetVelocityZ
}

func (e *UnitModel) SetTargetVelocityZ(tVelocityZ float64) {
	e.targetVelocityZ = tVelocityZ
}

func (e *UnitModel) CollisionRadius() float64 {
	return e.collisionRadius
}

func (e *UnitModel) SetCollisionRadius(collisionRadius float64) {
	e.collisionRadius = collisionRadius
}

func (e *UnitModel) CollisionHeight() float64 {
	return e.collisionHeight
}

func (e *UnitModel) SetCollisionHeight(collisionHeight float64) {
	e.collisionHeight = collisionHeight
}

func (e *UnitModel) CockpitOffset() *geom.Vector2 {
	return e.cockpitOffset
}

func (e *UnitModel) ApplyDamage(damage float64) {
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

func (e *UnitModel) ArmorPoints() float64 {
	return e.armor
}

func (e *UnitModel) SetArmorPoints(armor float64) {
	e.armor = armor
}

func (e *UnitModel) StructurePoints() float64 {
	return e.structure
}

func (e *UnitModel) SetStructurePoints(structure float64) {
	e.structure = structure
}

func (e *UnitModel) Parent() Entity {
	return e.parent
}

func (e *UnitModel) SetParent(parent Entity) {
	e.parent = parent
}

func (e *UnitModel) SetAsPlayer(isPlayer bool) {
	e.isPlayer = isPlayer
}

func (e *UnitModel) IsPlayer() bool {
	return e.isPlayer
}
