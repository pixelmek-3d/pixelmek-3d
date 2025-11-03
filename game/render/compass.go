package render

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/fonts"
	"github.com/tinne26/etxt"
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
	heading         float64
	turretAngle     float64
}

type compassIndicator struct {
	heading  float64
	enabled  bool
	friendly bool
}

// NewCompass creates a compass image to be rendered on demand
func NewCompass(font *fonts.Font) *Compass {
	// create and configure font renderer
	renderer := etxt.NewRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.Top | etxt.HorzCenter)
	renderer.SetColor(color.NRGBA{255, 255, 255, 255})

	c := &Compass{
		HUDSprite:       NewHUDSprite(nil, 1.0),
		fontRenderer:    renderer,
		targetIndicator: &compassIndicator{heading: 2 * geom.Pi / 3, enabled: true},
		navIndicator:    &compassIndicator{heading: geom.Pi / 3, enabled: true},
	}

	return c
}

func (c *Compass) updateFontSize(_, height int) {
	// set font size based on element size
	pxSize := float64(height) / 2
	if pxSize < 1 {
		pxSize = 1
	}

	c.fontRenderer.SetSize(pxSize)
}

func (c *Compass) SetTargetEnabled(b bool) {
	c.targetIndicator.enabled = b
}

func (c *Compass) SetTargetFriendly(friendly bool) {
	c.targetIndicator.friendly = friendly
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

func (c *Compass) SetValues(heading, turretAngle float64) {
	c.heading = heading
	c.turretAngle = turretAngle
}

func (c *Compass) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	c.updateFontSize(bW, bH)

	// turret angle appears opposite because it is relative to body heading which counts up counter clockwise
	relTurretAngle := -model.AngleDistance(c.heading, c.turretAngle)
	headingDeg := model.AngleToCardinal(c.heading)
	relTurretDeg := geom.Degrees(relTurretAngle)

	midX, topY := float32(bX)+float32(bW)/2, float32(bY)

	// turret indicator box
	turretColor := hudOpts.HudColor(_colorCompassTurret)

	// TODO: support 360 degree turret rotation option
	var maxTurretDeg float64 = 90
	relTurretRatio := relTurretDeg / maxTurretDeg
	tW, tH := float32(relTurretRatio)*float32(bW)/2, float32(bH/4)
	tAlpha := uint8(4 * int(turretColor.A) / 5)
	vector.DrawFilledRect(screen, midX, topY, tW, tH, color.NRGBA{turretColor.R, turretColor.G, turretColor.B, tAlpha}, false)

	// compass pips
	pipColor := hudOpts.HudColor(_colorCompassPips)
	c.fontRenderer.SetColor(pipColor)

	for i := int(-maxTurretDeg); i <= int(maxTurretDeg); i++ {
		compassDeg := i + int(math.Round(headingDeg))
		if compassDeg < 0 {
			compassDeg += 360
		} else if compassDeg >= 360 {
			compassDeg -= 360
		}

		var pipWidth, pipHeight float32
		if compassDeg%10 == 0 {
			pipWidth = 2
			pipHeight = float32(bH) / 4
		}
		if compassDeg%30 == 0 {
			pipWidth = 3
			pipHeight = float32(bH) / 3
		}

		if pipWidth > 0 {
			// clockwise pip shows relative based on index (i) where negative is left of center, positive is right
			iRatio := float32(i) / float32(maxTurretDeg)
			iX := float32(bX) + float32(bW)/2 + iRatio*float32(bW)/2
			vector.DrawFilledRect(screen, iX-pipWidth/2, topY, pipWidth, pipHeight, pipColor, false)

			var pipDegStr string
			switch {
			case compassDeg == 0:
				pipDegStr = "N"
			case compassDeg == 90:
				pipDegStr = "E"
			case compassDeg == 180:
				pipDegStr = "S"
			case compassDeg == 270:
				pipDegStr = "W"
			case compassDeg%30 == 0:
				pipDegStr = fmt.Sprintf("%d", compassDeg)
			}

			if pipDegStr != "" {
				c.fontRenderer.Draw(screen, pipDegStr, int(iX), int(topY+float32(bH)/2)+2)
			}
		}
	}

	// heading indicator line
	headingColor := hudOpts.HudColor(_colorAltimeter)

	hW, hH := float32(5.0), float32(bH)/2 // TODO: calculate line thickness based on image height
	vector.DrawFilledRect(screen, midX-hW/2, topY, hW, hH, headingColor, false)

	if c.navIndicator.enabled {
		// TODO: draw nav indicator slightly better
		iHeading := c.navIndicator.heading
		iDeg := int(model.AngleToCardinal(iHeading))

		iColor := hudOpts.HudColor(_colorNavPoint)

		iRendered := false
		for i := int(-maxTurretDeg); i <= int(maxTurretDeg); i++ {
			compassDeg := i + int(math.Round(headingDeg))
			if compassDeg < 0 {
				compassDeg += 360
			} else if compassDeg >= 360 {
				compassDeg -= 360
			}
			if iDeg == compassDeg {
				iRadius := float32(bH) / 8
				iRatio := float32(i) / float32(maxTurretDeg)
				iX := float32(bX) + float32(bW)/2 + iRatio*float32(bW)/2

				vector.DrawFilledCircle(screen, iX-iRadius, topY-iRadius, iRadius, iColor, false)
				iRendered = true
				break
			}
		}

		if !iRendered {
			// draw indicator that target is outside of current compass range
			compassMinDeg := headingDeg - maxTurretDeg
			iMinFound := model.IsBetweenDegrees(compassMinDeg, compassMinDeg-90, float64(iDeg))

			var iRatio float32
			if iMinFound {
				iRatio = 1
			} else {
				iRatio = -1
			}

			iRadius := float32(bH) / 12
			iX := float32(bX) + float32(bW)/2 + iRatio*float32(bW)/2
			vector.DrawFilledCircle(screen, iX-iRadius, topY-iRadius, iRadius, iColor, false)
		}
	}

	if c.targetIndicator.enabled {
		// TODO: draw target indicator slightly better
		iHeading := c.targetIndicator.heading
		iDeg := int(model.AngleToCardinal(iHeading))

		var iColor color.NRGBA
		if c.targetIndicator.friendly {
			iColor = hudOpts.HudColor(_colorFriendly)
		} else {
			iColor = hudOpts.HudColor(_colorEnemy)
		}

		iRendered := false
		for i := int(-maxTurretDeg); i <= int(maxTurretDeg); i++ {
			compassDeg := i + int(math.Round(headingDeg))
			if compassDeg < 0 {
				compassDeg += 360
			} else if compassDeg >= 360 {
				compassDeg -= 360
			}
			if iDeg == compassDeg {
				iRadius := float32(bH) / 4
				iRatio := float32(i) / float32(maxTurretDeg)
				iX := float32(bX) + float32(bW)/2 + iRatio*float32(bW)/2

				//vector.DrawFilledCircle(screen, iX-iRadius, topY-iRadius, iRadius, iColor, false)
				if c.targetIndicator.friendly {
					// differentiate friendly by not filling in the radar blip box
					vector.StrokeRect(screen, iX-iRadius, topY-iRadius-2, iRadius, iRadius, 2, iColor, false)
				} else {
					vector.DrawFilledRect(screen, iX-iRadius, topY-iRadius-2, iRadius, iRadius, iColor, false) // TODO: calculate thickness based on image size
				}
				iRendered = true
				break
			}
		}

		if !iRendered {
			// draw indicator that target is outside of current compass range
			compassMinDeg := headingDeg - maxTurretDeg
			iMinFound := model.IsBetweenDegrees(compassMinDeg, compassMinDeg-90, float64(iDeg))

			var iRatio float32
			if iMinFound {
				iRatio = 1
			} else {
				iRatio = -1
			}

			iRadius := float32(bH) / 8
			iX := float32(bX) + float32(bW)/2 + iRatio*float32(bW)/2
			//vector.DrawFilledCircle(screen, iX-iRadius, topY-iRadius, iRadius, iColor, false)
			if c.targetIndicator.friendly {
				// differentiate friendly by not filling in the radar blip box
				vector.StrokeRect(screen, iX-iRadius, topY-iRadius-2, iRadius, iRadius, 2, iColor, false)
			} else {
				vector.DrawFilledRect(screen, iX-iRadius, topY-iRadius-2, iRadius, iRadius, iColor, false) // TODO: calculate thickness based on image size
			}
		}
	}
}
