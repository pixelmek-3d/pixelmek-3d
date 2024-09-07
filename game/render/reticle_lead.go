package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
)

type ReticleLead struct {
	*Sprite
}

func NewReticleLead(pos geom3d.Vector3) *ReticleLead {

	// TODO: make it as a 1x1 empty image since it will be a guide for the HUD
	leadImg := ebiten.NewImage(10, 10)
	vector.StrokeLine(leadImg, 0, 0, 9, 9, 1, color.White, false)
	vector.StrokeLine(leadImg, 9, 0, 0, 9, 1, color.White, false)

	leadEntity := model.BasicVisualEntity(pos.X, pos.Y, pos.Z, raycaster.AnchorCenter)
	r := &ReticleLead{
		Sprite: NewSprite(leadEntity, 0.25, leadImg),
	}

	// reticle lead cannot be focused upon by player reticle
	r.focusable = false

	// TODO: remove illumination since it is intended to be invisible
	r.illumination = 5000

	return r
}

func (r *ReticleLead) SetPosition(pos geom3d.Vector3) {
	r.Pos().X = pos.X
	r.Pos().Y = pos.Y
	r.SetPosZ(pos.Z)
}
