package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tinne26/etxt"
)

type Throttle struct {
	HUDSprite
	image        *ebiten.Image
	fontRenderer *etxt.Renderer
}

//NewThrottle creates a speed indicator image to be rendered on demand
func NewThrottle(width, height int, font *Font) *Throttle {
	img := ebiten.NewImage(width, height)

	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetSizePx(16)
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.YCenter, etxt.Right)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	t := &Throttle{
		HUDSprite:    NewHUDSprite(img, 1.0),
		image:        img,
		fontRenderer: renderer,
	}

	return t
}

func (t *Throttle) Update(velocity, targetVelocity, maxVelocity, maxReverse float64) {
	t.image.Clear()

	t.fontRenderer.SetTarget(t.image)

	maxX, zeroY := float64(t.Width()), float64(t.Height())*maxVelocity/(maxVelocity+maxReverse)

	// current throttle velocity box
	var velocityRatio float64 = velocity / (maxVelocity + maxReverse)
	vW, vH := float64(t.Width())/6, -velocityRatio*float64(t.Height())
	ebitenutil.DrawRect(t.image, maxX-vW, zeroY, vW, vH, color.RGBA{255, 255, 255, 128})

	// throttle indicator outline
	// FIXME: when ebitengine v2.5 releases can draw rect outline using StrokeRect
	//        - import "github.com/hajimehoshi/ebiten/v2/vector"
	//        - StrokeRect(dst *ebiten.Image, x, y, width, height float32, strokeWidth float32, clr color.Color)
	var oT float64 = 2 // TODO: calculate line thickness based on image height
	oX, oY, oW, oH := maxX-vW, 0.0, vW, float64(t.Height())
	ebitenutil.DrawRect(t.image, oX, oY, oW, oT, color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(t.image, oX+oW-oT, oY, oT, oH, color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(t.image, oX, oY+oH-oT, oW, oT, color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(t.image, oX, oY, oT, oH, color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(t.image, oX, zeroY, oW, oT, color.RGBA{255, 255, 255, 255})

	// current throttle velocity text
	velocityStr := fmt.Sprintf("%0.1f kph", velocity)
	if velocity >= 0 {
		t.fontRenderer.SetAlign(etxt.Top, etxt.Right)
	} else {
		t.fontRenderer.SetAlign(etxt.Bottom, etxt.Right)
	}
	t.fontRenderer.Draw(velocityStr, int(oX)-3, int(zeroY+vH)) // TODO: calculate better margin spacing

	// target velocity throttle indicator line
	var tgtVelocityRatio float64 = targetVelocity / (maxVelocity + maxReverse)
	tH := -tgtVelocityRatio * float64(t.Height())
	iW, iH := vW, 5.0 // TODO: calculate line thickness based on image height
	iX, iY := oX, zeroY+tH-iH
	if iY < 0 {
		iY = 0
	} else if iY > float64(t.Height())-iH {
		iY = float64(t.Height()) - iH
	}
	ebitenutil.DrawRect(t.image, iX, iY, iW, iH, color.RGBA{255, 255, 255, 255})
}

func (t *Throttle) Texture() *ebiten.Image {
	return t.image
}
