package model

import (
	"fmt"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
)

const (
	VEHICLE_TURN_RATE_FACTOR   float64 = (0.125 * geom.Pi) / TICKS_PER_SECOND
	VEHICLE_TURRET_RATE_FACTOR float64 = 2.0 * VEHICLE_TURN_RATE_FACTOR
)

type Vehicle struct {
	*UnitModel
	Resource *ModelVehicleResource
}

func NewVehicle(r *ModelVehicleResource) *Vehicle {
	m := &Vehicle{
		Resource: r,
		UnitModel: &UnitModel{
			name:            r.Name,
			variant:         r.Variant,
			unitType:        VehicleUnitType,
			anchor:          raycaster.AnchorBottom,
			armor:           r.Armor,
			structure:       r.Structure,
			heatSinks:       r.HeatSinks.Quantity,
			heatSinkType:    r.HeatSinks.Type.HeatSinkType,
			armament:        make([]Weapon, 0),
			ammunition:      NewAmmoStock(),
			hasTurret:       true,
			maxTurretExtent: 2.0 * geom.Pi,
			maxVelocity:     r.Speed * KPH_TO_VELOCITY,
			maxTurnRate:     VEHICLE_TURN_RATE_FACTOR + (100 / r.Tonnage * VEHICLE_TURN_RATE_FACTOR),
			maxTurretRate:   VEHICLE_TURRET_RATE_FACTOR + (100 / r.Tonnage * VEHICLE_TURRET_RATE_FACTOR),
			jumpJets:        0,
			powered:         POWER_ON, // TODO: define initial power status or power on event in mission resource
		},
	}

	// calculate heat dissipation per tick
	m.heatDissipation = SECONDS_PER_TICK / 4 * float64(m.heatSinks) * float64(m.heatSinkType)

	// need to use the image size to find the unit collision conversion from pixels
	vehicleRelPath := fmt.Sprintf("%s/%s", VehicleResourceType, r.Image)
	vehicleImg := resources.GetSpriteFromFile(vehicleRelPath)
	width, height := vehicleImg.Bounds().Dx(), vehicleImg.Bounds().Dy()
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

	// position update needed
	return true
}
