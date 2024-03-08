package render

import (
	"image"
	"image/color"

	"github.com/tinne26/etxt"
)

var (
	_colorFPSText = _colorDefaultGreen
)

type FPSIndicator struct {
	HUDSprite
	fontRenderer *etxt.Renderer
	fpsText      string
}

// NewFPSIndicator creates an FPS indicator to be rendered on demand
func NewFPSIndicator(font *Font) *FPSIndicator {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)

	f := &FPSIndicator{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
	}

	return f
}

func (f *FPSIndicator) updateFontSize(_, height int) {
	// set font size based on element size
	pxSize := float64(height)
	if pxSize < 1 {
		pxSize = 1
	}

	f.fontRenderer.SetSizePx(int(pxSize))
}

func (f *FPSIndicator) SetFPSText(fpsText string) {
	f.fpsText = fpsText
}

func (f *FPSIndicator) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen
	f.fontRenderer.SetTarget(screen)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	f.updateFontSize(bW, bH)

	// fps indicator text
	tColor := hudOpts.HudColor(_colorFPSText)
	f.fontRenderer.SetColor(color.RGBA(tColor))
	f.fontRenderer.SetAlign(etxt.Top, etxt.Left)
	f.fontRenderer.Draw(f.fpsText, bX, bY)
}
