package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tinne26/etxt"
)

var (
	_colorJetsLevel   = _colorDefaultRed
	_colorJetsOutline = _colorDefaultRed
	_colorJetsText    = _colorDefaultGreen
)

type JumpJetIndicator struct {
	HUDSprite
	fontRenderer       *etxt.Renderer
	jumpJetDuration    float64
	maxJumpJetDuration float64
}

// NewJumpJetIndicator creates a jump jet indicator image to be rendered on demand
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

	j.fontRenderer.SetSizePx(int(pxSize))
}

func (j *JumpJetIndicator) SetValues(jumpJetDuration, maxJumpJetDuration float64) {
	j.jumpJetDuration = jumpJetDuration
	j.maxJumpJetDuration = maxJumpJetDuration
}

func (j *JumpJetIndicator) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen
	j.fontRenderer.SetTarget(screen)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	j.updateFontSize(bW, bH)

	midX := float32(bX) + float32(bW)/2
	jW, jH := float32(bW)/4, 7*float32(bH)/8

	// current jet level box
	jetRatio := j.jumpJetDuration / j.maxJumpJetDuration
	if jetRatio > 1 {
		jetRatio = 1
	}
	rW, rH := jW, float32(jetRatio)*jH
	rX, rY := midX-jW/2, float32(bY)+jH-rH

	rColor := hudOpts.HudColor(_colorJetsLevel)
	vector.DrawFilledRect(screen, rX, rY, rW, rH, rColor, false)

	// jet indicator outline
	oColor := hudOpts.HudColor(_colorJetsOutline)
	oColor.A = uint8(4 * (int(oColor.A) / 5))

	var oT float32 = 2 // TODO: calculate line thickness based on image height
	oX, oY, oW, oH := float32(midX-jW/2), float32(bY), float32(jW), float32(jH)
	vector.StrokeRect(screen, oX, oY, oW, oH, oT, oColor, false)

	// jet indicator text
	tColor := hudOpts.HudColor(_colorJetsText)
	j.fontRenderer.SetColor(color.RGBA(tColor))
	j.fontRenderer.SetAlign(etxt.Top, etxt.XCenter)
	j.fontRenderer.Draw("Jets", int(midX), int(oY+oH+2*oT)) // TODO: calculate better margin spacing
}
