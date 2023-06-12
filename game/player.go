package game

import (
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"

	log "github.com/sirupsen/logrus"
)

type Player struct {
	model.Unit
	sprite              *render.Sprite
	cameraZ             float64
	moved               bool
	convergenceDistance float64
	convergencePoint    *geom3d.Vector3
	convergenceSprite   *render.Sprite
	weaponGroups        [][]model.Weapon
	selectedWeapon      uint
	selectedGroup       uint
	fireMode            model.WeaponFireMode
	navPoint            *render.NavSprite
}

func NewPlayer(unit model.Unit, sprite *render.Sprite, x, y, z, angle, pitch float64) *Player {
	p := &Player{
		Unit:   unit,
		sprite: sprite,
		moved:  false,
	}

	p.SetAsPlayer(true)

	p.SetPos(&geom.Vector2{X: x, Y: y})
	p.SetPosZ(z)
	p.SetHeading(angle)
	p.SetPitch(pitch)
	p.SetVelocity(0)

	p.selectedWeapon = 0
	p.weaponGroups = make([][]model.Weapon, 3)
	for i := 0; i < cap(p.weaponGroups); i++ {
		p.weaponGroups[i] = make([]model.Weapon, 0, len(unit.Armament()))
	}
	// initialize all weapons as only in first weapon group
	p.weaponGroups[0] = append(p.weaponGroups[0], unit.Armament()...)

	// TODO: save/restore weapon groups for weapons per unit

	return p
}

func (p *Player) SetPosZ(z float64) {
	p.cameraZ = z + p.Unit.CockpitOffset().Y // TODO: support cockpit offset in sprite X direction
	p.Unit.SetPosZ(z)
}

func (g *Game) setPlayerUnitFromResourceFile(resourceType, resourceFile string) model.Unit {
	var unit model.Unit

	switch resourceType {
	case model.MechResourceType:
		unit = g.createModelMech(resourceFile)

	case model.VehicleResourceType:
		unit = g.createModelVehicle(resourceFile)

	case model.VTOLResourceType:
		unit = g.createModelVTOL(resourceFile)

	case model.InfantryResourceType:
		unit = g.createModelInfantry(resourceFile)

	default:
		log.Fatalf("unable to set player unit, resource type %s not handled", resourceType)
		return nil
	}

	if unit == nil {
		log.Fatalf("unable to set player unit, resource does not exist %s/%s", resourceType, resourceFile)
		return nil
	}

	g.SetPlayerUnit(unit)
	return unit
}

func (g *Game) SetPlayerUnit(unit model.Unit) {
	var unitSprite *render.Sprite

	var pX, pY, pZ, pH float64
	if g.player != nil {
		// handle in-mission player unit changes
		pX, pY = g.player.Pos().X, g.player.Pos().Y
		pZ = 0.0
		pH = g.player.Heading()
	}

	switch unitType := unit.(type) {
	case *model.Mech:
		unitSprite = g.createUnitSprite(unit).(*render.MechSprite).Sprite

	case *model.Vehicle:
		unitSprite = g.createUnitSprite(unit).(*render.VehicleSprite).Sprite

	case *model.VTOL:
		unitSprite = g.createUnitSprite(unit).(*render.VTOLSprite).Sprite
		if pZ < unit.CollisionHeight() {
			// for VTOL, adjust Z position to not be stuck in the ground
			pZ = unit.CollisionHeight()
		}

	case *model.Infantry:
		unitSprite = g.createUnitSprite(unit).(*render.InfantrySprite).Sprite

	default:
		log.Fatalf("unable to set player unit, resource type %s not handled", unitType)
		return
	}

	g.player = NewPlayer(unit, unitSprite, pX, pY, pZ, pH, 0)
	g.player.SetCollisionRadius(unit.CollisionRadius())
	g.player.SetCollisionHeight(unit.CollisionHeight())

	if unit.HasTurret() {
		g.mouseMode = MouseModeTurret
	} else {
		g.mouseMode = MouseModeBody
	}
}
