package render

import (
	"image"

	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
)

var (
	_colorFPSText = _colorDefaultGreen
)

type FPSIndicator struct {
	HUDSprite
	fontRenderer *etxt.Renderer
	fpsText      string
}

//NewFPSIndicator creates an FPS indicator to be rendered on demand
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

func (f *FPSIndicator) updateFontSize(width, height int) {
	// set font size based on element size
	pxSize := float64(height)
	if pxSize < 1 {
		pxSize = 1
	}

	fractSize, _ := efixed.FromFloat64(pxSize)
	f.fontRenderer.SetSizePxFract(fractSize)
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
	tColor := _colorFPSText
	if hudOpts.UseCustomColor {
		tColor = hudOpts.Color
	}
	f.fontRenderer.SetColor(tColor)
	f.fontRenderer.SetAlign(etxt.Top, etxt.Left)
	f.fontRenderer.Draw(f.fpsText, bX, bY)
}
