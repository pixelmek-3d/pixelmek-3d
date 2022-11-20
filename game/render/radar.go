package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/tinne26/etxt"
)

type Radar struct {
	HUDSprite
	image        *ebiten.Image
	fontRenderer *etxt.Renderer
}

//NewRadar creates a radar image to be rendered on demand
func NewRadar(width, height int, font *Font) *Radar {
	img := ebiten.NewImage(width, height)

	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetSizePx(16)
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.Top, etxt.Left)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	r := &Radar{
		HUDSprite:    NewHUDSprite(img, 1.0),
		image:        img,
		fontRenderer: renderer,
	}

	return r
}

func (r *Radar) Update(heading, turretAngle float64) {
	r.image.Clear()

	r.fontRenderer.SetTarget(r.image)

	// turret angle appears opposite because it is relative to body heading which counts up counter clockwise
	// and offset by -90 degrees to make 0 degree turret angle as relative from the forward (up) position
	radarTurretAngle := -turretAngle - geom.HalfPi

	midX, midY := float64(r.Width())/2, float64(r.Height())/2
	radius := midX - 1
	if midY < midX {
		radius = midY - 1
	}

	// Draw radar range text
	radarStr := fmt.Sprintf("R:%0.1fkm", 1.0)
	r.fontRenderer.Draw(radarStr, 3, 3) // TODO: calculate better margin spacing

	// Draw radar circle outline
	// FIXME: when ebitengine v2.5 releases can draw circle outline using StrokeCircle
	//        - import "github.com/hajimehoshi/ebiten/v2/vector"
	//        - vector.StrokeCircle(r.image, float32(midX), float32(midY), float32(radius), float32(3), clr)
	ebitenutil.DrawCircle(r.image, midX, midY, radius, color.RGBA{255, 255, 255, 48})

	// Draw turret angle reference lines
	// FIXME: when ebitengine v2.5 releases can draw lines with thickness using StrokeLine
	//        - vector.StrokeLine(r.image, float32(x1), float32(y1), float32(x2), float32(y2), float32(3), clr)
	quarterPi := geom.HalfPi / 2
	turretL := geom.LineFromAngle(midX, midY, radarTurretAngle-quarterPi, radius)
	turretR := geom.LineFromAngle(midX, midY, radarTurretAngle+quarterPi, radius)
	ebitenutil.DrawLine(r.image, turretL.X1, turretL.Y1, turretL.X2, turretL.Y2, color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawLine(r.image, turretR.X1, turretR.Y1, turretR.X2, turretR.Y2, color.RGBA{255, 255, 255, 255})

	// Draw unit reference shape
	var refW, refH, refT float64 = 14, 5, 3 // TODO: calculate line width based on image width
	ebitenutil.DrawRect(r.image, midX-refW/2, midY-refT/2, refW, refT, color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(r.image, midX-refW/2, midY-refH, refT, refH, color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(r.image, midX+refW/2-refT, midY-refH, refT, refH, color.RGBA{255, 255, 255, 255})
}

func (r *Radar) Texture() *ebiten.Image {
	return r.image
}
