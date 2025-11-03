package sprites

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
)

type ReticleLead struct {
	*Sprite
}

func NewReticleLead(pos geom3d.Vector3) *ReticleLead {
	// use smallest possible empty image to be a raycasted guide for the HUD
	leadImg := ebiten.NewImage(2, 2)
	leadEntity := model.BasicVisualEntity(pos.X, pos.Y, pos.Z, raycaster.AnchorCenter)
	r := &ReticleLead{
		Sprite: NewSprite(leadEntity, 0.25, leadImg),
	}

	// reticle lead cannot be focused upon by player reticle
	r.focusable = false

	return r
}

func (r *ReticleLead) SetPosition(pos geom3d.Vector3) {
	r.Pos().X = pos.X
	r.Pos().Y = pos.Y
	r.SetPosZ(pos.Z)
}
