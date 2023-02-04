package model

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type Infantry struct {
	*UnitModel
	Resource *ModelInfantryResource
}

func NewInfantry(r *ModelInfantryResource, collisionRadius, collisionHeight float64, cockpitOffset *geom.Vector2) *Infantry {
	m := &Infantry{
		Resource: r,
		UnitModel: &UnitModel{
			anchor:          raycaster.AnchorBottom,
			collisionRadius: collisionRadius,
			collisionHeight: collisionHeight,
			cockpitOffset:   cockpitOffset,
			armor:           r.Armor,
			structure:       r.Structure,
			armament:        make([]Weapon, 0),
			maxVelocity:     r.Speed * KPH_TO_VELOCITY,
			maxTurnRate:     0.05, // FIXME: testing
		},
	}
	return m
}

func (e *Infantry) CloneUnit() Unit {
	eClone := &Infantry{}
	copier.Copy(eClone, e)

	// weapons needs to be cloned since copier does not handle them automatically
	eClone.armament = make([]Weapon, 0, len(e.armament))
	for _, weapon := range e.armament {
		eClone.AddArmament(weapon.Clone())
	}

	return eClone
}

func (e *Infantry) Clone() Entity {
	return e.CloneUnit()
}

func (e *Infantry) Name() string {
	return e.Resource.Name
}

func (e *Infantry) Variant() string {
	return e.Resource.Variant
}

func (e *Infantry) Heat() float64 {
	return 0
}

func (e *Infantry) HeatDissipation() float64 {
	return 0
}

func (e *Infantry) TriggerWeapon(w Weapon) bool {
	if w.Cooldown() > 0 {
		return false
	}
	w.TriggerCooldown()
	return true
}

func (e *Infantry) Target() Entity {
	return e.target
}

func (e *Infantry) SetTarget(t Entity) {
	e.target = t
}

func (e *Infantry) HasTurret() bool {
	return false
}

func (e *Infantry) SetHasTurret(bool) {}

func (e *Infantry) TurretAngle() float64 {
	return 0
}

func (e *Infantry) SetTurretAngle(angle float64) {
	// infantry have no turret, just set heading
	e.SetHeading(angle)
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

func (e *Infantry) VelocityZ() float64 {
	return e.velocityZ
}

func (e *Infantry) SetVelocityZ(velocityZ float64) {
	e.velocityZ = velocityZ
}

func (e *Infantry) MaxVelocity() float64 {
	return e.maxVelocity
}

func (e *Infantry) TargetVelocity() float64 {
	return e.targetVelocity
}

func (e *Infantry) SetTargetVelocity(tVelocity float64) {
	maxV := e.MaxVelocity()
	if tVelocity > maxV {
		tVelocity = maxV
	} else if tVelocity < -maxV/2 {
		tVelocity = -maxV / 2
	}
	e.targetVelocity = tVelocity
}

func (e *Infantry) TargetVelocityZ() float64 {
	return e.targetVelocityZ
}

func (e *Infantry) SetTargetVelocityZ(tVelocityZ float64) {
	maxV := e.MaxVelocity()
	if tVelocityZ > maxV/2 {
		tVelocityZ = maxV / 2
	} else if tVelocityZ < -maxV/2 {
		tVelocityZ = -maxV / 2
	}
	e.targetVelocityZ = tVelocityZ
}

func (e *Infantry) TurnRate() float64 {
	return e.maxTurnRate
}

func (e *Infantry) SetTargetRelativeHeading(rHeading float64) {
	e.targetRelHeading = rHeading
}

func (e *Infantry) Update() bool {
	if e.targetRelHeading == 0 &&
		e.targetVelocity == 0 && e.velocity == 0 &&
		e.targetVelocityZ == 0 && e.velocityZ == 0 {
		// no position update needed
		return false
	}

	if e.velocity != e.targetVelocity {
		e.velocity = e.targetVelocity
	}

	if e.targetRelHeading != 0 {
		// move by relative heading amount allowed by calculated turn rate
		var deltaH, maxDeltaH, newH float64
		newH = e.Heading()
		maxDeltaH = e.TurnRate()
		if e.targetRelHeading > 0 {
			deltaH = e.targetRelHeading
			if deltaH > maxDeltaH {
				deltaH = maxDeltaH
			}
		} else {
			deltaH = e.targetRelHeading
			if deltaH < -maxDeltaH {
				deltaH = -maxDeltaH
			}
		}

		newH += deltaH

		if newH >= geom.Pi2 {
			newH = geom.Pi2 - newH
		} else if newH < 0 {
			newH = newH + geom.Pi2
		}

		if newH < 0 {
			// handle rounding errors
			newH = 0
		}

		e.targetRelHeading -= deltaH
		e.angle = newH
	}

	// position update needed
	return true
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

func (e *Infantry) SetAsPlayer(isPlayer bool) {
	e.isPlayer = isPlayer
}

func (e *Infantry) IsPlayer() bool {
	return e.isPlayer
}
