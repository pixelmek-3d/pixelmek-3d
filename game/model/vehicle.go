package model

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type Vehicle struct {
	*UnitModel
	Resource *ModelVehicleResource
}

func NewVehicle(r *ModelVehicleResource, collisionRadius, collisionHeight float64, cockpitOffset *geom.Vector2) *Vehicle {
	m := &Vehicle{
		Resource: r,
		UnitModel: &UnitModel{
			anchor:          raycaster.AnchorBottom,
			collisionRadius: collisionRadius,
			collisionHeight: collisionHeight,
			cockpitOffset:   cockpitOffset,
			armor:           r.Armor,
			structure:       r.Structure,
			heatSinks:       r.HeatSinks.Quantity,
			heatSinkType:    r.HeatSinks.Type.HeatSinkType,
			armament:        make([]Weapon, 0),
			hasTurret:       true,
			maxVelocity:     r.Speed * KPH_TO_VELOCITY,
			maxTurnRate:     100 / r.Tonnage * 0.015, // FIXME: testing
		},
	}

	// calculate heat dissipation per tick
	m.heatDissipation = SECONDS_PER_TICK / 4 * float64(m.heatSinks) * float64(m.heatSinkType)

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

func (e *Vehicle) Heat() float64 {
	return e.heat
}

func (e *Vehicle) HeatDissipation() float64 {
	return e.heatDissipation
}

func (e *Vehicle) TriggerWeapon(w Weapon) bool {
	if w.Cooldown() > 0 {
		return false
	}

	w.TriggerCooldown()
	e.heat += w.Heat()
	return true
}

func (e *Vehicle) Target() Entity {
	return e.target
}

func (e *Vehicle) SetTarget(t Entity) {
	e.target = t
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
	return 0
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

func (e *Vehicle) VelocityZ() float64 {
	return e.velocityZ
}

func (e *Vehicle) SetVelocityZ(velocityZ float64) {
	e.velocityZ = velocityZ
}

func (e *Vehicle) MaxVelocity() float64 {
	return e.maxVelocity
}

func (e *Vehicle) TargetVelocity() float64 {
	return e.targetVelocity
}

func (e *Vehicle) SetTargetVelocity(tVelocity float64) {
	maxV := e.MaxVelocity()
	if tVelocity > maxV {
		tVelocity = maxV
	} else if tVelocity < -maxV/2 {
		tVelocity = -maxV / 2
	}
	e.targetVelocity = tVelocity
}

func (e *Vehicle) TargetVelocityZ() float64 {
	return e.targetVelocityZ
}

func (e *Vehicle) SetTargetVelocityZ(tVelocityZ float64) {
	maxV := e.MaxVelocity()
	if tVelocityZ > maxV {
		tVelocityZ = maxV
	} else if tVelocityZ < -maxV/2 {
		tVelocityZ = -maxV / 2
	}
	e.targetVelocityZ = tVelocityZ
}

func (e *Vehicle) TurnRate() float64 {
	if e.velocity == 0 {
		return e.maxTurnRate
	}

	// dynamic turn rate is half of the max turn rate when at max velocity
	vTurnRatio := 0.5 + 0.5*(e.maxVelocity-math.Abs(e.velocity))/e.maxVelocity
	return e.maxTurnRate * vTurnRatio
}

func (e *Vehicle) SetTargetRelativeHeading(rHeading float64) {
	e.targetRelHeading = rHeading
}

func (e *Vehicle) Update() bool {
	if e.velocity == 0 && e.targetVelocity == 0 && e.targetRelHeading == 0 {
		// no position update needed
		return false
	}

	if e.targetVelocity != e.velocity {
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
