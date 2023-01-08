package render

import (
	"image"
	"image/color"

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

func (t *TargetReticle) Draw(screen *ebiten.Image, rect image.Rectangle, clr *color.RGBA) {
	// set minimum scale size based on screen size
	screenW, screenH := screen.Size()
	screenDim := screenW
	if screenH > screenW {
		screenDim = screenH
	}
	screenMinScale := float64(screenDim) / (50 * float64(t.Width()))

	// adjust scale based on size of rect target being placed around
	targetDim := rect.Dx()
	if rect.Dy() > targetDim {
		targetDim = rect.Dy()
	}
	rScale := float64(targetDim) / (10 * float64(t.Width()))
	if rScale < screenMinScale {
		rScale = screenMinScale
	}
	rOff := rScale * float64(t.Width()) / 2

	minX, minY, maxX, maxY := float64(rect.Min.X), float64(rect.Min.Y), float64(rect.Max.X), float64(rect.Max.Y)

	// setup some common draw modifications
	var op *ebiten.DrawImageOptions
	geoM := ebiten.GeoM{}
	geoM.Scale(rScale, rScale)
	colorM := ebiten.ColorM{}
	colorM.ScaleWithColor(clr)

	// top left corner
	t.SetTextureFrame(0)
	op = &ebiten.DrawImageOptions{ColorM: colorM, GeoM: geoM}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(minX-rOff, minY-rOff)
	screen.DrawImage(t.Texture(), op)

	// top right corner
	t.SetTextureFrame(1)
	op = &ebiten.DrawImageOptions{ColorM: colorM, GeoM: geoM}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(maxX-rOff, minY-rOff)
	screen.DrawImage(t.Texture(), op)

	// bottom left corner
	t.SetTextureFrame(2)
	op = &ebiten.DrawImageOptions{ColorM: colorM, GeoM: geoM}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(minX-rOff, maxY-rOff)
	screen.DrawImage(t.Texture(), op)

	// bottom right corner
	t.SetTextureFrame(3)
	op = &ebiten.DrawImageOptions{ColorM: colorM, GeoM: geoM}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(maxX-rOff, maxY-rOff)
	screen.DrawImage(t.Texture(), op)
}
