package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/fonts"
	"github.com/tinne26/etxt"
)

var (
	_colorBannerText       = _colorDefaultGreen
	_colorBannerBackground = color.NRGBA{R: 50, G: 50, B: 50, A: 200}
)

type MissionBanner struct {
	HUDSprite
	fontRenderer *etxt.Renderer
	bannerText   string
}

// NewMissionBanner creates an in-mission banner to be rendered
func NewMissionBanner(font *fonts.Font) *MissionBanner {
	// create and configure font renderer
	renderer := etxt.NewRenderer()
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

	f.fontRenderer.SetSize(pxSize)
}

func (f *MissionBanner) SetBannerText(bannerText string) {
	f.bannerText = bannerText
}

func (f *MissionBanner) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen

	sW := screen.Bounds().Dx()
	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	f.updateFontSize(bW, bH)

	// background box
	vector.DrawFilledRect(screen, 0, float32(bY), float32(sW), float32(bH), _colorBannerBackground, false)

	// mission banner text
	f.fontRenderer.SetColor(_colorBannerText)
	f.fontRenderer.SetAlign(etxt.Top | etxt.Left)
	f.fontRenderer.Draw(screen, f.bannerText, bX, bY+bH/8)
}
