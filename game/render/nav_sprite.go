package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
)

type NavSprite struct {
	*Sprite
	NavPoint *model.NavPoint
}

func NewNavSprite(
	navPoint *model.NavPoint, scale float64,
) *NavSprite {

	navPos := navPoint.Pos()
	navEntity := model.BasicVisualEntity(navPos.X, navPos.Y, 0.5, raycaster.AnchorCenter)
	n := &NavSprite{
		Sprite:   NewSprite(navEntity, scale, navPoint.Image()),
		NavPoint: navPoint,
	}

	// nav points cannot be focused upon by player reticle
	n.Focusable = false

	// nav points self illuminate so they do not get dimmed in night conditions
	n.illumination = 5000

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
