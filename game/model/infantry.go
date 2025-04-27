package model

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

const (
	INFANTRY_TURN_RATE_FACTOR   float64 = (0.5 * geom.Pi) / TICKS_PER_SECOND
	INFANTRY_TURRET_RATE_FACTOR float64 = 2.0 * INFANTRY_TURN_RATE_FACTOR
)

type Infantry struct {
	*UnitModel
	Resource *ModelInfantryResource
}

func NewInfantry(r *ModelInfantryResource, collisionRadius, collisionHeight float64, cockpitOffset *geom.Vector2) *Infantry {
	m := &Infantry{
		Resource: r,
		UnitModel: &UnitModel{
			name:               r.Name,
			variant:            r.Variant,
			unitType:           InfantryUnitType,
			anchor:             raycaster.AnchorBottom,
			collisionRadius:    collisionRadius,
			collisionHeight:    collisionHeight,
			cockpitOffset:      cockpitOffset,
			armor:              r.Armor,
			structure:          r.Structure,
			armament:           make([]Weapon, 0),
			ammunition:         NewAmmoStock(),
			maxVelocity:        r.Speed * KPH_TO_VELOCITY,
			maxTurnRate:        INFANTRY_TURN_RATE_FACTOR,
			maxTurretRate:      INFANTRY_TURRET_RATE_FACTOR,
			jumpJets:           r.JumpJets,
			maxJumpJetDuration: 1.0,
			powered:            POWER_ON,
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

func (e *Infantry) Tonnage() float64 {
	return 0.1
}

func (e *Infantry) MaxArmorPoints() float64 {
	return e.Resource.Armor
}

func (e *Infantry) MaxStructurePoints() float64 {
	return e.Resource.Structure
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

func (e *Infantry) TurnRate() float64 {
	return e.maxTurnRate
}

func (e *Infantry) Update() bool {
	if e.needsUpdate() {
		e.UnitModel.update()
	} else {
		return false
	}

	if e.velocity != e.targetVelocity {
		e.velocity = e.targetVelocity
	}

	// position update needed
	return true
}
