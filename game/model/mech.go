package model

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type Mech struct {
	*UnitModel
	Resource *ModelMechResource
}

func NewMech(r *ModelMechResource, collisionRadius, collisionHeight float64, cockpitOffset *geom.Vector2) *Mech {
	m := &Mech{
		Resource: r,
		UnitModel: &UnitModel{
			anchor:             raycaster.AnchorBottom,
			collisionRadius:    collisionRadius,
			collisionHeight:    collisionHeight,
			cockpitOffset:      cockpitOffset,
			armor:              r.Armor,
			structure:          r.Structure,
			heatSinks:          r.HeatSinks.Quantity,
			heatSinkType:       r.HeatSinks.Type.HeatSinkType,
			armament:           make([]Weapon, 0),
			hasTurret:          true,
			maxVelocity:        r.Speed * KPH_TO_VELOCITY,
			maxTurnRate:        100 / r.Tonnage * 0.02, // FIXME: testing
			jumpJets:           r.JumpJets,
			maxJumpJetDuration: float64(r.JumpJets) * 2.0,
		},
	}

	// calculate heat dissipation per tick
	m.heatDissipation = SECONDS_PER_TICK / 4 * float64(m.heatSinks) * float64(m.heatSinkType)

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

func (e *Mech) Tonnage() float64 {
	return e.Resource.Tonnage
}

func (e *Mech) MaxArmorPoints() float64 {
	return e.Resource.Armor
}

func (e *Mech) MaxStructurePoints() float64 {
	return e.Resource.Structure
}

func (e *Mech) Update() bool {
	if e.jumpJetsActive {
		// consume jump jet charge
		e.jumpJetDuration += SECONDS_PER_TICK
		if e.jumpJetDuration >= e.maxJumpJetDuration {
			e.jumpJetDuration = e.maxJumpJetDuration
		}

		// set jump jet heading and velocity only while active
		e.jumpJetHeading = e.heading
		e.jumpJetVelocity = e.velocity

	} else {
		if e.positionZ > 0 {
			if e.jumpJetVelocity != 0 {
				// reduce jump jet velocity in air while jets inactive
				// for simplicity, using gravity and unit tonnage as factor of resistance
				deltaV := 0.5 * GRAVITY_UNITS_PTT * (e.Tonnage() / 100)
				if e.jumpJetVelocity > 0 {
					deltaV = -deltaV
				}

				zeroV := math.Abs(deltaV) > math.Abs(e.jumpJetVelocity)
				if zeroV {
					e.jumpJetVelocity = 0
				} else {
					e.jumpJetVelocity += deltaV
				}
			}
		} else if e.jumpJetDuration > 0 {
			// recharge jump jets when back on solid ground
			e.jumpJetDuration -= float64(e.jumpJets) * SECONDS_PER_TICK / 10
			if e.jumpJetDuration < 0 {
				e.jumpJetDuration = 0
			}
		}
	}

	if e.heat > 0 {
		// TODO: apply heat from movement based on velocity and/or active jump jets

		// apply heat dissipation
		e.heat -= e.HeatDissipation()
		if e.heat < 0 {
			e.heat = 0
		}
	}

	if e.targetRelHeading == 0 && e.positionZ == 0 &&
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

		if e.jumpJetsActive {
			// set jump jet velocity only while jumping
			e.jumpJetVelocity = e.velocity
		}
	}

	if e.targetVelocityZ != e.velocityZ || e.positionZ > 0 {
		// TODO: move vertical velocity toward target by amount allowed by calculated vertical acceleration
		var zDeltaV, zNewV float64
		if e.targetVelocityZ > 0 {
			zDeltaV = 0.0005 // FIXME: testing
		} else if e.positionZ > 0 {
			zDeltaV = -GRAVITY_UNITS_PTT // TODO: model gravity multiplier into map yaml
		}

		zNewV = e.velocityZ + zDeltaV

		if zDeltaV > 0 && e.targetVelocityZ > 0 && zNewV > e.targetVelocityZ {
			// bound velocity changes to target velocity (for jump jets, ascent only)
			zNewV = e.targetVelocityZ
		} else if e.positionZ <= 0 && zNewV < 0 {
			// negative velocity returns to zero when back on the ground
			zNewV = 0
		}

		if zNewV > 0 && e.positionZ >= CEILING_JUMP {
			// restrict jump height
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
		newH = ClampAngle(newH)

		e.targetRelHeading -= deltaH
		e.heading = newH

		if e.jumpJetsActive {
			// set jump jet heading only while jumping
			e.jumpJetHeading = e.heading
		}
	}

	// position update needed
	return true
}
