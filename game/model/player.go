package model

import (
	"image/color"
	"math"

	"github.com/harbdog/raycaster-go/geom"
)

type Player struct {
	Entity
	CameraZ        float64
	Moved          bool
	TestProjectile *Projectile
	TestCooldown   int
	MapColor       color.RGBA
}

func NewPlayer(x, y, angle, pitch float64) *Player {
	p := &Player{
		Entity: &BasicEntity{
			position:  &geom.Vector2{X: x, Y: y},
			positionZ: 0,
			angle:     angle,
			pitch:     pitch,
			velocity:  0,
			hitPoints: math.MaxFloat64,
		},
		CameraZ:  0.5,
		Moved:    false,
		MapColor: color.RGBA{255, 0, 0, 255},
	}

	return p
}
