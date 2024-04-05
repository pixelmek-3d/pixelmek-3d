package model

import (
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
			unitType:        VehicleUnitType,
			anchor:          raycaster.AnchorBottom,
			collisionRadius: collisionRadius,
			collisionHeight: collisionHeight,
			cockpitOffset:   cockpitOffset,
			armor:           r.Armor,
			structure:       r.Structure,
			heatSinks:       r.HeatSinks.Quantity,
			heatSinkType:    r.HeatSinks.Type.HeatSinkType,
			armament:        make([]Weapon, 0),
			ammunition:      NewAmmoStock(),
			hasTurret:       true,
			maxVelocity:     r.Speed * KPH_TO_VELOCITY,
			maxTurnRate:     100 / r.Tonnage * 0.015, // FIXME: testing
			maxTurretRate:   100 / r.Tonnage * 0.03,  // FIXME: testing
			jumpJets:        0,
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

func (e *Vehicle) Tonnage() float64 {
	return e.Resource.Tonnage
}

func (e *Vehicle) MaxArmorPoints() float64 {
	return e.Resource.Armor
}

func (e *Vehicle) MaxStructurePoints() float64 {
	return e.Resource.Structure
}

func (e *Vehicle) Update() bool {
	e.UnitModel.update()

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
		newH = ClampAngle(newH)

		e.targetRelHeading -= deltaH
		e.heading = newH
	}

	// position update needed
	return true
}
