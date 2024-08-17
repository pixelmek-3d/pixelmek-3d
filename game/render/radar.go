package render

import (
	"fmt"
	"image"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/tinne26/etxt"
)

var (
	_colorRadar        = _colorDefaultGreen
	_colorRadarOutline = color.NRGBA{R: 197, G: 145, B: 0, A: 255}
	radarRangeMeters   = 1000.0
)

type Radar struct {
	HUDSprite
	fontRenderer *etxt.Renderer
	mapLines     []*geom.Line
	radarBlips   []*RadarBlip
	navPoints    []*RadarNavPoint
	navLines     []*geom.Line
	position     *geom.Vector2
	heading      float64
	turretAngle  float64
	fovDegrees   float64
	radarRange   float64
}

type RadarBlip struct {
	Unit       model.Unit
	Angle      float64
	Heading    float64
	Distance   float64
	IsTarget   bool
	IsFriendly bool
}

type RadarNavPoint struct {
	NavPoint *model.NavPoint
	Angle    float64
	Distance float64
	IsTarget bool
}

// NewRadar creates a radar image to be rendered on demand
func NewRadar(font *Font) *Radar {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.Top, etxt.Left)
	renderer.SetColor(color.NRGBA{255, 255, 255, 255})

	r := &Radar{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
		radarRange:   radarRangeMeters / model.METERS_PER_UNIT,
	}

	return r
}

func (r *Radar) updateFontSize(_, height int) {
	// set font size based on element size
	pxSize := float64(height) / 15
	if pxSize < 1 {
		pxSize = 1
	}

	r.fontRenderer.SetSizePx(int(pxSize))
}

func (r *Radar) SetMapLines(lines []*geom.Line) {
	r.mapLines = lines
}

func (r *Radar) SetNavLines(lines []*geom.Line) {
	r.navLines = lines
}

func (r *Radar) SetNavPoints(radarNavPoints []*RadarNavPoint) {
	// sort nav points from furthest to closest from player position
	sort.Slice(radarNavPoints, func(i, j int) bool {
		return radarNavPoints[i].Distance > radarNavPoints[j].Distance
	})
	r.navPoints = radarNavPoints
}

func (r *Radar) SetRadarBlips(blips []*RadarBlip) {
	// sort blips from furthest to closest from player position to draw on top
	sort.Slice(blips, func(i, j int) bool {
		// player target blip always comes last
		switch {
		case blips[i].IsTarget:
			return false
		case blips[j].IsTarget:
			return true
		}
		return blips[i].Distance > blips[j].Distance
	})
	r.radarBlips = blips
}

func (r *Radar) SetValues(position *geom.Vector2, heading, turretAngle, fovDegrees float64) {
	r.position = position
	r.heading = heading
	r.turretAngle = turretAngle
	r.fovDegrees = fovDegrees
}

func (r *Radar) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen
	r.fontRenderer.SetTarget(screen)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	r.updateFontSize(bW, bH)

	// turret angle appears opposite because it is relative to body heading which counts up counter clockwise
	// and offset by -90 degrees to make 0 degree turret angle as relative from the forward (up) position
	relTurretAngle := -model.AngleDistance(r.heading, r.turretAngle)
	radarTurretAngle := relTurretAngle - geom.HalfPi

	midX, midY := float64(bW)/2, float64(bH)/2
	radius := midX - 1
	if midY < midX {
		radius = midY - 1
	}

	// add target bounds offset for draw position
	midX += float64(bX)
	midY += float64(bY)

	// Draw radar range text
	rColor := hudOpts.HudColor(_colorRadar)
	r.fontRenderer.SetColor(rColor)

	radarStr := fmt.Sprintf("R:%0.1fkm", 1.0)
	r.fontRenderer.Draw(radarStr, bX, bY)

	// Draw radar circle outline
	oColor := hudOpts.HudColor(_colorRadarOutline)

	var oT float32 = 2 // TODO: calculate line thickness based on image height
	oAlpha := uint8(int(oColor.A) / 5)
	vector.StrokeCircle(screen, float32(midX), float32(midY), float32(radius), oT, color.NRGBA{oColor.R, oColor.G, oColor.B, oAlpha}, false)

	// Draw any walls/boundaries within the radar range using lines that make up the map wall boundaries
	radarHudSizeFactor := radius / r.radarRange
	wColor := hudOpts.HudColor(_colorRadarOutline)
	for _, borderLine := range r.mapLines {
		r.drawRadarLine(screen, borderLine, midX, midY, radarHudSizeFactor, oT, wColor)
	}

	// Draw turret angle reference lines
	fovAngle := geom.Radians(r.fovDegrees)
	turretL := geom.LineFromAngle(midX, midY, radarTurretAngle-fovAngle/2, radius)
	turretR := geom.LineFromAngle(midX, midY, radarTurretAngle+fovAngle/2, radius)
	vector.StrokeLine(screen, float32(turretL.X1), float32(turretL.Y1), float32(turretL.X2), float32(turretL.Y2), oT, oColor, false)
	vector.StrokeLine(screen, float32(turretR.X1), float32(turretR.Y1), float32(turretR.X2), float32(turretR.Y2), oT, oColor, false)

	// Draw unit reference shape
	var refW, refH, refT float32 = 14, 5, 3 // TODO: calculate line thickness based on image size
	vector.DrawFilledRect(screen, float32(midX)-refW/2, float32(midY)-refT/2, refW, refT, rColor, false)
	vector.DrawFilledRect(screen, float32(midX)-refW/2, float32(midY)-refH, refT, refH, rColor, false)
	vector.DrawFilledRect(screen, float32(midX)+refW/2-refT, float32(midY)-refH, refT, refH, rColor, false)

	// Draw nav points
	nColor := hudOpts.HudColor(_colorRadarOutline)

	for _, nav := range r.navPoints {
		// convert heading angle into relative radar angle where "up" is forward
		radarAngle := nav.Angle - geom.HalfPi

		radarDistancePx := nav.Distance * radarHudSizeFactor
		nLine := geom.LineFromAngle(midX, midY, radarAngle, radarDistancePx)

		if nav.IsTarget {
			// draw target nav circle around lighter colored nav
			tAlpha := uint8(int(nColor.A) / 3)
			tColor := color.NRGBA{R: nColor.R, G: nColor.G, B: nColor.B, A: tAlpha}

			var navTargetRadius float32 = 8
			if nav.NavPoint.Visited() {
				navTargetRadius = 4
			}
			vector.DrawFilledCircle(screen, float32(nLine.X2), float32(nLine.Y2), navTargetRadius, tColor, false) // TODO: calculate thickness based on image size
		}

		var navRadius float32 = 3
		if nav.NavPoint.Visited() {
			navRadius = 1
		}
		vector.DrawFilledCircle(screen, float32(nLine.X2), float32(nLine.Y2), navRadius, nColor, false) // TODO: calculate thickness based on image size
	}

	// Draw radar blips
	eColor := hudOpts.HudColor(_colorEnemy)
	fColor := hudOpts.HudColor(_colorFriendly)

	for _, blip := range r.radarBlips {
		// convert direction angle into relative radar angle where "up" is forward
		radarAngle := blip.Angle - geom.HalfPi

		radarDistancePx := blip.Distance * radarHudSizeFactor
		bLine := geom.LineFromAngle(midX, midY, radarAngle, radarDistancePx)

		var bColor color.NRGBA
		if blip.IsFriendly {
			bColor = fColor
		} else {
			bColor = eColor
		}

		// convert blip unit heading into relative radar angle
		radarHeading := blip.Heading - geom.HalfPi

		if blip.IsTarget {
			// draw target square around lighter colored blip
			tAlpha := uint8(int(bColor.A) / 3)
			tColor := color.NRGBA{R: bColor.R, G: bColor.G, B: bColor.B, A: tAlpha}
			vector.DrawFilledRect(screen, float32(bLine.X2-6), float32(bLine.Y2-6), 12, 12, tColor, false) // TODO: calculate thickness based on image size

			hLine := geom.LineFromAngle(bLine.X2, bLine.Y2, radarHeading, 10)
			vector.StrokeLine(screen, float32(hLine.X1), float32(hLine.Y1), float32(hLine.X2), float32(hLine.Y2), 3, bColor, false)
		} else {
			hLine := geom.LineFromAngle(bLine.X2, bLine.Y2, radarHeading, 8)
			vector.StrokeLine(screen, float32(hLine.X1), float32(hLine.Y1), float32(hLine.X2), float32(hLine.Y2), 2, bColor, false)
		}

		if blip.IsFriendly {
			// differentiate friendly by not filling in the radar blip box
			vector.StrokeRect(screen, float32(bLine.X2)-2, float32(bLine.Y2-2), 4, 4, 2, bColor, false)
		} else {
			vector.DrawFilledRect(screen, float32(bLine.X2)-2, float32(bLine.Y2-2), 4, 4, bColor, false) // TODO: calculate thickness based on image size
		}
	}
}

func (r *Radar) drawRadarLine(dst *ebiten.Image, line *geom.Line, centerX, centerY, hudSizeFactor float64, lineWidth float32, clr color.Color) {
	posX, posY := r.position.X, r.position.Y
	// quick range check for nearby wall cells
	if !(model.PointInProximity(r.radarRange, posX, posY, line.X1, line.Y1) ||
		model.PointInProximity(r.radarRange, posX, posY, line.X2, line.Y2)) {
		return
	}

	// determine distance to wall line, convert to relative radar angle and draw
	line1 := geom.Line{X1: posX, Y1: posY, X2: line.X1, Y2: line.Y1}
	angle1 := r.heading - line1.Angle() - geom.HalfPi
	dist1 := line1.Distance()

	line2 := geom.Line{X1: posX, Y1: posY, X2: line.X2, Y2: line.Y2}
	angle2 := r.heading - line2.Angle() - geom.HalfPi
	dist2 := line2.Distance()

	if dist1 > r.radarRange || dist2 > r.radarRange {
		return
	}

	rLine1 := geom.LineFromAngle(centerX, centerY, angle1, dist1*hudSizeFactor)
	rLine2 := geom.LineFromAngle(centerX, centerY, angle2, dist2*hudSizeFactor)

	vector.StrokeLine(dst, float32(rLine1.X2), float32(rLine1.Y2), float32(rLine2.X2), float32(rLine2.Y2), lineWidth, clr, false)
}
