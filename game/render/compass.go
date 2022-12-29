package render

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/tinne26/etxt"
)

type Compass struct {
	HUDSprite
	fontRenderer *etxt.Renderer
}

//NewCompass creates a compass image to be rendered on demand
func NewCompass(font *Font) *Compass {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetSizePx(16)
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.Top, etxt.XCenter)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	c := &Compass{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
	}

	return c
}

func (c *Compass) Draw(screen *ebiten.Image, bounds image.Rectangle, clr *color.RGBA, heading, turretAngle float64) {
	c.fontRenderer.SetTarget(screen)
	c.fontRenderer.SetColor(clr)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()

	// turret angle appears opposite because it is relative to body heading which counts up counter clockwise
	compassTurretAngle := -turretAngle
	headingDeg := geom.Degrees(heading)
	relTurretDeg := geom.Degrees(compassTurretAngle)

	midX, topY := float64(bX)+float64(bW)/2, float64(bY)

	// turret indicator box
	var maxTurretDeg float64 = 90
	relTurretRatio := relTurretDeg / maxTurretDeg
	tW, tH := relTurretRatio*float64(bW)/2, float64(bH/4)
	tAlpha := uint8(4 * int(clr.A) / 5)
	ebitenutil.DrawRect(screen, midX, topY, tW, tH, color.RGBA{clr.R, clr.G, clr.B, tAlpha})
	// compass pips
	for i := int(-maxTurretDeg); i <= int(maxTurretDeg); i++ {
		actualDeg := i + int(math.Round(headingDeg))
		if actualDeg < 0 {
			actualDeg += 360
		} else if actualDeg >= 360 {
			actualDeg -= 360
		}

		var pipWidth, pipHeight float64
		if actualDeg%10 == 0 {
			pipWidth = 2
			pipHeight = float64(bH / 4)
		}
		if actualDeg%30 == 0 {
			pipWidth = 3
			pipHeight = float64(bH / 2)
		}

		if pipWidth > 0 {
			// pip shows relative based on index (i) where negative is right of center, positive is left
			iRatio := float64(-i) / maxTurretDeg
			iX := float64(bX) + float64(bW)/2 + iRatio*float64(bW)/2
			ebitenutil.DrawRect(screen, iX-pipWidth/2, topY, pipWidth, pipHeight, clr)

			// TODO: switch statement
			var pipDegStr string
			if actualDeg == 0 {
				pipDegStr = "E"
			} else if actualDeg == 90 {
				pipDegStr = "N"
			} else if actualDeg == 180 {
				pipDegStr = "W"
			} else if actualDeg == 270 {
				pipDegStr = "S"
			} else if actualDeg%30 == 0 {
				pipDegStr = fmt.Sprintf("%d", actualDeg)
			}

			if pipDegStr != "" {
				c.fontRenderer.Draw(pipDegStr, int(iX), int(float64(bH/2))+2)
			}
		}
	}

	// heading indicator line
	hW, hH := 5.0, float64(bH/2) // TODO: calculate line thickness based on image height
	ebitenutil.DrawRect(screen, midX-hW/2, topY, hW, hH, clr)
}
