package render

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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

//NewNavStatus creates a nav status element image to be rendered on demand
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

func GenerateNavImage(navPoint *model.NavPoint, imageSize int, font *Font, color *color.RGBA) *ebiten.Image {
	if navPoint == nil {
		return nil
	}

	navImage := ebiten.NewImage(imageSize, imageSize)
	renderer := etxt.NewStdRenderer()

	if color == nil {
		color = &_colorNavPoint
	}

	renderer.SetTarget(navImage)
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)
	renderer.SetColor(color)

	// set font size based on image size
	fontPxSize := float64(imageSize) / 3
	if fontPxSize < 1 {
		fontPxSize = 1
	}

	fractSize, _ := efixed.FromFloat64(fontPxSize)
	renderer.SetSizePxFract(fractSize)
	renderer.SetAlign(etxt.YCenter, etxt.XCenter)

	navChar := navPoint.Name[0:1]
	renderer.Draw(navChar, imageSize/2, imageSize/2)

	// draw nav diamond shape
	minX, minY := float64(imageSize)/8, float64(imageSize)/8
	maxX, maxY := 7*float64(imageSize)/8, 7*float64(imageSize)/8
	midX, midY := float64(imageSize)/2, float64(imageSize/2)
	ebitenutil.DrawLine(navImage, minX, midY, midX, minY, color)
	ebitenutil.DrawLine(navImage, midX, minY, maxX, midY, color)
	ebitenutil.DrawLine(navImage, minX, midY, midX, maxY, color)
	ebitenutil.DrawLine(navImage, midX, maxY, maxX, midY, color)

	return navImage
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

	sW, sH := float64(bW), float64(bH)
	sX, sY := float64(bX), float64(bY)

	// background box
	bColor := _colorStatusBackground
	if hudOpts.UseCustomColor {
		bColor = hudOpts.Color
	}

	sAlpha := uint8(int(bColor.A) / 3)
	ebitenutil.DrawRect(screen, sX, sY, sW, sH, color.RGBA{bColor.R, bColor.G, bColor.B, sAlpha})

	// draw nav image
	nTexture := n.navPoint.Image()

	iH := bounds.Dy()
	nH := nTexture.Bounds().Dy()
	nScale := (0.7 * float64(iH)) / float64(nH)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(nScale, nScale)
	op.GeoM.Translate(sX+sW/2-nScale*float64(nH)/2, sY+sH/2-nScale*float64(nH)/2)
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
