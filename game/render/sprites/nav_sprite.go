package sprites

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/colors"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/fonts"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/shapes"
	"github.com/tinne26/etxt"
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
	n.focusable = false

	// nav points self illuminate so they do not get dimmed in night conditions
	n.illumination = 5000

	return n
}

func GenerateNavImage(navPoint *model.NavPoint, imageSize int, font *fonts.Font, clr *color.NRGBA) *ebiten.Image {
	if navPoint == nil {
		return nil
	}

	navImage := ebiten.NewImage(imageSize, imageSize)
	renderer := etxt.NewRenderer()

	if clr == nil {
		clr = &colors.NavPoint
	}

	nColor := color.NRGBA{R: clr.R, G: clr.G, B: clr.B, A: 255}

	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)
	renderer.SetColor(nColor)

	// set font size based on image size
	fontPxSize := float64(imageSize) / 3
	if fontPxSize < 1 {
		fontPxSize = 1
	}

	renderer.SetSize(fontPxSize)
	renderer.SetAlign(etxt.VertCenter | etxt.HorzCenter)

	navChar := navPoint.Name[0:1]
	renderer.Draw(navImage, navChar, imageSize/2, imageSize/2)

	// draw nav diamond shape
	oT := float32(2)
	r := float32(3*imageSize) / 8
	midX, midY := float32(imageSize)/2, float32(imageSize/2)
	shapes.StrokeDiamond(navImage, midX, midY, r, r, oT, nColor, false)

	return navImage
}
