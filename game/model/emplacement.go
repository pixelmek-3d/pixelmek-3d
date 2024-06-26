package model

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type Emplacement struct {
	*UnitModel
	Resource *ModelEmplacementResource
}

func NewEmplacement(r *ModelEmplacementResource, collisionRadius, collisionHeight float64, cockpitOffset *geom.Vector2) *Emplacement {
	m := &Emplacement{
		Resource: r,
		UnitModel: &UnitModel{
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
			maxTurnRate:     0.05, // FIXME: testing
			maxTurretRate:   0.05, // FIXME: testing
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

func (e *Emplacement) Name() string {
	return e.Resource.Name
}

func (e *Emplacement) Variant() string {
	return e.Resource.Variant
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
