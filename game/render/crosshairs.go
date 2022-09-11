package render

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Crosshairs struct {
	HUDSprite
}

func NewCrosshairs(
	img *ebiten.Image, scale float64, columns, rows, crosshairIndex int,
) *Crosshairs {
	c := &Crosshairs{
		HUDSprite: NewHUDSpriteFromSheet(img, scale, columns, rows, crosshairIndex),
	}

	return c
}
