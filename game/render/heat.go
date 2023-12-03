package render

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tinne26/etxt"
)

var (
	_colorHeatHot  = color.NRGBA{R: 225, G: 0, B: 0, A: 255}
	_colorHeatWarm = color.NRGBA{R: 255, G: 205, B: 0, A: 255}
	_colorHeatCool = color.NRGBA{R: 0, G: 155, B: 255, A: 255}
	_colorHeatText = _colorDefaultGreen
)

type HeatIndicator struct {
	HUDSprite
	fontRenderer *etxt.Renderer
	heat         float64
	maxHeat      float64
	dissipation  float64
}

// NewHeatIndicator creates a heat indicator image to be rendered on demand
func NewHeatIndicator(font *Font) *HeatIndicator {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.Top, etxt.Left)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	h := &HeatIndicator{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
	}

	return h
}

func (h *HeatIndicator) updateFontSize(width, height int) {
	// set font size based on element size
	pxSize := float64(height) / 3
	if pxSize < 1 {
		pxSize = 1
	}

	h.fontRenderer.SetSizePx(int(pxSize))
}

func (h *HeatIndicator) SetValues(heat, maxHeat, dissipation float64) {
	h.heat = heat
	h.maxHeat = maxHeat
	h.dissipation = dissipation
}

func (h *HeatIndicator) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen
	h.fontRenderer.SetTarget(screen)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	h.updateFontSize(bW, bH)

	midX := float32(bX) + float32(bW)/2

	// current heat level box
	heatRatio := float32(h.heat / h.maxHeat)
	if heatRatio > 1 {
		heatRatio = 1
	}
	hW, hH := heatRatio*float32(bW), float32(bH)/2
	hX, hY := midX-hW/2, float32(bY)

	hColor := hudOpts.HudColor(_colorHeatCool)
	if heatRatio > 0.7 {
		hColor = hudOpts.HudColor(_colorHeatHot)
	} else if heatRatio > 0.35 {
		hColor = hudOpts.HudColor(_colorHeatWarm)
	}

	vector.DrawFilledRect(screen, hX, hY, hW, hH, hColor, false)

	// TODO: make current heat level box appear to flash when near/over maxHeat?

	// heat indicator outline
	oAlpha := uint8(4 * (int(hColor.A) / 5))
	oColor := color.NRGBA{hColor.R, hColor.G, hColor.B, oAlpha}

	var oT float32 = 2 // TODO: calculate line thickness based on image height
	oX, oY, oW, oH := float32(bX), float32(bY), float32(bW), float32(bH)/2
	vector.StrokeRect(screen, oX, oY, oW, oH, oT, oColor, false)

	// current heat level text
	tColor := hudOpts.HudColor(_colorHeatText)
	h.fontRenderer.SetColor(color.RGBA(tColor))

	heatStr := fmt.Sprintf("Heat: %0.1f", h.heat)
	h.fontRenderer.SetAlign(etxt.Top, etxt.Left)
	h.fontRenderer.Draw(heatStr, int(oX+2*oT), int(oY+oH+2*oT)) // TODO: calculate better margin spacing

	// current heat dissipation text
	dissipationStr := fmt.Sprintf("dH/dT: %0.1f", -h.dissipation)
	h.fontRenderer.SetAlign(etxt.Top, etxt.Right)
	h.fontRenderer.Draw(dissipationStr, int(oX+oW-2*oT), int(oY+oH+2*oT)) // TODO: calculate better margin spacing
}
