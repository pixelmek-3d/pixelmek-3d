package render

import (
	"image"
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

func (c *Crosshairs) Draw(screen *ebiten.Image, bounds image.Rectangle, clr *color.RGBA) {
	bX, bY, bW := bounds.Min.X, bounds.Min.Y, bounds.Dx()

	cScale := float64(bW) / float64(c.Width())

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.ColorM.ScaleWithColor(clr)

	op.GeoM.Scale(cScale, cScale)
	op.GeoM.Translate(float64(bX), float64(bY))
	screen.DrawImage(c.Texture(), op)
}
