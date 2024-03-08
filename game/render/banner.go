package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tinne26/etxt"
)

var (
	_colorBannerText       = _colorDefaultGreen
	_colorBannerBackground = color.NRGBA{R: 50, G: 50, B: 50, A: 255}
)

type MissionBanner struct {
	HUDSprite
	fontRenderer *etxt.Renderer
	bannerText   string
}

// NewMissionBanner creates an in-mission banner to be rendered
func NewMissionBanner(font *Font) *MissionBanner {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)

	b := &MissionBanner{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
	}

	return b
}

func (f *MissionBanner) updateFontSize(_, height int) {
	// set font size based on element size
	pxSize := 5 * float64(height) / 8
	if pxSize < 1 {
		pxSize = 1
	}

	f.fontRenderer.SetSizePx(int(pxSize))
}

func (f *MissionBanner) SetBannerText(bannerText string) {
	f.bannerText = bannerText
}

func (f *MissionBanner) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen
	f.fontRenderer.SetTarget(screen)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	f.updateFontSize(bW, bH)

	// background box
	bColor := hudOpts.HudColor(_colorBannerBackground)
	sAlpha := uint8(2 * int(bColor.A) / 3)
	vector.DrawFilledRect(screen, float32(bX), float32(bY), float32(bW), float32(bH), color.NRGBA{bColor.R, bColor.G, bColor.B, sAlpha}, false)

	// mission banner text
	tColor := hudOpts.HudColor(_colorBannerText)
	f.fontRenderer.SetColor(color.RGBA(tColor))
	f.fontRenderer.SetAlign(etxt.Top, etxt.Left)
	f.fontRenderer.Draw(f.bannerText, bX, bY)
}
