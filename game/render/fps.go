package render

import (
	"image"

	"github.com/pixelmek-3d/pixelmek-3d/game/render/fonts"
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
func NewFPSIndicator(font *fonts.Font) *FPSIndicator {
	// create and configure font renderer
	renderer := etxt.NewRenderer()
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

	f.fontRenderer.SetSize(pxSize)
}

func (f *FPSIndicator) SetFPSText(fpsText string) {
	f.fpsText = fpsText
}

func (f *FPSIndicator) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	f.updateFontSize(bW, bH)

	// fps indicator text
	tColor := hudOpts.HudColor(_colorFPSText)
	f.fontRenderer.SetColor(tColor)
	f.fontRenderer.SetAlign(etxt.Top | etxt.Left)
	f.fontRenderer.Draw(screen, f.fpsText, bX, bY)
}
