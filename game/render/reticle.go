package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
)

type TargetReticle struct {
	*Sprite
}

//NewTargetReticle creates a target reticle from an image with 2 rows and 2 columns, representing the four corners of it
func NewTargetReticle(scale float64, img *ebiten.Image) *TargetReticle {
	r := &TargetReticle{
		Sprite: NewSpriteFromSheet(0, 0, scale, img, color.RGBA{}, 2, 2, 0, raycaster.AnchorCenter, 0, 0),
	}

	return r
}
