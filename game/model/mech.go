package model

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type Mech struct {
	Resource         *ModelMechResource
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
	targetVelocity   float64
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
	parent           Entity
	isPlayer         bool
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
		heatSinkType:    r.HeatSinks.Type.HeatSinkType,
		armament:        make([]Weapon, 0),
		hasTurret:       true,
		maxVelocity:     r.Speed * KPH_TO_VELOCITY,
		maxTurnRate:     100 / r.Tonnage * 0.02, // FIXME: testing
	}

	// calculate heat dissipation per tick
	m.heatDissipation = 5 * SECONDS_PER_TICK * float64(m.heatSinks) * float64(m.heatSinkType+1)

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

func (e *Mech) Heat() float64 {
	return e.heat
}

func (e *Mech) HeatDissipation() float64 {
	return e.heatDissipation
}

func (e *Mech) TriggerWeapon(w Weapon) bool {
	if w.Cooldown() > 0 {
		return false
	}

	w.TriggerCooldown()
	e.heat += w.Heat()
	return true
}

func (e *Mech) HasTurret() bool {
	return e.hasTurret
}

func (e *Mech) SetHasTurret(hasTurret bool) {
	e.hasTurret = hasTurret
}

func (e *Mech) TurretAngle() float64 {
	if e.hasTurret {
		return e.turretAngle
	}
	return 0
}

func (e *Mech) SetTurretAngle(angle float64) {
	if e.hasTurret {
		e.turretAngle = angle
	} else {
		e.SetHeading(angle)
	}
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

func (e *Mech) MaxVelocity() float64 {
	return e.maxVelocity
}

func (e *Mech) TargetVelocity() float64 {
	return e.targetVelocity
}

func (e *Mech) SetTargetVelocity(tVelocity float64) {
	maxV := e.MaxVelocity()
	if tVelocity > maxV {
		tVelocity = maxV
	} else if tVelocity < -maxV/2 {
		tVelocity = -maxV / 2
	}
	e.targetVelocity = tVelocity
}

func (e *Mech) TurnRate() float64 {
	if e.velocity == 0 {
		return e.maxTurnRate
	}

	// dynamic turn rate is half of the max turn rate when at max velocity
	vTurnRatio := 0.5 + 0.5*(e.maxVelocity-math.Abs(e.velocity))/e.maxVelocity
	return e.maxTurnRate * vTurnRatio
}

func (e *Mech) SetTargetRelativeHeading(rHeading float64) {
	e.targetRelHeading = rHeading
}

func (e *Mech) Update() bool {
	if e.heat > 0 {
		// TODO: apply heat from movement based on velocity

		// apply heat dissipation
		e.heat -= e.HeatDissipation()
		if e.heat < 0 {
			e.heat = 0
		}
	}

	if e.targetVelocity == 0 && e.velocity == 0 && e.targetRelHeading == 0 {
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
