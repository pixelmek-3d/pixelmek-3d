package render

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/raycaster-go/geom"
)

type Compass struct {
	HUDSprite
	image *ebiten.Image
}

//NewCompass creates a compass image to be rendered on demand
func NewCompass(width, height int) *Compass {
	img := ebiten.NewImage(width, height)
	c := &Compass{
		HUDSprite: NewHUDSprite(img, 1.0),
		image:     img,
	}

	return c
}

func (c *Compass) Update(heading, turretAngle float64) {
	c.image.Clear()

	// turret angle appears opposite because it is relative to body heading which counts up counter clockwise
	compassTurretAngle := -turretAngle
	headingDeg := geom.Degrees(heading)
	relTurretDeg := geom.Degrees(compassTurretAngle)

	midX, topY := float64(c.Width())/2, float64(0)

	// turret indicator box
	var maxTurretDeg float64 = 90
	relTurretRatio := relTurretDeg / maxTurretDeg
	tW, tH := relTurretRatio*float64(c.Width())/2, float64(c.Height()/4)
	ebitenutil.DrawRect(c.image, midX, topY, tW, tH, color.RGBA{255, 255, 255, 192})

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
			pipHeight = float64(c.Height() / 4)
		}
		if actualDeg%30 == 0 {
			pipWidth = 3
			pipHeight = float64(c.Height() / 2)
		}

		if pipWidth > 0 {
			// pip shows relative based on index (i) where negative is right of center, positive is left
			iRatio := float64(-i) / maxTurretDeg
			iX := float64(c.Width())/2 + iRatio*float64(c.Width())/2
			ebitenutil.DrawRect(c.image, iX-pipWidth/2, topY, pipWidth, pipHeight, color.RGBA{255, 255, 255, 255})

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
				ebitenutil.DebugPrintAt(c.image, pipDegStr, int(iX-pipWidth), int(float64(c.Height()/2)))
			}
		}
	}

	// heading indicator line
	hW, hH := 5.0, float64(c.Height()/2)
	ebitenutil.DrawRect(c.image, midX-hW/2, topY, hW, hH, color.RGBA{255, 255, 255, 255})
}

func (c *Compass) Texture() *ebiten.Image {
	return c.image
}
