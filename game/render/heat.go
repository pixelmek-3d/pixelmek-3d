package render

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tinne26/etxt"
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
	renderer.SetSizePx(16)
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.Top, etxt.Left)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	h := &HeatIndicator{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
	}

	return h
}

func (h *HeatIndicator) Draw(screen *ebiten.Image, bounds image.Rectangle, clr *color.RGBA, heat, maxHeat, dissipation float64) {
	h.fontRenderer.SetTarget(screen)
	h.fontRenderer.SetColor(clr)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()

	midX := float64(bX) + float64(bW)/2

	// current heat level box
	heatRatio := heat / maxHeat
	if heatRatio > 1 {
		heatRatio = 1
	}
	hW, hH := heatRatio*float64(bW), float64(bH)/2
	hX, hY := midX-hW/2, float64(bY)
	hAlpha := uint8(4 * (int(clr.A) / 5))
	ebitenutil.DrawRect(screen, hX, hY, hW, hH, color.RGBA{clr.R, clr.G, clr.B, hAlpha})

	// TODO: make current heat level box appear to flash when near/over maxHeat?

	// heat indicator outline
	// FIXME: when ebitengine v2.5 releases can draw rect outline using StrokeRect
	//        - import "github.com/hajimehoshi/ebiten/v2/vector"
	//        - StrokeRect(dst *ebiten.Image, x, y, width, height float32, strokeWidth float32, clr color.Color)
	var oT float64 = 2 // TODO: calculate line thickness based on image height
	oX, oY, oW, oH := float64(bX), float64(bY), float64(bW), float64(bH)/2
	ebitenutil.DrawRect(screen, oX, oY, oW, oT, clr)
	ebitenutil.DrawRect(screen, oX+oW-oT, oY, oT, oH, clr)
	ebitenutil.DrawRect(screen, oX, oY+oH-oT, oW, oT, clr)
	ebitenutil.DrawRect(screen, oX, oY, oT, oH, clr)

	// current heat level text
	heatStr := fmt.Sprintf("Heat: %0.1f", heat)
	h.fontRenderer.SetAlign(etxt.Top, etxt.Left)
	h.fontRenderer.Draw(heatStr, int(oX+2*oT), int(oY+oH+2*oT)) // TODO: calculate better margin spacing

	// current heat dissipation text
	dissipationStr := fmt.Sprintf("dH/dT: %0.1f", -dissipation)
	h.fontRenderer.SetAlign(etxt.Top, etxt.Right)
	h.fontRenderer.Draw(dissipationStr, int(oX+oW-2*oT), int(oY+oH+2*oT)) // TODO: calculate better margin spacing
}
