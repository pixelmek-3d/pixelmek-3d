package render

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	_colorCrosshair = _colorDefaultGreen
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

func (c *Crosshairs) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen
	bX, bY, bW := bounds.Min.X, bounds.Min.Y, bounds.Dx()

	cScale := float64(bW) / float64(c.Width())

	cColor := _colorCrosshair
	if hudOpts.UseCustomColor {
		cColor = hudOpts.Color
	} else {
		cColor.A = hudOpts.Color.A
	}

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.ColorScale.ScaleWithColor(cColor)

	op.GeoM.Scale(cScale, cScale)
	op.GeoM.Translate(float64(bX), float64(bY))
	screen.DrawImage(c.Texture(), op)
}
