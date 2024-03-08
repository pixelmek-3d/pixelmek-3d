package render

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/tinne26/etxt"
)

var (
	// define default colors
	_colorAltimeter      = color.NRGBA{R: 0, G: 255, B: 67, A: 255}
	_colorAltimeterPips  = _colorDefaultGreen
	_colorAltimeterPitch = color.NRGBA{R: 0, G: 127, B: 0, A: 255}
)

type Altimeter struct {
	HUDSprite
	fontRenderer *etxt.Renderer
	altitude     float64
	pitch        float64
}

// NewAltimeter creates a compass image to be rendered on demand
func NewAltimeter(font *Font) *Altimeter {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.YCenter, etxt.Right)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	a := &Altimeter{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
	}

	return a
}

func (a *Altimeter) updateFontSize(_, height int) {
	// set font size based on element size
	pxSize := float64(height) / 14
	if pxSize < 1 {
		pxSize = 1
	}

	a.fontRenderer.SetSizePx(int(pxSize))
}

func (a *Altimeter) SetValues(altitude, pitch float64) {
	a.altitude = altitude
	a.pitch = pitch
}

func (a *Altimeter) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen
	a.fontRenderer.SetTarget(screen)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	a.updateFontSize(bW, bH)

	// use opposite pitch value so indicator will draw upward from center when postive angle
	relPitchAngle := -a.pitch
	relPitchDeg := geom.Degrees(relPitchAngle)

	midX, midY := float32(bX)+float32(bW)/2, float32(bY)+float32(bH)/2

	// pitch indicator box
	pitchColor := hudOpts.HudColor(_colorAltimeterPitch)

	var maxPitchDeg float64 = 45
	pitchRatio := relPitchDeg / maxPitchDeg
	tW, tH := float32(bW)/4, float32(pitchRatio)*float32(bH/2)
	pAlpha := uint8(4 * int(pitchColor.A) / 5)
	vector.DrawFilledRect(screen, midX, midY, tW, tH, color.NRGBA{pitchColor.R, pitchColor.G, pitchColor.B, pAlpha}, false)

	// altimeter pips
	pipColor := hudOpts.HudColor(_colorAltimeterPips)
	a.fontRenderer.SetColor(color.RGBA(pipColor))

	var maxAltitude float32 = float32(model.METERS_PER_UNIT)
	for i := int(-maxAltitude); i <= int(maxAltitude); i++ {
		actualAlt := i + int(math.Round(a.altitude))

		var pipWidth, pipHeight float32
		if actualAlt%5 == 0 {
			pipWidth = float32(bW) / 4
			pipHeight = 2
		}
		if actualAlt%10 == 0 {
			pipWidth = float32(bW) / 2
			pipHeight = 3
		}

		if pipWidth > 0 {
			// pip shows relative based on index (i) where negative is above center, positive is below
			iRatio := float32(-i) / maxAltitude
			iY := float32(bY) + float32(bH)/2 + iRatio*float32(bH)/2
			vector.DrawFilledRect(screen, midX, iY-pipHeight/2, pipWidth, pipHeight, pipColor, false)

			var pipAltStr string = fmt.Sprintf("%d", actualAlt)

			if pipAltStr != "" {
				a.fontRenderer.Draw(pipAltStr, int(midX), int(iY))
			}
		}
	}

	// altitude indicator line
	altColor := hudOpts.HudColor(_colorAltimeter)

	hW, hH := 2*float32(bW)/3, float32(5.0) // TODO: calculate line thickness based on image height
	vector.DrawFilledRect(screen, midX, midY-hH/2, hW, hH, altColor, false)
}
