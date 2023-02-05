package model

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type VTOL struct {
	*UnitModel
	Resource *ModelVTOLResource
}

func NewVTOL(r *ModelVTOLResource, collisionRadius, collisionHeight float64, cockpitOffset *geom.Vector2) *VTOL {
	m := &VTOL{
		Resource: r,
		UnitModel: &UnitModel{
			anchor:          raycaster.AnchorCenter,
			collisionRadius: collisionRadius,
			collisionHeight: collisionHeight,
			cockpitOffset:   cockpitOffset,
			armor:           r.Armor,
			structure:       r.Structure,
			heatSinks:       r.HeatSinks.Quantity,
			heatSinkType:    r.HeatSinks.Type.HeatSinkType,
			armament:        make([]Weapon, 0),
			maxVelocity:     r.Speed * KPH_TO_VELOCITY,
			maxTurnRate:     100 / r.Tonnage * 0.03, // FIXME: testing
			jumpJets:        0,
		},
	}

	// calculate heat dissipation per tick
	m.heatDissipation = SECONDS_PER_TICK / 4 * float64(m.heatSinks) * float64(m.heatSinkType)

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

func (e *VTOL) MaxArmorPoints() float64 {
	return e.Resource.Armor
}

func (e *VTOL) MaxStructurePoints() float64 {
	return e.Resource.Structure
}

func (e *VTOL) SetTargetVelocityZ(tVelocityZ float64) {
	// VTOL have throttle based vertical velocity
	maxV := e.MaxVelocity()
	if tVelocityZ > maxV/2 {
		tVelocityZ = maxV / 2
	} else if tVelocityZ < -maxV/2 {
		tVelocityZ = -maxV / 2
	}
	e.UnitModel.SetTargetVelocityZ(tVelocityZ)
}

func (e *VTOL) Update() bool {
	if e.heat > 0 {
		// TODO: apply heat from movement based on velocity

		// apply heat dissipation
		e.heat -= e.HeatDissipation()
		if e.heat < 0 {
			e.heat = 0
		}
	}

	if e.targetRelHeading == 0 &&
		e.targetVelocity == 0 && e.velocity == 0 &&
		e.targetVelocityZ == 0 && e.velocityZ == 0 {
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

	if e.targetVelocityZ != e.velocityZ || e.positionZ >= CEILING_VTOL {
		// TODO: move vertical velocity toward target by amount allowed by calculated vertical acceleration
		var zDeltaV, zNewV float64
		if e.targetVelocityZ > e.velocityZ {
			zDeltaV = 0.0004 // FIXME: testing
		} else {
			zDeltaV = -0.0004 // FIXME: testing
		}

		zNewV = e.velocityZ + zDeltaV
		if (zDeltaV > 0 && e.targetVelocityZ >= 0 && zNewV > e.targetVelocityZ) ||
			(zDeltaV < 0 && e.targetVelocityZ <= 0 && zNewV < e.targetVelocityZ) {
			// bound velocity changes to target velocity
			zNewV = e.targetVelocityZ
		}

		if zNewV > 0 && e.positionZ >= CEILING_VTOL {
			// restrict vertical flight height
			zNewV = 0
		}

		e.velocityZ = zNewV
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
		e.heading = newH
	}

	// position update needed
	return true
}
