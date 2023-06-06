package game

import (
	"fmt"

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

func (g *Game) SetPlayerUnit(unitType, unitResource string) model.Unit {
	var unit model.Unit
	var unitSprite *render.Sprite

	switch unitType {
	case model.MechResourceType:
		unit = g.createModelMech(unitResource)
		unitSprite = g.createUnitSprite(unit).(*render.MechSprite).Sprite

	case model.VehicleResourceType:
		vUnit := g.createModelVehicle(unitResource)
		unit = vUnit

		unitImgPath := fmt.Sprintf("%s/%s", unitType, vUnit.Resource.Image)
		unitImg := getSpriteFromFile(unitImgPath)
		scale := convertHeightToScale(vUnit.Resource.Height, vUnit.Resource.HeightPxRatio)
		unitSprite = render.NewVehicleSprite(vUnit, scale, unitImg).Sprite

	case model.VTOLResourceType:
		vUnit := g.createModelVTOL(unitResource)
		unit = vUnit

		unitImgPath := fmt.Sprintf("%s/%s", unitType, vUnit.Resource.Image)
		unitImg := getSpriteFromFile(unitImgPath)
		scale := convertHeightToScale(vUnit.Resource.Height, vUnit.Resource.HeightPxRatio)
		unitSprite = render.NewVTOLSprite(vUnit, scale, unitImg).Sprite

	case model.InfantryResourceType:
		iUnit := g.createModelInfantry(unitResource)
		unit = iUnit

		unitImgPath := fmt.Sprintf("%s/%s", unitType, iUnit.Resource.Image)
		unitImg := getSpriteFromFile(unitImgPath)
		scale := convertHeightToScale(iUnit.Resource.Height, iUnit.Resource.HeightPxRatio)
		unitSprite = render.NewInfantrySprite(iUnit, scale, unitImg).Sprite

	default:
		log.Fatalf("unable to set player unit, resource type %s does not exist", unitType)
		return nil
	}

	if unit == nil {
		log.Fatalf("unable to set player unit, resource does not exist %s/%s", unitType, unitResource)
		return nil
	}

	var pX, pY, pZ, pH float64
	if g.player != nil {
		pX, pY, pZ, pH = g.player.Pos().X, g.player.Pos().Y, g.player.PosZ(), g.player.Heading()
	}

	if unitType == model.VTOLResourceType {
		if pZ < unit.CollisionHeight() {
			// for VTOL, adjust Z position to not be stuck in the ground
			pZ = unit.CollisionHeight()
		}
	} else {
		// adjust Z position to be on the ground
		pZ = 0
	}

	g.player = NewPlayer(unit, unitSprite, pX, pY, pZ, pH, 0)
	g.player.SetCollisionRadius(unit.CollisionRadius())
	g.player.SetCollisionHeight(unit.CollisionHeight())
	g.armament.SetWeapons(g.player.Armament())

	if unit.HasTurret() {
		g.mouseMode = MouseModeTurret
	} else {
		g.mouseMode = MouseModeBody
	}

	return unit
}
