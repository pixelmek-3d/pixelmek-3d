package model

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

const (
	EMPLACEMENT_TURRET_RATE_FACTOR float64 = (0.25 * geom.Pi) / TICKS_PER_SECOND
)

type Emplacement struct {
	*UnitModel
	Resource *ModelEmplacementResource
}

func NewEmplacement(r *ModelEmplacementResource, collisionRadius, collisionHeight float64, cockpitOffset *geom.Vector2) *Emplacement {
	m := &Emplacement{
		Resource: r,
		UnitModel: &UnitModel{
			name:            r.Name,
			variant:         r.Variant,
			unitType:        EmplacementUnitType,
			anchor:          raycaster.AnchorBottom,
			collisionRadius: collisionRadius,
			collisionHeight: collisionHeight,
			cockpitOffset:   cockpitOffset,
			armor:           r.Armor,
			structure:       r.Structure,
			armament:        make([]Weapon, 0),
			ammunition:      NewAmmoStock(),
			maxVelocity:     0,
			maxTurnRate:     EMPLACEMENT_TURRET_RATE_FACTOR,
			maxTurretRate:   EMPLACEMENT_TURRET_RATE_FACTOR,
			powered:         POWER_ON, // TODO: define initial power status or power on event in mission resource
		},
	}
	return m
}

func (e *Emplacement) CloneUnit() Unit {
	eClone := &Emplacement{}
	copier.Copy(eClone, e)

	// weapons needs to be cloned since copier does not handle them automatically
	eClone.armament = make([]Weapon, 0, len(e.armament))
	for _, weapon := range e.armament {
		eClone.AddArmament(weapon.Clone())
	}

	return eClone
}

func (e *Emplacement) Clone() Entity {
	return e.CloneUnit()
}

func (e *Emplacement) Tonnage() float64 {
	return 0.1
}

func (e *Emplacement) MaxArmorPoints() float64 {
	return e.Resource.Armor
}

func (e *Emplacement) MaxStructurePoints() float64 {
	return e.Resource.Structure
}

func (e *Emplacement) Heat() float64 {
	return 0
}

func (e *Emplacement) HeatDissipation() float64 {
	return 0
}

func (e *Emplacement) TriggerWeapon(w Weapon) bool {
	if w.Cooldown() > 0 {
		return false
	}
	w.TriggerCooldown()
	return true
}

func (e *Emplacement) TurnRate() float64 {
	return e.maxTurnRate
}

func (e *Emplacement) Update() bool {
	isOverHeated := e.OverHeated()
	if e.powered == POWER_ON {
		// if heat is too high, auto shutdown
		if isOverHeated {
			e.SetPowered(POWER_OFF_HEAT)
		}
	} else {
		switch {
		case isOverHeated:
			// continue cooling down
			break

		case e.powered == POWER_OFF_HEAT && !isOverHeated:
			// set power on automatically after overheat status is cleared
			e.SetPowered(POWER_ON)
		}
	}

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
