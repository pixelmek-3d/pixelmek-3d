package game

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"
)

func (g *Game) initInteractiveTypes() {
	g.interactiveSpriteTypes = map[SpriteType]struct{}{
		MechSpriteType:     {},
		VehicleSpriteType:  {},
		VTOLSpriteType:     {},
		InfantrySpriteType: {},
	}
}

func (g *Game) isInteractiveType(spriteType SpriteType) bool {
	if _, containsType := g.interactiveSpriteTypes[spriteType]; containsType {
		return true
	}
	return false
}

// loadHUD loads HUD elements
func (g *Game) loadHUD() {

	g.compass = render.NewCompass(g.fonts.HUDFont)

	g.altimeter = render.NewAltimeter(g.fonts.HUDFont)

	g.heat = render.NewHeatIndicator(g.fonts.HUDFont)

	g.radar = render.NewRadar(g.fonts.HUDFont)

	g.armament = render.NewArmament(g.fonts.HUDFont)

	g.throttle = render.NewThrottle(g.fonts.HUDFont)

	g.playerStatus = render.NewUnitStatus(g.fonts.HUDFont)
	g.targetStatus = render.NewUnitStatus(g.fonts.HUDFont)

	crosshairsSheet := getSpriteFromFile("hud/crosshairs_sheet.png")
	g.crosshairs = render.NewCrosshairs(crosshairsSheet, 1.0, 20, 10, 190)

	reticleSheet := getSpriteFromFile("hud/target_reticle.png")
	g.reticle = render.NewTargetReticle(1.0, reticleSheet)
}

// drawHUD draws HUD elements on the screen
func (g *Game) drawHUD(screen *ebiten.Image) {
	if !g.hudEnabled {
		return
	}

	// draw target reticle
	g.drawTargetReticle(screen)

	// draw crosshairs
	g.drawCrosshairs(screen)

	// draw compass with heading/turret orientation
	g.drawCompass(screen)

	// draw altimeter with altitude and pitch
	g.drawAltimeter(screen)

	// draw heat indicator
	g.drawHeatIndicator(screen)

	// draw radar with turret orientation
	g.drawRadar(screen)

	// draw armament display
	g.drawArmament(screen)

	// draw throttle display
	g.drawThrottle(screen)

	// draw player status display
	g.drawPlayerStatus(screen)

	// draw target status display
	g.drawTargetStatus(screen)
}

func (g *Game) drawPlayerStatus(screen *ebiten.Image) {
	if g.playerStatus == nil {
		return
	}

	statusScale := g.playerStatus.Scale() * g.renderScale * g.hudScale
	statusWidth, statusHeight := int(statusScale*float64(g.width)/5), int(statusScale*float64(g.height)/5)
	// FIXME: terrible arbitrary offsets
	sX, sY := 4*float64(g.width)/5-2*float64(statusWidth)/3, float64(g.height)-float64(statusHeight)
	sBounds := image.Rect(
		int(sX), int(sY), int(sX)+statusWidth, int(sY)+statusHeight,
	)
	g.playerStatus.Draw(screen, sBounds, &g.hudRGBA)
}

func (g *Game) drawTargetStatus(screen *ebiten.Image) {
	if g.targetStatus == nil {
		return
	}

	statusScale := g.targetStatus.Scale() * g.renderScale * g.hudScale
	statusWidth, statusHeight := int(statusScale*float64(g.width)/5), int(statusScale*float64(g.height)/5)
	// FIXME: terrible arbitrary offsets
	sX, sY := 0.0, float64(g.height)-float64(statusHeight)
	sBounds := image.Rect(
		int(sX), int(sY), int(sX)+statusWidth, int(sY)+statusHeight,
	)
	g.targetStatus.Draw(screen, sBounds, &g.hudRGBA)
}

func (g *Game) drawArmament(screen *ebiten.Image) {
	if g.armament == nil {
		return
	}

	armamentScale := g.armament.Scale() * g.renderScale * g.hudScale
	armamentWidth, armamentHeight := int(armamentScale*float64(g.width)/3), int(armamentScale*float64(3*g.height)/8)
	aX, aY := float64(g.width)-float64(armamentWidth), 0.0
	aBounds := image.Rect(
		int(aX), int(aY), int(aX)+armamentWidth, int(aY)+armamentHeight,
	)
	g.armament.Draw(screen, aBounds, &g.hudRGBA)
}

func (g *Game) drawCompass(screen *ebiten.Image) {
	if g.compass == nil {
		return
	}

	compassScale := g.compass.Scale() * g.renderScale * g.hudScale
	compassWidth, compassHeight := int(compassScale*float64(3*g.width)/10), int(compassScale*float64(g.height)/21)
	cX, cY := float64(g.width)/2-float64(compassWidth)/2, 0.0
	cBounds := image.Rect(
		int(cX), int(cY), int(cX)+compassWidth, int(cY)+compassHeight,
	)
	g.compass.Draw(screen, cBounds, &g.hudRGBA, g.player.Heading(), g.player.TurretAngle())
}

func (g *Game) drawAltimeter(screen *ebiten.Image) {
	if g.altimeter == nil {
		return
	}

	// convert Z position to meters of altitude
	altitude := g.player.PosZ() * model.METERS_PER_UNIT

	altScale := g.altimeter.Scale() * g.renderScale * g.hudScale
	altWidth, altHeight := int(altScale*float64(g.width)/24), int(altScale*float64(3*g.height)/12)
	aX, aY := 0.0, float64(g.height)/2-float64(altHeight)/2
	aBounds := image.Rect(
		int(aX), int(aY), int(aX)+altWidth, int(aY)+altHeight,
	)
	g.altimeter.Draw(screen, aBounds, &g.hudRGBA, altitude, g.player.Pitch())
}

func (g *Game) drawHeatIndicator(screen *ebiten.Image) {
	if g.heat == nil {
		return
	}

	// convert heat dissipation to seconds
	heat, maxHeat := g.player.Heat(), 100.0 // FIXME: add MaxHeat to model, determined based on # of heat sinks
	dissipationPerSec := g.player.HeatDissipation() * model.TICKS_PER_SECOND

	heatScale := g.heat.Scale() * g.renderScale * g.hudScale
	heatWidth, heatHeight := int(heatScale*float64(3*g.width)/10), int(heatScale*float64(g.height)/18)
	hX, hY := float64(g.width)/2-float64(heatWidth)/2, float64(g.height-heatHeight)
	hBounds := image.Rect(
		int(hX), int(hY), int(hX)+heatWidth, int(hY)+heatHeight,
	)
	g.heat.Draw(screen, hBounds, &g.hudRGBA, heat, maxHeat, dissipationPerSec)
}

func (g *Game) drawThrottle(screen *ebiten.Image) {
	if g.throttle == nil {
		return
	}

	// convert velocity from units per tick to kilometers per hour
	kphVelocity := g.player.Velocity() * model.VELOCITY_TO_KPH
	kphTgtVelocity := g.player.TargetVelocity() * model.VELOCITY_TO_KPH
	kphMax := g.player.MaxVelocity() * model.VELOCITY_TO_KPH

	throttleScale := g.throttle.Scale() * g.renderScale * g.hudScale
	throttleWidth, throttleHeight := int(throttleScale*float64(g.width)/8), int(throttleScale*float64(3*g.height)/8)
	tBounds := image.Rect(
		g.width-throttleWidth, g.height-throttleHeight,
		g.width, g.height,
	)
	g.throttle.Draw(screen, tBounds, &g.hudRGBA, kphVelocity, kphTgtVelocity, kphMax, kphMax/2)
}

func (g *Game) drawRadar(screen *ebiten.Image) {
	if g.radar == nil {
		return
	}

	radarScale := g.radar.Scale() * g.renderScale * g.hudScale
	radarWidth, radarHeight := int(radarScale*float64(g.width)/3), int(radarScale*float64(g.height)/3)
	rX, rY := 0.0, 0.0
	radarBounds := image.Rect(
		int(rX), int(rY), int(rX)+radarWidth, int(rY)+radarHeight,
	)
	g.radar.Draw(screen, radarBounds, &g.hudRGBA, g.player.Heading(), g.player.TurretAngle())
}

func (g *Game) drawCrosshairs(screen *ebiten.Image) {
	if g.crosshairs == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.ColorM.ScaleWithColor(g.hudRGBA)

	crosshairScale := g.crosshairs.Scale() * g.renderScale * g.hudScale
	op.GeoM.Scale(crosshairScale, crosshairScale)
	op.GeoM.Translate(
		float64(g.width)/2-float64(g.crosshairs.Width())*crosshairScale/2,
		float64(g.height)/2-float64(g.crosshairs.Height())*crosshairScale/2,
	)
	screen.DrawImage(g.crosshairs.Texture(), op)
}

func (g *Game) drawTargetReticle(screen *ebiten.Image) {
	if g.reticle == nil || g.player.Target() == nil {
		return
	}

	rScale := g.reticle.Scale() * g.renderScale * g.hudScale
	rOff := rScale * float64(g.reticle.Width()) / 2

	var op *ebiten.DrawImageOptions

	// setup some common draw modifications
	geoM := ebiten.GeoM{}
	geoM.Scale(rScale, rScale)

	colorM := ebiten.ColorM{}
	colorM.ScaleWithColor(g.hudRGBA)

	s := g.getSpriteFromEntity(g.player.Target())
	if s == nil {
		return
	}

	rect := s.ScreenRect()
	if rect == nil {
		return
	}

	minX, minY, maxX, maxY := float64(rect.Min.X), float64(rect.Min.Y), float64(rect.Max.X), float64(rect.Max.Y)

	// top left corner
	g.reticle.SetTextureFrame(0)
	op = &ebiten.DrawImageOptions{ColorM: colorM, GeoM: geoM}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(minX-rOff, minY-rOff)
	screen.DrawImage(g.reticle.Texture(), op)

	// top right corner
	g.reticle.SetTextureFrame(1)
	op = &ebiten.DrawImageOptions{ColorM: colorM, GeoM: geoM}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(maxX-rOff, minY-rOff)
	screen.DrawImage(g.reticle.Texture(), op)

	// bottom left corner
	g.reticle.SetTextureFrame(2)
	op = &ebiten.DrawImageOptions{ColorM: colorM, GeoM: geoM}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(minX-rOff, maxY-rOff)
	screen.DrawImage(g.reticle.Texture(), op)

	// bottom right corner
	g.reticle.SetTextureFrame(3)
	op = &ebiten.DrawImageOptions{ColorM: colorM, GeoM: geoM}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(maxX-rOff, maxY-rOff)
	screen.DrawImage(g.reticle.Texture(), op)
}
