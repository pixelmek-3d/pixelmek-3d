package render

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type TargetReticle struct {
	HUDSprite
}

type NavReticle struct {
	HUDSprite
}

//NewTargetReticle creates a target reticle from an image with 2 rows and 2 columns, representing the four corners of it
func NewTargetReticle(scale float64, img *ebiten.Image) *TargetReticle {
	r := &TargetReticle{
		HUDSprite: NewHUDSpriteFromSheet(img, scale, 2, 2, 0),
	}

	return r
}

//NewNavReticle creates a nav reticle from an image
func NewNavReticle(scale float64, img *ebiten.Image) *NavReticle {
	r := &NavReticle{
		HUDSprite: NewHUDSprite(img, scale),
	}

	return r
}

func (t *TargetReticle) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen

	// set minimum scale size based on screen size
	screenW, screenH := screen.Size()
	screenDim := int(float64(screenW) * hudOpts.RenderScale)
	if screenH > screenW {
		screenDim = int(float64(screenH) * hudOpts.RenderScale)
	}
	screenMinScale := float64(screenDim) / (50 * float64(t.Width()))

	// adjust scale based on size of rect target being placed around
	targetDim := bounds.Dx()
	if bounds.Dy() > targetDim {
		targetDim = bounds.Dy()
	}
	rScale := float64(targetDim) / (10 * float64(t.Width()))
	if rScale < screenMinScale {
		rScale = screenMinScale
	}
	rOff := rScale * float64(t.Width()) / 2

	minX, minY, maxX, maxY := float64(bounds.Min.X), float64(bounds.Min.Y), float64(bounds.Max.X), float64(bounds.Max.Y)

	rColor := _colorEnemy
	if hudOpts.UseCustomColor {
		rColor = hudOpts.Color
	}

	// setup some common draw modifications
	var op *ebiten.DrawImageOptions
	geoM := ebiten.GeoM{}
	geoM.Scale(rScale, rScale)
	colorM := ebiten.ColorM{}
	colorM.ScaleWithColor(rColor)

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

func (t *NavReticle) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen

	// set minimum scale size based on screen size
	screenW, screenH := screen.Size()
	screenDim := int(float64(screenW) * hudOpts.RenderScale)
	if screenH > screenW {
		screenDim = int(float64(screenH) * hudOpts.RenderScale)
	}
	screenMinScale := float64(screenDim) / (50 * float64(t.Width()))

	// adjust scale based on size of rect target being placed around
	targetDim := bounds.Dx()
	if bounds.Dy() > targetDim {
		targetDim = bounds.Dy()
	}
	rScale := float64(targetDim) / float64(t.Width())
	if rScale < screenMinScale {
		rScale = screenMinScale
	}

	minX, minY, dX, dY := float64(bounds.Min.X), float64(bounds.Min.Y), float64(bounds.Dx()), float64(bounds.Dy())

	rColor := _colorNavPoint
	if hudOpts.UseCustomColor {
		rColor = hudOpts.Color
	}

	// setup some common draw modifications
	var op *ebiten.DrawImageOptions
	geoM := ebiten.GeoM{}
	geoM.Scale(rScale, rScale)
	colorM := ebiten.ColorM{}
	colorM.ScaleWithColor(rColor)

	rX, rY := 1+minX+dX/2-rScale*float64(t.Width())/2, 1+minY+dY/2-rScale*float64(t.Height())/2

	op = &ebiten.DrawImageOptions{ColorM: colorM, GeoM: geoM}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(rX, rY)
	screen.DrawImage(t.Texture(), op)
}
