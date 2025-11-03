package model

import (
	"fmt"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
)

const (
	VTOL_TURN_RATE_FACTOR float64 = (0.25 * geom.Pi) / TICKS_PER_SECOND
)

type VTOL struct {
	*UnitModel
	Resource *ModelVTOLResource
}

func NewVTOL(r *ModelVTOLResource) *VTOL {
	m := &VTOL{
		Resource: r,
		UnitModel: &UnitModel{
			name:          r.Name,
			variant:       r.Variant,
			unitType:      VTOLUnitType,
			anchor:        raycaster.AnchorCenter,
			armor:         r.Armor,
			structure:     r.Structure,
			heatSinks:     r.HeatSinks.Quantity,
			heatSinkType:  r.HeatSinks.Type.HeatSinkType,
			armament:      make([]Weapon, 0),
			ammunition:    NewAmmoStock(),
			maxVelocity:   r.Speed * KPH_TO_VELOCITY,
			maxTurnRate:   VTOL_TURN_RATE_FACTOR + (100 / r.Tonnage * VTOL_TURN_RATE_FACTOR),
			maxTurretRate: VTOL_TURN_RATE_FACTOR + (100 / r.Tonnage * VTOL_TURN_RATE_FACTOR),
			jumpJets:      0,
			powered:       POWER_ON, // TODO: define initial power status or power on event in mission resource
		},
	}

	// calculate heat dissipation per tick
	m.heatDissipation = SECONDS_PER_TICK / 4 * float64(m.heatSinks) * float64(m.heatSinkType)

	// need to use the image size to find the unit collision conversion from pixels
	vtolRelPath := fmt.Sprintf("%s/%s", VTOLResourceType, r.Image)
	vtolImg := resources.GetSpriteFromFile(vtolRelPath)
	width, height := vtolImg.Bounds().Dx(), vtolImg.Bounds().Dy()
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

func (e *VTOL) Tonnage() float64 {
	return e.Resource.Tonnage
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

	// position update needed
	return true
}
