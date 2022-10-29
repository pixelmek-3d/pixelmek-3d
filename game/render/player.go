package render

import (
	"image/color"
	"math"

	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
)

type Player struct {
	model.Entity
	CameraZ             float64
	Moved               bool
	MapColor            color.RGBA
	ConvergenceDistance float64
	ConvergencePoint    *geom3d.Vector3
}

func NewPlayer(unit model.Entity, x, y, angle, pitch float64) *Player {
	p := &Player{
		Entity:   unit,
		CameraZ:  unit.CockpitOffset().Y, // TODO: support cockpit offset in sprite X direction
		Moved:    false,
		MapColor: color.RGBA{255, 0, 0, 255},
	}

	p.SetAsPlayer(true)

	p.SetPos(&geom.Vector2{X: x, Y: y})
	p.SetPosZ(0)
	p.SetAngle(angle)
	p.SetPitch(pitch)
	p.SetVelocity(0)
	p.SetArmorPoints(math.MaxFloat64) // TODO: get from mech model
	p.SetStructurePoints(math.MaxFloat64)

	return p
}
