package render

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type TargetReticle struct {
	HUDSprite
}

//NewTargetReticle creates a target reticle from an image with 2 rows and 2 columns, representing the four corners of it
func NewTargetReticle(scale float64, img *ebiten.Image) *TargetReticle {
	r := &TargetReticle{
		HUDSprite: NewHUDSpriteFromSheet(img, scale, 2, 2, 0),
	}

	return r
}
