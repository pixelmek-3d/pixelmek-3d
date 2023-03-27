package render

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
)

var (
	_colorNavPoint = _colorDefaultYellow
)

type NavStatus struct {
	HUDSprite
	fontRenderer *etxt.Renderer
	navPoint     *model.NavPoint
	navDistance  float64
}

// NewNavStatus creates a nav status element image to be rendered on demand
func NewNavStatus(font *Font) *NavStatus {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)

	n := &NavStatus{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
		navDistance:  -1,
	}

	return n
}

func (n *NavStatus) SetNavPoint(navPoint *model.NavPoint) {
	n.navPoint = navPoint
}

func (n *NavStatus) SetNavDistance(distance float64) {
	n.navDistance = distance
}

func (n *NavStatus) updateFontSize(width, height int) {
	// set font size based on element size
	pxSize := float64(height) / 8
	if pxSize < 1 {
		pxSize = 1
	}

	fractSize, _ := efixed.FromFloat64(pxSize)
	n.fontRenderer.SetSizePxFract(fractSize)
}

func (n *NavStatus) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen
	n.fontRenderer.SetTarget(screen)
	n.fontRenderer.SetAlign(etxt.YCenter, etxt.XCenter)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	n.updateFontSize(bW, bH)

	sW, sH := float32(bW), float32(bH)
	sX, sY := float32(bX), float32(bY)

	// background box
	bColor := _colorStatusBackground
	if hudOpts.UseCustomColor {
		bColor = hudOpts.Color
	}

	sAlpha := uint8(int(bColor.A) / 3)
	vector.DrawFilledRect(screen, sX, sY, sW, sH, color.RGBA{bColor.R, bColor.G, bColor.B, sAlpha}, false)

	// draw nav image
	nTexture := n.navPoint.Image()

	iH := bounds.Dy()
	nH := nTexture.Bounds().Dy()
	nScale := (0.7 * float64(iH)) / float64(nH)

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(nScale, nScale)
	op.GeoM.Translate(float64(sX+sW/2)-nScale*float64(nH)/2, float64(sY+sH/2)-nScale*float64(nH)/2)
	screen.DrawImage(nTexture, op)

	// setup text color
	tColor := _colorStatusText
	if hudOpts.UseCustomColor {
		tColor = hudOpts.Color
	}
	n.fontRenderer.SetColor(tColor)

	// nav point distance
	if n.navDistance >= 0 {
		n.fontRenderer.SetAlign(etxt.Bottom, etxt.XCenter)
		distanceStr := fmt.Sprintf("%0.0fm", n.navDistance)
		n.fontRenderer.Draw(distanceStr, bX+bW/2, bY+bH)
	}

	// nav point name
	nColor := _colorNavPoint
	if hudOpts.UseCustomColor {
		nColor = hudOpts.Color
	}
	n.fontRenderer.SetColor(nColor)
	n.fontRenderer.SetAlign(etxt.Top, etxt.XCenter)

	navName := "NAV " + strings.ToUpper(n.navPoint.Name)
	n.fontRenderer.Draw(navName, bX+bW/2, bY)
}
