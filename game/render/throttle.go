package render

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
)

type Throttle struct {
	HUDSprite
	fontRenderer *etxt.Renderer
}

//NewThrottle creates a speed indicator image to be rendered on demand
func NewThrottle(font *Font) *Throttle {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.YCenter, etxt.Right)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	t := &Throttle{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
	}

	return t
}

func (t *Throttle) updateFontSize(width, height int) {
	// set font size based on element size
	pxSize := float64(height) / 18
	if pxSize < 1 {
		pxSize = 1
	}

	fractSize, _ := efixed.FromFloat64(pxSize)
	t.fontRenderer.SetSizePxFract(fractSize)
}

func (t *Throttle) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions, velocity, targetVelocity, maxVelocity, maxReverse float64) {
	screen := hudOpts.Screen
	t.fontRenderer.SetTarget(screen)
	t.fontRenderer.SetColor(hudOpts.Color)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	t.updateFontSize(bW, bH)

	maxX, zeroY := float64(bW), float64(bH)*maxVelocity/(maxVelocity+maxReverse)

	// current throttle velocity box
	var velocityRatio float64 = velocity / (maxVelocity + maxReverse)
	vW, vH := float64(bW)/6, -velocityRatio*float64(bH)
	vAlpha := uint8(4 * int(hudOpts.Color.A) / 5)
	ebitenutil.DrawRect(screen, float64(bX)+maxX-vW, float64(bY)+zeroY, vW, vH, color.RGBA{hudOpts.Color.R, hudOpts.Color.G, hudOpts.Color.B, vAlpha})

	// throttle indicator outline
	// FIXME: when ebitengine v2.5 releases can draw rect outline using StrokeRect
	//        - import "github.com/hajimehoshi/ebiten/v2/vector"
	//        - StrokeRect(dst *ebiten.Image, x, y, width, height float32, strokeWidth float32, hudOpts.Color color.Color)
	var oT float64 = 2 // TODO: calculate line thickness based on image height
	oX, oY, oW, oH := float64(bX)+maxX-vW, float64(bY), vW, float64(bH)
	ebitenutil.DrawRect(screen, oX, oY, oW, oT, hudOpts.Color)
	ebitenutil.DrawRect(screen, oX+oW-oT, oY, oT, oH, hudOpts.Color)
	ebitenutil.DrawRect(screen, oX, oY+oH-oT, oW, oT, hudOpts.Color)
	ebitenutil.DrawRect(screen, oX, oY, oT, oH, hudOpts.Color)
	ebitenutil.DrawRect(screen, oX, float64(bY)+zeroY, oW, oT, hudOpts.Color)

	// current throttle velocity text
	velocityStr := fmt.Sprintf("%0.1f kph", velocity)
	if velocity >= 0 {
		t.fontRenderer.SetAlign(etxt.Top, etxt.Right)
	} else {
		t.fontRenderer.SetAlign(etxt.Bottom, etxt.Right)
	}
	t.fontRenderer.Draw(velocityStr, int(oX)-3, bY+int(zeroY+vH)) // TODO: calculate better margin spacing

	// target velocity throttle indicator line
	var tgtVelocityRatio float64 = targetVelocity / (maxVelocity + maxReverse)
	tH := -tgtVelocityRatio * float64(bH)
	iW, iH := vW, 5.0 // TODO: calculate line thickness based on image height
	iX, iY := oX, zeroY+tH-iH
	if iY < 0 {
		iY = 0
	} else if iY > float64(bH)-iH {
		iY = float64(bH) - iH
	}
	ebitenutil.DrawRect(screen, iX, float64(bY)+iY, iW, iH, hudOpts.Color)
}
