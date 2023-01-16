package render

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
)

var (
	_colorHeatHot  = color.RGBA{R: 225, G: 0, B: 0, A: 255}
	_colorHeatWarm = color.RGBA{R: 255, G: 205, B: 0, A: 255}
	_colorHeatCool = color.RGBA{R: 0, G: 155, B: 255, A: 255}
	_colorHeatText = _colorDefaultGreen
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

	var hColor color.RGBA
	if hudOpts.UseCustomColor {
		hColor = hudOpts.Color
	} else {
		hColor = _colorHeatCool
		if heatRatio > 0.7 {
			hColor = _colorHeatHot
		} else if heatRatio > 0.35 {
			hColor = _colorHeatWarm
		}
	}

	ebitenutil.DrawRect(screen, hX, hY, hW, hH, hColor)

	// TODO: make current heat level box appear to flash when near/over maxHeat?

	// heat indicator outline
	oAlpha := uint8(4 * (int(hColor.A) / 5))
	oColor := color.RGBA{hColor.R, hColor.G, hColor.B, oAlpha}

	// FIXME: when ebitengine v2.5 releases can draw rect outline using StrokeRect
	//        - import "github.com/hajimehoshi/ebiten/v2/vector"
	//        - StrokeRect(dst *ebiten.Image, x, y, width, height float32, strokeWidth float32, hudOpts.Color color.Color)
	var oT float64 = 2 // TODO: calculate line thickness based on image height
	oX, oY, oW, oH := float64(bX), float64(bY), float64(bW), float64(bH)/2
	ebitenutil.DrawRect(screen, oX, oY, oW, oT, oColor)
	ebitenutil.DrawRect(screen, oX+oW-oT, oY, oT, oH, oColor)
	ebitenutil.DrawRect(screen, oX, oY+oH-oT, oW, oT, oColor)
	ebitenutil.DrawRect(screen, oX, oY, oT, oH, oColor)

	// current heat level text
	tColor := _colorHeatText
	if hudOpts.UseCustomColor {
		tColor = hudOpts.Color
	}
	h.fontRenderer.SetColor(tColor)

	heatStr := fmt.Sprintf("Heat: %0.1f", heat)
	h.fontRenderer.SetAlign(etxt.Top, etxt.Left)
	h.fontRenderer.Draw(heatStr, int(oX+2*oT), int(oY+oH+2*oT)) // TODO: calculate better margin spacing

	// current heat dissipation text
	dissipationStr := fmt.Sprintf("dH/dT: %0.1f", -dissipation)
	h.fontRenderer.SetAlign(etxt.Top, etxt.Right)
	h.fontRenderer.Draw(dissipationStr, int(oX+oW-2*oT), int(oY+oH+2*oT)) // TODO: calculate better margin spacing
}
