package render

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
)

type HeatIndicator struct {
	HUDSprite
	fontRenderer *etxt.Renderer
}

//NewHeatIndicator creates a heat indicator image to be rendered on demand
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

	fractSize, _ := efixed.FromFloat64(pxSize)
	h.fontRenderer.SetSizePxFract(fractSize)
}

func (h *HeatIndicator) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions, heat, maxHeat, dissipation float64) {
	screen := hudOpts.Screen
	h.fontRenderer.SetTarget(screen)
	h.fontRenderer.SetColor(hudOpts.Color)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	h.updateFontSize(bW, bH)

	midX := float64(bX) + float64(bW)/2

	// current heat level box
	heatRatio := heat / maxHeat
	if heatRatio > 1 {
		heatRatio = 1
	}
	hW, hH := heatRatio*float64(bW), float64(bH)/2
	hX, hY := midX-hW/2, float64(bY)
	hAlpha := uint8(4 * (int(hudOpts.Color.A) / 5))
	ebitenutil.DrawRect(screen, hX, hY, hW, hH, color.RGBA{hudOpts.Color.R, hudOpts.Color.G, hudOpts.Color.B, hAlpha})

	// TODO: make current heat level box appear to flash when near/over maxHeat?

	// heat indicator outline
	// FIXME: when ebitengine v2.5 releases can draw rect outline using StrokeRect
	//        - import "github.com/hajimehoshi/ebiten/v2/vector"
	//        - StrokeRect(dst *ebiten.Image, x, y, width, height float32, strokeWidth float32, hudOpts.Color color.Color)
	var oT float64 = 2 // TODO: calculate line thickness based on image height
	oX, oY, oW, oH := float64(bX), float64(bY), float64(bW), float64(bH)/2
	ebitenutil.DrawRect(screen, oX, oY, oW, oT, hudOpts.Color)
	ebitenutil.DrawRect(screen, oX+oW-oT, oY, oT, oH, hudOpts.Color)
	ebitenutil.DrawRect(screen, oX, oY+oH-oT, oW, oT, hudOpts.Color)
	ebitenutil.DrawRect(screen, oX, oY, oT, oH, hudOpts.Color)

	// current heat level text
	heatStr := fmt.Sprintf("Heat: %0.1f", heat)
	h.fontRenderer.SetAlign(etxt.Top, etxt.Left)
	h.fontRenderer.Draw(heatStr, int(oX+2*oT), int(oY+oH+2*oT)) // TODO: calculate better margin spacing

	// current heat dissipation text
	dissipationStr := fmt.Sprintf("dH/dT: %0.1f", -dissipation)
	h.fontRenderer.SetAlign(etxt.Top, etxt.Right)
	h.fontRenderer.Draw(dissipationStr, int(oX+oW-2*oT), int(oY+oH+2*oT)) // TODO: calculate better margin spacing
}
