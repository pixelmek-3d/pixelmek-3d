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

func (c *Compass) Update(heading float64) {
	c.image.Clear()

	x1, y1 := float64(c.Width())/2, float64(0)
	x2, y2 := float64(c.Width())/2, float64(c.Height()/2)
	ebitenutil.DrawLine(c.image, x1, y1, x2, y2, color.RGBA{255, 255, 255, 255})

	headingStr := fmt.Sprintf("%0.0f", geom.Degrees(heading))
	ebitenutil.DebugPrintAt(c.image, headingStr, int(x2), int(y2))
}

func (c *Compass) Texture() *ebiten.Image {
	return c.image
}
