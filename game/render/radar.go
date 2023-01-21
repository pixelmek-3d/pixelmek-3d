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

var (
	_colorRadar        = _colorDefaultGreen
	_colorRadarOutline = color.RGBA{R: 197, G: 145, B: 0, A: 255}
	radarRangeMeters   = 1000.0
)

type Radar struct {
	HUDSprite
	fontRenderer *etxt.Renderer
	mapLines     []*geom.Line
	radarBlips   []*RadarBlip
	navPoints    []*RadarNavPoint
}

type RadarBlip struct {
	Unit     model.Unit
	Angle    float64
	Distance float64
	IsTarget bool
}

type RadarNavPoint struct {
	NavPoint *model.NavPoint
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

func (r *Radar) SetMapLines(lines []*geom.Line) {
	r.mapLines = lines
}

func (r *Radar) SetNavPoints(radarNavPoints []*RadarNavPoint) {
	r.navPoints = radarNavPoints
}

func (r *Radar) SetRadarBlips(blips []*RadarBlip) {
	r.radarBlips = blips
}

func (r *Radar) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions, position *geom.Vector2, heading, turretAngle, fovDegrees float64) {
	screen := hudOpts.Screen
	r.fontRenderer.SetTarget(screen)

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
	rColor := _colorRadar
	if hudOpts.UseCustomColor {
		rColor = hudOpts.Color
	}
	r.fontRenderer.SetColor(rColor)

	radarStr := fmt.Sprintf("R:%0.1fkm", 1.0)
	r.fontRenderer.Draw(radarStr, 3, 3) // TODO: calculate better margin spacing

	// Draw radar circle outline
	oColor := _colorRadarOutline
	if hudOpts.UseCustomColor {
		oColor = hudOpts.Color
	}

	// FIXME: when ebitengine v2.5 releases can draw circle outline using StrokeCircle
	//        - import "github.com/hajimehoshi/ebiten/v2/vector"
	//        - vector.StrokeCircle(r.image, float32(midX), float32(midY), float32(radius), float32(3), hudOpts.Color)
	oAlpha := uint8(oColor.A / 5)
	ebitenutil.DrawCircle(screen, midX, midY, radius, color.RGBA{oColor.R, oColor.G, oColor.B, oAlpha})

	// Draw any walls/boundaries within the radar range using lines that make up the map wall boundaries
	posX, posY := position.X, position.Y
	radarRange := radarRangeMeters / model.METERS_PER_UNIT
	radarHudSizeFactor := radius / radarRange
	for _, borderLine := range r.mapLines {
		// quick range check for nearby wall cells
		if !(model.PointInProximity(radarRange, posX, posY, borderLine.X1, borderLine.Y1) ||
			model.PointInProximity(radarRange, posX, posY, borderLine.X2, borderLine.Y2)) {
			continue
		}

		wColor := _colorRadarOutline
		if hudOpts.UseCustomColor {
			wColor = hudOpts.Color
		}

		// determine distance to wall line, convert to relative radar angle and draw
		line1 := geom.Line{X1: posX, Y1: posY, X2: borderLine.X1, Y2: borderLine.Y1}
		angle1 := heading - line1.Angle() - geom.HalfPi
		dist1 := line1.Distance()

		line2 := geom.Line{X1: posX, Y1: posY, X2: borderLine.X2, Y2: borderLine.Y2}
		angle2 := heading - line2.Angle() - geom.HalfPi
		dist2 := line2.Distance()

		if dist1 > radarRange || dist2 > radarRange {
			continue
		}

		rLine1 := geom.LineFromAngle(midX, midY, angle1, dist1*radarHudSizeFactor)
		rLine2 := geom.LineFromAngle(midX, midY, angle2, dist2*radarHudSizeFactor)

		ebitenutil.DrawLine(screen, rLine1.X2, rLine1.Y2, rLine2.X2, rLine2.Y2, wColor)
	}

	// Draw turret angle reference lines
	// FIXME: when ebitengine v2.5 releases can draw lines with thickness using StrokeLine
	//        - vector.StrokeLine(r.image, float32(x1), float32(y1), float32(x2), float32(y2), float32(3), hudOpts.Color)
	fovAngle := geom.Radians(fovDegrees)
	turretL := geom.LineFromAngle(midX, midY, radarTurretAngle-fovAngle/2, radius)
	turretR := geom.LineFromAngle(midX, midY, radarTurretAngle+fovAngle/2, radius)
	ebitenutil.DrawLine(screen, turretL.X1, turretL.Y1, turretL.X2, turretL.Y2, oColor)
	ebitenutil.DrawLine(screen, turretR.X1, turretR.Y1, turretR.X2, turretR.Y2, oColor)

	// Draw unit reference shape
	var refW, refH, refT float64 = 14, 5, 3 // TODO: calculate line thickness based on image size
	ebitenutil.DrawRect(screen, midX-refW/2, midY-refT/2, refW, refT, rColor)
	ebitenutil.DrawRect(screen, midX-refW/2, midY-refH, refT, refH, rColor)
	ebitenutil.DrawRect(screen, midX+refW/2-refT, midY-refH, refT, refH, rColor)

	// Draw nav points
	nColor := _colorRadarOutline
	if hudOpts.UseCustomColor {
		nColor = hudOpts.Color
	}

	for _, nav := range r.navPoints {
		// convert heading angle into relative radar angle where "up" is forward
		radarAngle := nav.Angle - geom.HalfPi

		radarDistancePx := nav.Distance * radarHudSizeFactor
		nLine := geom.LineFromAngle(midX, midY, radarAngle, radarDistancePx)

		if nav.IsTarget {
			// draw target nav circle around lighter colored nav
			tAlpha := uint8(nColor.A / 3)
			tColor := color.RGBA{R: nColor.R, G: nColor.G, B: nColor.B, A: tAlpha}
			ebitenutil.DrawCircle(screen, nLine.X2, nLine.Y2, 8, tColor) // TODO: calculate thickness based on image size
		}

		ebitenutil.DrawCircle(screen, nLine.X2, nLine.Y2, 3, nColor) // TODO: calculate thickness based on image size
	}

	// Draw radar blips
	bColor := _colorEnemy
	if hudOpts.UseCustomColor {
		bColor = hudOpts.Color
	}

	for _, blip := range r.radarBlips {
		// convert heading angle into relative radar angle where "up" is forward
		radarAngle := blip.Angle - geom.HalfPi

		radarDistancePx := blip.Distance * radarHudSizeFactor
		bLine := geom.LineFromAngle(midX, midY, radarAngle, radarDistancePx)

		if blip.IsTarget {
			// draw target square around lighter colored blip
			tAlpha := uint8(bColor.A / 3)
			tColor := color.RGBA{R: bColor.R, G: bColor.G, B: bColor.B, A: tAlpha}
			ebitenutil.DrawRect(screen, bLine.X2-6, bLine.Y2-6, 12, 12, tColor) // TODO: calculate thickness based on image size
		}

		ebitenutil.DrawRect(screen, bLine.X2-2, bLine.Y2-2, 4, 4, bColor) // TODO: calculate thickness based on image size
	}
}
