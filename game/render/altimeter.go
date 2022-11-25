package render

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/tinne26/etxt"
)

type Altimeter struct {
	HUDSprite
	image        *ebiten.Image
	fontRenderer *etxt.Renderer
}

//NewAltimeter creates a compass image to be rendered on demand
func NewAltimeter(width, height int, font *Font) *Altimeter {
	img := ebiten.NewImage(width, height)

	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetSizePx(12)
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.YCenter, etxt.Right)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	a := &Altimeter{
		HUDSprite:    NewHUDSprite(img, 1.0),
		image:        img,
		fontRenderer: renderer,
	}

	return a
}

func (a *Altimeter) Update(altitude, pitch float64) {
	a.image.Clear()

	a.fontRenderer.SetTarget(a.image)

	// use opposite pitch value so indicator will draw upward from center when postive angle
	relPitchAngle := -pitch
	relPitchDeg := geom.Degrees(relPitchAngle)

	midX, midY := float64(a.Width())/2, float64(a.Height())/2

	// pitch indicator box
	var maxPitchDeg float64 = 45
	pitchRatio := relPitchDeg / maxPitchDeg
	tW, tH := float64(a.Width())/4, pitchRatio*float64(a.Height()/2)
	ebitenutil.DrawRect(a.image, midX, midY, tW, tH, color.RGBA{255, 255, 255, 192})

	// altimeter pips
	var maxAltitude float64 = model.METERS_PER_UNIT
	for i := int(-maxAltitude); i <= int(maxAltitude); i++ {
		actualAlt := i + int(math.Round(altitude))

		var pipWidth, pipHeight float64
		if actualAlt%5 == 0 {
			pipWidth = float64(a.Width() / 4)
			pipHeight = 2
		}
		if actualAlt%10 == 0 {
			pipWidth = float64(a.Width() / 2)
			pipHeight = 3
		}

		if pipWidth > 0 {
			// pip shows relative based on index (i) where negative is above center, positive is below
			iRatio := float64(-i) / maxAltitude
			iY := float64(a.Height())/2 + iRatio*float64(a.Height())/2
			ebitenutil.DrawRect(a.image, midX, iY-pipHeight/2, pipWidth, pipHeight, color.RGBA{255, 255, 255, 255})

			var pipAltStr string = fmt.Sprintf("%d", actualAlt)

			if pipAltStr != "" {
				a.fontRenderer.Draw(pipAltStr, int(midX), int(iY))
			}
		}
	}

	// heading indicator line
	hW, hH := float64(a.Width()/2), 5.0 // TODO: calculate line thickness based on image height
	ebitenutil.DrawRect(a.image, midX, midY-hH/2, hW, hH, color.RGBA{255, 255, 255, 255})
}

func (a *Altimeter) Texture() *ebiten.Image {
	return a.image
}
