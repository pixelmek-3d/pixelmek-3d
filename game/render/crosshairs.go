package render

import (
	"image/color"

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

func (c *Crosshairs) Draw(screen *ebiten.Image, scale float64, clr *color.RGBA) {

	screenW, screenH := screen.Size()

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.ColorM.ScaleWithColor(clr)

	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(
		float64(screenW)/2-scale*float64(c.Width())/2,
		float64(screenH)/2-scale*float64(c.Height())/2,
	)
	screen.DrawImage(c.Texture(), op)
}
