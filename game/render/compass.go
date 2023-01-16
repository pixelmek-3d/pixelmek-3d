package render

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
)

var (
	// define default colors
	_colorCompass       color.RGBA = color.RGBA{R: 0, G: 255, B: 67, A: 255}
	_colorCompassPips   color.RGBA = color.RGBA{R: 0, G: 214, B: 0, A: 255}
	_colorCompassTurret color.RGBA = color.RGBA{R: 0, G: 127, B: 0, A: 255}
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
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.Top, etxt.XCenter)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	c := &Compass{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
	}

	return c
}

func (c *Compass) updateFontSize(width, height int) {
	// set font size based on element size
	pxSize := float64(height) / 2
	if pxSize < 1 {
		pxSize = 1
	}

	fractSize, _ := efixed.FromFloat64(pxSize)
	c.fontRenderer.SetSizePxFract(fractSize)
}

func (c *Compass) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions, heading, turretAngle float64) {
	screen := hudOpts.Screen
	c.fontRenderer.SetTarget(screen)
	c.fontRenderer.SetColor(hudOpts.Color)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	c.updateFontSize(bW, bH)

	// turret angle appears opposite because it is relative to body heading which counts up counter clockwise
	compassTurretAngle := -turretAngle
	headingDeg := geom.Degrees(heading)
	relTurretDeg := geom.Degrees(compassTurretAngle)

	midX, topY := float64(bX)+float64(bW)/2, float64(bY)

	// turret indicator box
	turretColor := _colorCompassTurret
	if hudOpts.UseCustomColor {
		turretColor = hudOpts.Color
	}

	var maxTurretDeg float64 = 90
	relTurretRatio := relTurretDeg / maxTurretDeg
	tW, tH := relTurretRatio*float64(bW)/2, float64(bH/4)
	tAlpha := uint8(4 * int(turretColor.A) / 5)
	ebitenutil.DrawRect(screen, midX, topY, tW, tH, color.RGBA{turretColor.R, turretColor.G, turretColor.B, tAlpha})

	// compass pips
	pipColor := _colorCompassPips
	if hudOpts.UseCustomColor {
		pipColor = hudOpts.Color
	}
	c.fontRenderer.SetColor(pipColor)

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
			pipHeight = float64(bH) / 4
		}
		if actualDeg%30 == 0 {
			pipWidth = 3
			pipHeight = float64(bH) / 3
		}

		if pipWidth > 0 {
			// pip shows relative based on index (i) where negative is right of center, positive is left
			iRatio := float64(-i) / maxTurretDeg
			iX := float64(bX) + float64(bW)/2 + iRatio*float64(bW)/2
			ebitenutil.DrawRect(screen, iX-pipWidth/2, topY, pipWidth, pipHeight, pipColor)

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
				c.fontRenderer.Draw(pipDegStr, int(iX), int(topY+float64(bH)/2)+2)
			}
		}
	}

	// heading indicator line
	headingColor := _colorAltimeter
	if hudOpts.UseCustomColor {
		headingColor = hudOpts.Color
	}

	hW, hH := 5.0, float64(bH)/2 // TODO: calculate line thickness based on image height
	ebitenutil.DrawRect(screen, midX-hW/2, topY, hW, hH, headingColor)
}
