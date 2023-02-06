package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
)

var (
	_colorJetsLevel   = _colorDefaultRed
	_colorJetsOutline = _colorDefaultRed
	_colorJetsText    = _colorDefaultGreen
)

type JumpJetIndicator struct {
	HUDSprite
	fontRenderer *etxt.Renderer
}

//NewJumpJetIndicator creates a jump jet indicator image to be rendered on demand
func NewJumpJetIndicator(font *Font) *JumpJetIndicator {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)

	j := &JumpJetIndicator{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
	}

	return j
}

func (j *JumpJetIndicator) updateFontSize(width, height int) {
	// set font size based on element size
	pxSize := float64(height) / 10
	if pxSize < 1 {
		pxSize = 1
	}

	fractSize, _ := efixed.FromFloat64(pxSize)
	j.fontRenderer.SetSizePxFract(fractSize)
}

func (j *JumpJetIndicator) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions, jumpJetDuration, maxJumpJetDuration float64) {
	screen := hudOpts.Screen
	j.fontRenderer.SetTarget(screen)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	j.updateFontSize(bW, bH)

	midX := float64(bX) + float64(bW)/2
	jW, jH := float64(bW)/4, 7*float64(bH)/8

	// current jet level box
	jetRatio := jumpJetDuration / maxJumpJetDuration
	if jetRatio > 1 {
		jetRatio = 1
	}
	rW, rH := jW, jetRatio*jH
	rX, rY := midX-jW/2, float64(bY)+jH-rH

	rColor := _colorJetsLevel
	if hudOpts.UseCustomColor {
		rColor = hudOpts.Color
	}

	ebitenutil.DrawRect(screen, rX, rY, rW, rH, rColor)

	// jet indicator outline
	oColor := _colorJetsOutline
	if hudOpts.UseCustomColor {
		oColor = hudOpts.Color
	}
	oAlpha := uint8(4 * (int(oColor.A) / 5))
	oColor = color.RGBA{oColor.R, oColor.G, oColor.B, oAlpha}

	// FIXME: when ebitengine v2.5 releases can draw rect outline using StrokeRect
	//        - import "github.com/hajimehoshi/ebiten/v2/vector"
	//        - StrokeRect(dst *ebiten.Image, x, y, width, height float32, strokeWidth float32, hudOpts.Color color.Color)
	var oT float64 = 2 // TODO: calculate line thickness based on image height
	oX, oY, oW, oH := midX-jW/2, float64(bY), jW, jH
	ebitenutil.DrawRect(screen, oX, oY, oW, oT, oColor)
	ebitenutil.DrawRect(screen, oX+oW-oT, oY, oT, oH, oColor)
	ebitenutil.DrawRect(screen, oX, oY+oH-oT, oW, oT, oColor)
	ebitenutil.DrawRect(screen, oX, oY, oT, oH, oColor)

	// jet indicator text
	tColor := _colorJetsText
	if hudOpts.UseCustomColor {
		tColor = hudOpts.Color
	}
	j.fontRenderer.SetColor(tColor)
	j.fontRenderer.SetAlign(etxt.Top, etxt.XCenter)
	j.fontRenderer.Draw("Jets", int(midX), int(oY+oH+2*oT)) // TODO: calculate better margin spacing
}
