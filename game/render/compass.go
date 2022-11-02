package render

import (
	"fmt"
	"image/color"

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
	turretAngle = -turretAngle

	// heading indicator line
	w, h := 3.0, float64(c.Height()/2)
	x1, y1 := float64(c.Width())/2, float64(0)
	ebitenutil.DrawRect(c.image, x1-w/2, y1, w, h, color.RGBA{255, 255, 255, 255})

	headingDeg := geom.Degrees(heading)
	headingStr := fmt.Sprintf("%0.0f", headingDeg)
	ebitenutil.DebugPrintAt(c.image, headingStr, int(x1), int(float64(c.Height()/2)))

	// turret indicator line
	relTurretRatio := turretAngle / geom.Pi
	tW := relTurretRatio * float64(c.Width()) / 2
	ebitenutil.DrawRect(c.image, x1, y1, tW, h, color.RGBA{255, 255, 255, 192})

	relTurretDeg := geom.Degrees(turretAngle)
	relTurretStr := fmt.Sprintf("%0.0f", relTurretDeg)
	ebitenutil.DebugPrintAt(c.image, relTurretStr, int(x1)+int(tW), 0)
}

func (c *Compass) Texture() *ebiten.Image {
	return c.image
}
