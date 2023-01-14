package render

import (
	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
)

type Player struct {
	model.Unit
	Sprite              *Sprite
	CameraZ             float64
	Moved               bool
	ConvergenceDistance float64
	ConvergencePoint    *geom3d.Vector3
}

func NewPlayer(unit model.Unit, sprite *Sprite, x, y, z, angle, pitch float64) *Player {
	p := &Player{
		Unit:    unit,
		Sprite:  sprite,
		CameraZ: z + unit.CockpitOffset().Y, // TODO: support cockpit offset in sprite X direction
		Moved:   false,
	}

	p.SetAsPlayer(true)

	p.SetPos(&geom.Vector2{X: x, Y: y})
	p.SetPosZ(z)
	p.SetHeading(angle)
	p.SetPitch(pitch)
	p.SetVelocity(0)

	return p
}
