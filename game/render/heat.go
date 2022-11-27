package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tinne26/etxt"
)

type HeatIndicator struct {
	HUDSprite
	image        *ebiten.Image
	fontRenderer *etxt.Renderer
}

//NewHeatIndicator creates a heat indicator image to be rendered on demand
func NewHeatIndicator(width, height int, font *Font) *HeatIndicator {
	img := ebiten.NewImage(width, height)

	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetSizePx(16)
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.Top, etxt.Left)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	h := &HeatIndicator{
		HUDSprite:    NewHUDSprite(img, 1.0),
		image:        img,
		fontRenderer: renderer,
	}

	return h
}

func (h *HeatIndicator) Update(heat, maxHeat, dissipation float64) {
	h.image.Clear()

	h.fontRenderer.SetTarget(h.image)

	midX := float64(h.Width()) / 2

	// current heat level box
	heatRatio := heat / maxHeat
	if heatRatio > 1 {
		heatRatio = 1
	}
	hW, hH := heatRatio*float64(h.Width()), float64(h.Height())/2
	hX, hY := midX-hW/2, 0.0
	ebitenutil.DrawRect(h.image, hX, hY, hW, hH, color.RGBA{255, 255, 255, 160})

	// TODO: make current heat level box appear to flash when near/over maxHeat?

	// heat indicator outline
	// FIXME: when ebitengine v2.5 releases can draw rect outline using StrokeRect
	//        - import "github.com/hajimehoshi/ebiten/v2/vector"
	//        - StrokeRect(dst *ebiten.Image, x, y, width, height float32, strokeWidth float32, clr color.Color)
	var oT float64 = 2 // TODO: calculate line thickness based on image height
	oX, oY, oW, oH := 0.0, 0.0, float64(h.Width()), float64(h.Height())/2
	ebitenutil.DrawRect(h.image, oX, oY, oW, oT, color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(h.image, oX+oW-oT, oY, oT, oH, color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(h.image, oX, oY+oH-oT, oW, oT, color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(h.image, oX, oY, oT, oH, color.RGBA{255, 255, 255, 255})

	// current heat level text
	heatStr := fmt.Sprintf("Heat: %0.1f", heat)
	h.fontRenderer.SetAlign(etxt.Top, etxt.Left)
	h.fontRenderer.Draw(heatStr, int(oX+2*oT), int(oY+oH+2*oT)) // TODO: calculate better margin spacing

	// current heat dissipation text
	dissipationStr := fmt.Sprintf("dH/dT: %0.1f", -dissipation)
	h.fontRenderer.SetAlign(etxt.Top, etxt.Right)
	h.fontRenderer.Draw(dissipationStr, int(oX+oW-2*oT), int(oY+oH+2*oT)) // TODO: calculate better margin spacing
}

func (h *HeatIndicator) Texture() *ebiten.Image {
	return h.image
}
