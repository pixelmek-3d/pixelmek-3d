package model

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
)

type Crosshairs struct {
	*Sprite
}

func NewCrosshairs(
	scale float64, img *ebiten.Image, columns, rows, crosshairIndex int,
) *Crosshairs {
	c := &Crosshairs{
		Sprite: NewSpriteFromSheet(0, 0, scale, img, color.RGBA{}, columns, rows, crosshairIndex, raycaster.AnchorCenter, 0, 0),
	}

	return c
}
