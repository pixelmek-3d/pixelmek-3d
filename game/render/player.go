package render

import (
	"image/color"
	"math"

	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/harbdog/raycaster-go/geom"
)

type Player struct {
	model.Entity
	CameraZ        float64
	Moved          bool
	TestProjectile *ProjectileSprite
	TestCooldown   int
	MapColor       color.RGBA
}

func NewPlayer(x, y, angle, pitch float64) *Player {
	p := &Player{
		Entity:   &model.BasicEntity{},
		CameraZ:  0.5,
		Moved:    false,
		MapColor: color.RGBA{255, 0, 0, 255},
	}

	p.SetPosition(&geom.Vector2{X: x, Y: y})
	p.SetPositionZ(0)
	p.SetAngle(angle)
	p.SetPitch(pitch)
	p.SetVelocity(0)
	p.SetHitPoints(math.MaxFloat64) // TODO: get from mech model

	return p
}