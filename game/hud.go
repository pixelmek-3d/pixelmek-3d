package game

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"
	"github.com/harbdog/raycaster-go/geom3d"
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

	g.playerStatus = render.NewUnitStatus(true, g.fonts.HUDFont)
	g.targetStatus = render.NewUnitStatus(false, g.fonts.HUDFont)

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

	screenW, screenH := screen.Size()
	marginX, marginY := screenW/50, screenH/50

	// draw target reticle
	g.drawTargetReticle(screen)

	// draw crosshairs
	g.drawCrosshairs(screen)

	// draw compass with heading/turret orientation
	g.drawCompass(screen, marginX, marginY)

	// draw altimeter with altitude and pitch
	g.drawAltimeter(screen, marginX, marginY)

	// draw heat indicator
	g.drawHeatIndicator(screen, marginX, marginY)

	// draw radar with turret orientation
	g.drawRadar(screen, marginX, marginY)

	// draw armament display
	g.drawArmament(screen, marginX, marginY)

	// draw throttle display
	g.drawThrottle(screen, marginX, marginY)

	// draw player status display
	g.drawPlayerStatus(screen, marginX, marginY)

	// draw target status display
	g.drawTargetStatus(screen, marginX, marginY)
}

func (g *Game) drawPlayerStatus(screen *ebiten.Image, marginX, marginY int) {
	if g.playerStatus == nil {
		return
	}

	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	statusScale := g.playerStatus.Scale() * g.renderScale * g.hudScale
	statusWidth, statusHeight := int(statusScale*float64(hudW)/5), int(statusScale*float64(hudH)/5)

	sX, sY := 4*float64(g.width)/5-2*float64(statusWidth)/3-float64(marginX), float64(g.height-statusHeight-marginY)
	sBounds := image.Rect(
		int(sX), int(sY), int(sX)+statusWidth, int(sY)+statusHeight,
	)
	g.playerStatus.Draw(screen, sBounds, &g.hudRGBA)
}

func (g *Game) drawTargetStatus(screen *ebiten.Image, marginX, marginY int) {
	if g.targetStatus == nil {
		return
	}

	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	statusScale := g.targetStatus.Scale() * g.renderScale * g.hudScale
	statusWidth, statusHeight := int(statusScale*float64(hudW)/5), int(statusScale*float64(hudH)/5)

	sX, sY := marginX, g.height-statusHeight-marginY
	sBounds := image.Rect(
		sX, sY, sX+statusWidth, sY+statusHeight,
	)

	targetUnit := g.targetStatus.Unit()
	if targetUnit != nil {
		pPos, pZ := g.player.Pos(), g.player.PosZ()
		tPos, tZ := targetUnit.Pos(), targetUnit.PosZ()
		targetLine := geom3d.Line3d{
			X1: pPos.X, Y1: pPos.Y, Z1: pZ,
			X2: tPos.X, Y2: tPos.Y, Z2: tZ,
		}
		targetDistance := targetLine.Distance() * model.METERS_PER_UNIT
		g.targetStatus.SetUnitDistance(targetDistance)
	}
	g.targetStatus.SetTargetReticle(g.reticle)
	g.targetStatus.Draw(screen, sBounds, &g.hudRGBA)
}

func (g *Game) drawArmament(screen *ebiten.Image, marginX, marginY int) {
	if g.armament == nil {
		return
	}

	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	armamentScale := g.armament.Scale() * g.renderScale * g.hudScale
	armamentWidth, armamentHeight := int(armamentScale*float64(hudW)/3), int(armamentScale*float64(3*hudH)/8)
	aX, aY := g.width-armamentWidth-marginX, marginY
	aBounds := image.Rect(
		aX, aY, aX+armamentWidth, aY+armamentHeight,
	)
	g.armament.Draw(screen, aBounds, &g.hudRGBA)
}

func (g *Game) drawCompass(screen *ebiten.Image, marginX, marginY int) {
	if g.compass == nil {
		return
	}

	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	compassScale := g.compass.Scale() * g.renderScale * g.hudScale
	compassWidth, compassHeight := int(compassScale*float64(3*hudW)/10), int(compassScale*float64(hudH)/21)
	cX, cY := float64(g.width)/2-float64(compassWidth)/2, float64(marginY)
	cBounds := image.Rect(
		int(cX), int(cY), int(cX)+compassWidth, int(cY)+compassHeight,
	)
	g.compass.Draw(screen, cBounds, &g.hudRGBA, g.player.Heading(), g.player.TurretAngle())
}

func (g *Game) drawAltimeter(screen *ebiten.Image, marginX, marginY int) {
	if g.altimeter == nil {
		return
	}

	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	// convert Z position to meters of altitude
	altitude := g.player.PosZ() * model.METERS_PER_UNIT

	altScale := g.altimeter.Scale() * g.renderScale * g.hudScale
	altWidth, altHeight := int(altScale*float64(hudW)/24), int(altScale*float64(3*hudH)/12)
	aX, aY := float64(marginX), float64(g.height)/2-float64(altHeight)/2-float64(marginY)
	aBounds := image.Rect(
		int(aX), int(aY), int(aX)+altWidth, int(aY)+altHeight,
	)
	g.altimeter.Draw(screen, aBounds, &g.hudRGBA, altitude, g.player.Pitch())
}

func (g *Game) drawHeatIndicator(screen *ebiten.Image, marginX, marginY int) {
	if g.heat == nil {
		return
	}

	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	// convert heat dissipation to seconds
	heat, maxHeat := g.player.Heat(), 100.0 // FIXME: add MaxHeat to model, determined based on # of heat sinks
	dissipationPerSec := g.player.HeatDissipation() * model.TICKS_PER_SECOND

	heatScale := g.heat.Scale() * g.renderScale * g.hudScale
	heatWidth, heatHeight := int(heatScale*float64(3*hudW)/10), int(heatScale*float64(hudH)/18)
	hX, hY := float64(g.width)/2-float64(heatWidth)/2, float64(g.height-heatHeight-marginY)
	hBounds := image.Rect(
		int(hX), int(hY), int(hX)+heatWidth, int(hY)+heatHeight,
	)
	g.heat.Draw(screen, hBounds, &g.hudRGBA, heat, maxHeat, dissipationPerSec)
}

func (g *Game) drawThrottle(screen *ebiten.Image, marginX, marginY int) {
	if g.throttle == nil {
		return
	}

	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	// convert velocity from units per tick to kilometers per hour
	kphVelocity := g.player.Velocity() * model.VELOCITY_TO_KPH
	kphTgtVelocity := g.player.TargetVelocity() * model.VELOCITY_TO_KPH
	kphMax := g.player.MaxVelocity() * model.VELOCITY_TO_KPH

	throttleScale := g.throttle.Scale() * g.renderScale * g.hudScale
	throttleWidth, throttleHeight := int(throttleScale*float64(hudW)/8), int(throttleScale*float64(3*hudH)/8)
	tBounds := image.Rect(
		g.width-throttleWidth-marginX, g.height-throttleHeight-marginY,
		g.width-marginX, g.height-marginY,
	)
	g.throttle.Draw(screen, tBounds, &g.hudRGBA, kphVelocity, kphTgtVelocity, kphMax, kphMax/2)
}

func (g *Game) drawRadar(screen *ebiten.Image, marginX, marginY int) {
	if g.radar == nil {
		return
	}

	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	radarScale := g.radar.Scale() * g.renderScale * g.hudScale
	radarWidth, radarHeight := int(radarScale*float64(hudW)/3), int(radarScale*float64(hudH)/3)
	rX, rY := marginX, marginY
	radarBounds := image.Rect(
		rX, rY, rX+radarWidth, rY+radarHeight,
	)
	g.radar.Draw(screen, radarBounds, &g.hudRGBA, g.player.Heading(), g.player.TurretAngle())
}

func (g *Game) drawCrosshairs(screen *ebiten.Image) {
	if g.crosshairs == nil {
		return
	}

	crosshairScale := g.crosshairs.Scale() * g.renderScale * g.hudScale
	g.crosshairs.Draw(screen, crosshairScale, &g.hudRGBA)
}

func (g *Game) drawTargetReticle(screen *ebiten.Image) {
	if g.reticle == nil || g.player.Target() == nil {
		return
	}

	s := g.getSpriteFromEntity(g.player.Target())
	if s == nil {
		return
	}

	rect := s.ScreenRect()
	if rect == nil {
		return
	}

	g.reticle.Draw(screen, *rect, &g.hudRGBA)
}
