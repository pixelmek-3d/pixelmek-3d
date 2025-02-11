package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/raycaster-go/geom"
)

type TargetReticle struct {
	HUDSprite
	Friendly          bool
	ReticleLeadBounds *image.Rectangle
}

type NavReticle struct {
	HUDSprite
}

// NewTargetReticle creates a target reticle from an image with 2 rows and 2 columns, representing the four corners of it
func NewTargetReticle(img *ebiten.Image) *TargetReticle {
	r := &TargetReticle{
		HUDSprite: NewHUDSpriteFromSheet(img, 1.0, 2, 2, 0),
	}

	return r
}

// NewNavReticle creates a nav reticle from an image
func NewNavReticle(img *ebiten.Image) *NavReticle {
	r := &NavReticle{
		HUDSprite: NewHUDSprite(img, 1.0),
	}

	return r
}

func (t *TargetReticle) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen

	// set minimum scale size based on screen size
	screenW, screenH := screen.Bounds().Dx(), screen.Bounds().Dy()
	screenDim := int(float64(screenW))
	if screenH > screenW {
		screenDim = int(float64(screenH))
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

	var rColor color.NRGBA
	if t.Friendly {
		// TODO: friendly reticle needs to look different in case custom HUD color is used
		rColor = hudOpts.HudColor(_colorFriendly)
	} else {
		rColor = hudOpts.HudColor(_colorEnemy)
	}

	// setup some common draw modifications
	var op *ebiten.DrawImageOptions
	geoM := ebiten.GeoM{}
	geoM.Scale(rScale, rScale)
	colorScale := ebiten.ColorScale{}
	colorScale.ScaleWithColor(rColor)

	// top left corner
	t.SetTextureFrame(0)
	op = &ebiten.DrawImageOptions{ColorScale: colorScale, GeoM: geoM}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(minX-rOff, minY-rOff)
	screen.DrawImage(t.Texture(), op)

	// top right corner
	t.SetTextureFrame(1)
	op = &ebiten.DrawImageOptions{ColorScale: colorScale, GeoM: geoM}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(maxX-rOff, minY-rOff)
	screen.DrawImage(t.Texture(), op)

	// bottom left corner
	t.SetTextureFrame(2)
	op = &ebiten.DrawImageOptions{ColorScale: colorScale, GeoM: geoM}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(minX-rOff, maxY-rOff)
	screen.DrawImage(t.Texture(), op)

	// bottom right corner
	t.SetTextureFrame(3)
	op = &ebiten.DrawImageOptions{ColorScale: colorScale, GeoM: geoM}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(maxX-rOff, maxY-rOff)
	screen.DrawImage(t.Texture(), op)

	// target lead reticle if outside target reticle bounds
	if t.ReticleLeadBounds != nil && !t.ReticleLeadBounds.In(bounds) {
		rlMidX := float32(t.ReticleLeadBounds.Min.X) + float32(t.ReticleLeadBounds.Dx())/2
		rlMidY := float32(t.ReticleLeadBounds.Min.Y) + float32(t.ReticleLeadBounds.Dy())/2
		rlRadius := float32(geom.Clamp(float64(targetDim)/4, 5, float64(screenDim)/50))

		rlColor := rColor
		if t.ReticleLeadBounds.Overlaps(bounds) {
			// make reticle lead mostly transparent if touching target reticle
			rlAlpha := uint8(2 * int(rColor.A) / 4)
			rlColor = color.NRGBA{rColor.R, rColor.G, rColor.B, rlAlpha}
		}

		vector.StrokeCircle(screen, rlMidX, rlMidY, rlRadius, 1, rlColor, false)
	}
}

func (t *NavReticle) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen

	// set minimum scale size based on screen size
	screenW, screenH := screen.Bounds().Dx(), screen.Bounds().Dy()
	screenDim := int(float64(screenW))
	if screenH > screenW {
		screenDim = int(float64(screenH))
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

	rColor := hudOpts.HudColor(_colorNavPoint)

	// setup some common draw modifications
	var op *ebiten.DrawImageOptions
	geoM := ebiten.GeoM{}
	geoM.Scale(rScale, rScale)
	colorScale := ebiten.ColorScale{}
	colorScale.ScaleWithColor(rColor)

	rX, rY := 1+minX+dX/2-rScale*float64(t.Width())/2, 1+minY+dY/2-rScale*float64(t.Height())/2

	op = &ebiten.DrawImageOptions{ColorScale: colorScale, GeoM: geoM}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(rX, rY)
	screen.DrawImage(t.Texture(), op)
}
