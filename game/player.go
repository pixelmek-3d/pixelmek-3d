package game

import (
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
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
	navPoint            *model.NavPoint
}

func NewPlayer(unit model.Unit, sprite *render.Sprite, x, y, z, angle, pitch float64) *Player {
	p := &Player{
		Unit:    unit,
		sprite:  sprite,
		cameraZ: z + unit.CockpitOffset().Y, // TODO: support cockpit offset in sprite X direction
		moved:   false,
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
