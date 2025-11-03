package model

import (
	"fmt"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
)

const (
	INFANTRY_TURN_RATE_FACTOR   float64 = (0.5 * geom.Pi) / TICKS_PER_SECOND
	INFANTRY_TURRET_RATE_FACTOR float64 = 2.0 * INFANTRY_TURN_RATE_FACTOR
)

type Infantry struct {
	*UnitModel
	Resource *ModelInfantryResource
}

func NewInfantry(r *ModelInfantryResource) *Infantry {
	m := &Infantry{
		Resource: r,
		UnitModel: &UnitModel{
			name:               r.Name,
			variant:            r.Variant,
			unitType:           InfantryUnitType,
			anchor:             raycaster.AnchorBottom,
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

	// need to use the image size to find the unit collision conversion from pixels
	infantryRelPath := fmt.Sprintf("%s/%s", InfantryResourceType, r.Image)
	infantryImg := resources.GetSpriteFromFile(infantryRelPath)
	width, height := infantryImg.Bounds().Dx(), infantryImg.Bounds().Dy()
	// handle if image has multiple rows/cols
	if r.ImageSheet != nil {
		width = int(float64(width) / float64(r.ImageSheet.Columns))
		height = int(float64(height) / float64(r.ImageSheet.Rows))
	}
	scale := ConvertHeightToScale(r.Height, height, r.HeightPxGap)
	collisionRadius, collisionHeight := ConvertOffsetFromPx(
		r.CollisionPxRadius, r.CollisionPxHeight, width, height, scale,
	)
	cockpitPxX, cockpitPxY := r.CockpitPxOffset[0], r.CockpitPxOffset[1]
	cockpitOffX, cockpitOffY := ConvertOffsetFromPx(cockpitPxX, cockpitPxY, width, height, scale)

	m.pxWidth, m.pxHeight = width, height
	m.pxScale = scale
	m.collisionRadius = collisionRadius
	m.collisionHeight = collisionHeight
	m.cockpitOffset = &geom.Vector2{X: cockpitOffX, Y: cockpitOffY}

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
