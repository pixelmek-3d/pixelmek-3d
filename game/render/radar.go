package render

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
)

type Radar struct {
	HUDSprite
	fontRenderer *etxt.Renderer
	radarBlips   []*RadarBlip
}

type RadarBlip struct {
	Unit     model.Unit
	Angle    float64
	Distance float64
	IsTarget bool
}

//NewRadar creates a radar image to be rendered on demand
func NewRadar(font *Font) *Radar {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.Top, etxt.Left)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	r := &Radar{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
	}

	return r
}

func (r *Radar) updateFontSize(width, height int) {
	// set font size based on element size
	pxSize := float64(height) / 12
	if pxSize < 1 {
		pxSize = 1
	}

	fractSize, _ := efixed.FromFloat64(pxSize)
	r.fontRenderer.SetSizePxFract(fractSize)
}

func (r *Radar) SetRadarBlips(blips []*RadarBlip) {
	r.radarBlips = blips
}

func (r *Radar) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions, heading, turretAngle float64) {
	screen := hudOpts.Screen
	r.fontRenderer.SetTarget(screen)
	r.fontRenderer.SetColor(hudOpts.Color)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	r.updateFontSize(bW, bH)

	// turret angle appears opposite because it is relative to body heading which counts up counter clockwise
	// and offset by -90 degrees to make 0 degree turret angle as relative from the forward (up) position
	radarTurretAngle := -turretAngle - geom.HalfPi

	midX, midY := float64(bW)/2, float64(bH)/2
	radius := midX - 1
	if midY < midX {
		radius = midY - 1
	}

	// add target bounds offset for draw position
	midX += float64(bX)
	midY += float64(bY)

	// Draw radar range text
	radarStr := fmt.Sprintf("R:%0.1fkm", 1.0)
	r.fontRenderer.Draw(radarStr, 3, 3) // TODO: calculate better margin spacing

	// Draw radar circle outline
	// FIXME: when ebitengine v2.5 releases can draw circle outline using StrokeCircle
	//        - import "github.com/hajimehoshi/ebiten/v2/vector"
	//        - vector.StrokeCircle(r.image, float32(midX), float32(midY), float32(radius), float32(3), hudOpts.Color)
	oAlpha := uint8(hudOpts.Color.A / 5)
	ebitenutil.DrawCircle(screen, midX, midY, radius, color.RGBA{hudOpts.Color.R, hudOpts.Color.G, hudOpts.Color.B, oAlpha})

	// Draw turret angle reference lines
	// FIXME: when ebitengine v2.5 releases can draw lines with thickness using StrokeLine
	//        - vector.StrokeLine(r.image, float32(x1), float32(y1), float32(x2), float32(y2), float32(3), hudOpts.Color)
	quarterPi := geom.HalfPi / 2
	turretL := geom.LineFromAngle(midX, midY, radarTurretAngle-quarterPi, radius)
	turretR := geom.LineFromAngle(midX, midY, radarTurretAngle+quarterPi, radius)
	ebitenutil.DrawLine(screen, turretL.X1, turretL.Y1, turretL.X2, turretL.Y2, hudOpts.Color)
	ebitenutil.DrawLine(screen, turretR.X1, turretR.Y1, turretR.X2, turretR.Y2, hudOpts.Color)

	// Draw unit reference shape
	var refW, refH, refT float64 = 14, 5, 3 // TODO: calculate line thickness based on image size
	ebitenutil.DrawRect(screen, midX-refW/2, midY-refT/2, refW, refT, hudOpts.Color)
	ebitenutil.DrawRect(screen, midX-refW/2, midY-refH, refT, refH, hudOpts.Color)
	ebitenutil.DrawRect(screen, midX+refW/2-refT, midY-refH, refT, refH, hudOpts.Color)

	// Draw radar blips
	if len(r.radarBlips) > 0 {
		for _, blip := range r.radarBlips {
			// convert heading angle into relative radar angle where "up" is forward
			radarAngle := blip.Angle - geom.HalfPi

			// TODO: assumes radar is always 1km range
			radarDistancePx := radius * blip.Distance * model.METERS_PER_UNIT / 1000
			bLine := geom.LineFromAngle(midX, midY, radarAngle, radarDistancePx)

			if blip.IsTarget {
				// draw target square around lighter colored blip
				tAlpha := uint8(hudOpts.Color.A / 3)
				tColor := color.RGBA{R: hudOpts.Color.R, G: hudOpts.Color.G, B: hudOpts.Color.B, A: tAlpha}
				ebitenutil.DrawRect(screen, bLine.X2-6, bLine.Y2-6, 12, 12, tColor) // TODO: calculate thickness based on image size
			}

			ebitenutil.DrawRect(screen, bLine.X2-2, bLine.Y2-2, 4, 4, hudOpts.Color) // TODO: calculate thickness based on image size
		}
	}
}
