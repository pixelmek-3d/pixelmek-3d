package render

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
)

var (
	_colorThrottleForward = _colorDefaultGreen
	_colorThrottleReverse = _colorDefaultBlue
	_colorThrottleOutline = _colorDefaultRed
	_colorThrottleText    = _colorDefaultGreen
)

type Throttle struct {
	HUDSprite
	fontRenderer *etxt.Renderer
}

// NewThrottle creates a speed indicator image to be rendered on demand
func NewThrottle(font *Font) *Throttle {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.YCenter, etxt.Right)

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

func (t *Throttle) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions, velocity, targetVelocity, velocityZ, maxVelocity, maxReverse float64) {
	screen := hudOpts.Screen
	t.fontRenderer.SetTarget(screen)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	t.updateFontSize(bW, bH)

	maxX, zeroY := float32(bW), float32(bH)*float32(maxVelocity/(maxVelocity+maxReverse))

	// current throttle velocity box
	vColor := hudOpts.HudColor(_colorThrottleForward)
	if velocity < 0 {
		vColor = hudOpts.HudColor(_colorThrottleReverse)
	}
	vColor.A = hudOpts.Color.A

	var velocityRatio float32 = float32(velocity / (maxVelocity + maxReverse))
	vW, vH := float32(bW)/6, -velocityRatio*float32(bH)
	//vAlpha := uint8(4 * int(hudOpts.Color.A) / 5)
	vector.DrawFilledRect(screen, float32(bX)+maxX-vW, float32(bY)+zeroY, vW, vH, vColor, false)

	// throttle indicator outline
	oColor := hudOpts.HudColor(_colorThrottleOutline)

	var oT float32 = 2 // TODO: calculate line thickness based on image height
	oX, oY, oW, oH := float32(bX)+float32(maxX-vW), float32(bY), float32(vW), float32(bH)
	vector.StrokeRect(screen, oX, oY, oW, oH, oT, oColor, false)

	// current throttle velocity text
	tColor := hudOpts.HudColor(_colorThrottleText)
	t.fontRenderer.SetColor(color.RGBA(tColor))

	velocityStr := fmt.Sprintf("%0.1f kph", velocity)
	if velocityZ != 0 {
		velocityStr += fmt.Sprintf("\n%0.1fvert", velocityZ)
	}
	if velocity >= 0 {
		t.fontRenderer.SetAlign(etxt.Top, etxt.Right)
	} else {
		t.fontRenderer.SetAlign(etxt.Bottom, etxt.Right)
	}
	t.fontRenderer.Draw(velocityStr, int(oX)-3, bY+int(zeroY+vH)) // TODO: calculate better margin spacing

	// target velocity throttle indicator line
	vColor = hudOpts.HudColor(_colorThrottleForward)
	if targetVelocity < 0 {
		vColor = hudOpts.HudColor(_colorThrottleReverse)
	}

	var tgtVelocityRatio float32 = float32(targetVelocity / (maxVelocity + maxReverse))
	tH := -tgtVelocityRatio * float32(bH)
	iW, iH := vW, float32(5.0) // TODO: calculate line thickness based on image height
	iX, iY := float32(oX), zeroY+tH-iH
	if iY < 0 {
		iY = 0
	} else if iY > float32(bH)-iH {
		iY = float32(bH) - iH
	}
	vector.DrawFilledRect(screen, iX, float32(bY)+iY, iW, iH, vColor, false)
}
