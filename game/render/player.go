package render

import (
	"image/color"
	"math"

	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
)

type Player struct {
	model.Unit
	CameraZ             float64
	Moved               bool
	MapColor            color.RGBA
	ConvergenceDistance float64
	ConvergencePoint    *geom3d.Vector3
}

func NewPlayer(unit model.Unit, x, y, z, angle, pitch float64) *Player {
	p := &Player{
		Unit:     unit,
		CameraZ:  z + unit.CockpitOffset().Y, // TODO: support cockpit offset in sprite X direction
		Moved:    false,
		MapColor: color.RGBA{255, 0, 0, 255},
	}

	p.SetAsPlayer(true)

	p.SetPos(&geom.Vector2{X: x, Y: y})
	p.SetPosZ(z)
	p.SetHeading(angle)
	p.SetPitch(pitch)
	p.SetVelocity(0)
	p.SetArmorPoints(math.MaxFloat64) // TODO: get from mech model
	p.SetStructurePoints(math.MaxFloat64)

	return p
}
