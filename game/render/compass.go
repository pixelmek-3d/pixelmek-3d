package render

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
)

var (
	// define default colors
	_colorCompassPips   = _colorDefaultGreen
	_colorCompassTurret = color.NRGBA{R: 0, G: 127, B: 0, A: 255}
)

type Compass struct {
	HUDSprite
	fontRenderer    *etxt.Renderer
	targetIndicator *compassIndicator
	navIndicator    *compassIndicator
}

type compassIndicator struct {
	heading float64
	enabled bool
}

// NewCompass creates a compass image to be rendered on demand
func NewCompass(font *Font) *Compass {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.Top, etxt.XCenter)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	c := &Compass{
		HUDSprite:       NewHUDSprite(nil, 1.0),
		fontRenderer:    renderer,
		targetIndicator: &compassIndicator{heading: 2 * geom.Pi / 3, enabled: true},
		navIndicator:    &compassIndicator{heading: geom.Pi / 3, enabled: true},
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

func (c *Compass) SetTargetEnabled(b bool) {
	c.targetIndicator.enabled = b
}

func (c *Compass) SetTargetHeading(heading float64) {
	if heading < 0 {
		heading += geom.Pi2
	}
	c.targetIndicator.heading = heading
}

func (c *Compass) SetNavEnabled(b bool) {
	c.navIndicator.enabled = b
}

func (c *Compass) SetNavHeading(heading float64) {
	if heading < 0 {
		heading += geom.Pi2
	}
	c.navIndicator.heading = heading
}

func (c *Compass) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions, heading, turretAngle float64) {
	screen := hudOpts.Screen
	c.fontRenderer.SetTarget(screen)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	c.updateFontSize(bW, bH)

	// turret angle appears opposite because it is relative to body heading which counts up counter clockwise
	compassTurretAngle := -turretAngle
	headingDeg := geom.Degrees(heading)
	relTurretDeg := geom.Degrees(compassTurretAngle)

	midX, topY := float32(bX)+float32(bW)/2, float32(bY)

	// turret indicator box
	turretColor := _colorCompassTurret
	if hudOpts.UseCustomColor {
		turretColor = hudOpts.Color
	} else {
		turretColor.A = hudOpts.Color.A
	}

	var maxTurretDeg float64 = 90
	relTurretRatio := relTurretDeg / maxTurretDeg
	tW, tH := float32(relTurretRatio)*float32(bW)/2, float32(bH/4)
	tAlpha := uint8(4 * int(turretColor.A) / 5)
	vector.DrawFilledRect(screen, midX, topY, tW, tH, color.NRGBA{turretColor.R, turretColor.G, turretColor.B, tAlpha}, false)

	// compass pips
	pipColor := _colorCompassPips
	if hudOpts.UseCustomColor {
		pipColor = hudOpts.Color
	} else {
		pipColor.A = hudOpts.Color.A
	}
	c.fontRenderer.SetColor(color.RGBA(pipColor))

	for i := int(-maxTurretDeg); i <= int(maxTurretDeg); i++ {
		actualDeg := i + int(math.Round(headingDeg))
		if actualDeg < 0 {
			actualDeg += 360
		} else if actualDeg >= 360 {
			actualDeg -= 360
		}

		var pipWidth, pipHeight float32
		if actualDeg%10 == 0 {
			pipWidth = 2
			pipHeight = float32(bH) / 4
		}
		if actualDeg%30 == 0 {
			pipWidth = 3
			pipHeight = float32(bH) / 3
		}

		if pipWidth > 0 {
			// pip shows relative based on index (i) where negative is right of center, positive is left
			iRatio := float32(-i) / float32(maxTurretDeg)
			iX := float32(bX) + float32(bW)/2 + iRatio*float32(bW)/2
			vector.DrawFilledRect(screen, iX-pipWidth/2, topY, pipWidth, pipHeight, pipColor, false)

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
				c.fontRenderer.Draw(pipDegStr, int(iX), int(topY+float32(bH)/2)+2)
			}
		}
	}

	// heading indicator line
	headingColor := _colorAltimeter
	if hudOpts.UseCustomColor {
		headingColor = hudOpts.Color
	} else {
		headingColor.A = hudOpts.Color.A
	}

	hW, hH := float32(5.0), float32(bH)/2 // TODO: calculate line thickness based on image height
	vector.DrawFilledRect(screen, midX-hW/2, topY, hW, hH, headingColor, false)

	if c.targetIndicator.enabled {
		// TODO: draw target indicator slightly better
		iHeading := c.targetIndicator.heading
		iDeg := int(geom.Degrees(iHeading))

		iColor := _colorEnemy
		if hudOpts.UseCustomColor {
			iColor = hudOpts.Color
		} else {
			iColor.A = hudOpts.Color.A
		}

		iRendered := false
		for i := int(-maxTurretDeg); i <= int(maxTurretDeg); i++ {
			actualDeg := i + int(math.Round(headingDeg))
			if actualDeg < 0 {
				actualDeg += 360
			} else if actualDeg >= 360 {
				actualDeg -= 360
			}
			if iDeg == actualDeg {
				iRadius := float32(bH) / 8
				iRatio := float32(-i) / float32(maxTurretDeg)
				iX := float32(bX) + float32(bW)/2 + iRatio*float32(bW)/2

				vector.DrawFilledCircle(screen, iX-iRadius, topY-iRadius, iRadius, iColor, false)
				iRendered = true
				break
			}
		}

		if !iRendered {
			// draw indicator that target is outside of current compass range
			actualMinDeg := headingDeg - maxTurretDeg
			iMinFound := model.IsBetweenDegrees(actualMinDeg, actualMinDeg-90, float64(iDeg))

			var iRatio float32
			if iMinFound {
				iRatio = -1
			} else {
				iRatio = 1
			}

			iRadius := float32(bH) / 12
			iX := float32(bX) + float32(bW)/2 + iRatio*float32(bW)/2
			vector.DrawFilledCircle(screen, iX-iRadius, topY-iRadius, iRadius, iColor, false)
		}
	}

	if c.navIndicator.enabled {
		// TODO: draw nav indicator slightly better
		iHeading := c.navIndicator.heading
		iDeg := int(geom.Degrees(iHeading))

		iColor := _colorNavPoint
		if hudOpts.UseCustomColor {
			iColor = hudOpts.Color
		} else {
			iColor.A = hudOpts.Color.A
		}

		iRendered := false
		for i := int(-maxTurretDeg); i <= int(maxTurretDeg); i++ {
			actualDeg := i + int(math.Round(headingDeg))
			if actualDeg < 0 {
				actualDeg += 360
			} else if actualDeg >= 360 {
				actualDeg -= 360
			}
			if iDeg == actualDeg {
				iRadius := float32(bH) / 8
				iRatio := float32(-i) / float32(maxTurretDeg)
				iX := float32(bX) + float32(bW)/2 + iRatio*float32(bW)/2

				vector.DrawFilledCircle(screen, iX-iRadius, topY-iRadius, iRadius, iColor, false)
				iRendered = true
				break
			}
		}

		if !iRendered {
			// draw indicator that target is outside of current compass range
			actualMinDeg := headingDeg - maxTurretDeg
			iMinFound := model.IsBetweenDegrees(actualMinDeg, actualMinDeg-90, float64(iDeg))

			var iRatio float32
			if iMinFound {
				iRatio = -1
			} else {
				iRatio = 1
			}

			iRadius := float32(bH) / 12
			iX := float32(bX) + float32(bW)/2 + iRatio*float32(bW)/2
			vector.DrawFilledCircle(screen, iX-iRadius, topY-iRadius, iRadius, iColor, false)
		}
	}
}
