package game

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
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
	g.jets = render.NewJumpJetIndicator(g.fonts.HUDFont)

	g.radar = render.NewRadar(g.fonts.HUDFont)
	g.radar.SetMapLines(g.collisionMap)

	g.armament = render.NewArmament(g.fonts.HUDFont)

	g.throttle = render.NewThrottle(g.fonts.HUDFont)

	g.playerStatus = render.NewUnitStatus(true, g.fonts.HUDFont)
	g.targetStatus = render.NewUnitStatus(false, g.fonts.HUDFont)
	g.navStatus = render.NewNavStatus(g.fonts.HUDFont)

	crosshairsSheet := getSpriteFromFile("hud/crosshairs_sheet.png")
	g.crosshairs = render.NewCrosshairs(crosshairsSheet, 1.0, 20, 10, 190)

	tgtReticleSheet := getSpriteFromFile("hud/target_reticle.png")
	g.targetReticle = render.NewTargetReticle(1.0, tgtReticleSheet)

	navReticleSheet := getSpriteFromFile("hud/nav_reticle.png")
	g.navReticle = render.NewNavReticle(1.0, navReticleSheet)
}

// drawHUD draws HUD elements on the screen
func (g *Game) drawHUD(screen *ebiten.Image) {
	if !g.hudEnabled {
		return
	}

	screenW, screenH := screen.Size()
	marginX, marginY := screenW/50, screenH/50

	hudOpts := &render.DrawHudOptions{
		Screen:         screen,
		RenderScale:    g.renderScale,
		MarginX:        marginX,
		MarginY:        marginY,
		UseCustomColor: g.hudUseCustomColor,
		Color:          g.hudRGBA,
	}

	// draw target reticle
	g.drawTargetReticle(hudOpts)

	// draw nav reticle
	g.drawNavReticle(hudOpts)

	// draw crosshairs
	g.drawCrosshairs(hudOpts)

	// draw compass with heading/turret orientation
	g.drawCompass(hudOpts)

	// draw altimeter with altitude and pitch
	g.drawAltimeter(hudOpts)

	// draw heat indicator
	g.drawHeatIndicator(hudOpts)

	// draw jump jet indicator
	g.drawJumpJetIndicator(hudOpts)

	// draw radar with turret orientation
	g.drawRadar(hudOpts)

	// draw armament display
	g.drawArmament(hudOpts)

	// draw throttle display
	g.drawThrottle(hudOpts)

	// draw player status display
	g.drawPlayerStatus(hudOpts)

	// draw target status display
	g.drawTargetStatus(hudOpts)

	// draw nav status display
	g.drawNavStatus(hudOpts)
}

func (g *Game) drawPlayerStatus(hudOpts *render.DrawHudOptions) {
	if g.playerStatus == nil {
		return
	}

	marginX, marginY := hudOpts.MarginX, hudOpts.MarginY
	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	statusScale := g.playerStatus.Scale() * g.renderScale * g.hudScale
	statusWidth, statusHeight := int(statusScale*float64(hudW)/5), int(statusScale*float64(hudH)/5)

	sX, sY := 4*float64(g.width)/5-2*float64(statusWidth)/3-float64(marginX), float64(g.height-statusHeight-marginY)
	sBounds := image.Rect(
		int(sX), int(sY), int(sX)+statusWidth, int(sY)+statusHeight,
	)
	g.playerStatus.SetUnit(g.player.sprite)
	g.playerStatus.Draw(sBounds, hudOpts)
}

func (g *Game) drawTargetStatus(hudOpts *render.DrawHudOptions) {
	if g.targetStatus == nil || g.player.Target() == nil {
		return
	}

	marginX, marginY := hudOpts.MarginX, hudOpts.MarginY
	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	statusScale := g.targetStatus.Scale() * g.renderScale * g.hudScale
	statusWidth, statusHeight := int(statusScale*float64(hudW)/5), int(statusScale*float64(hudH)/5)

	sX, sY := marginX, g.height-statusHeight-marginY
	sBounds := image.Rect(
		sX, sY, sX+statusWidth, sY+statusHeight,
	)

	targetEntity := g.player.Target()
	targetUnit := g.targetStatus.Unit()
	if targetUnit == nil || targetUnit.Entity != targetEntity {
		targetUnit = g.getSpriteFromEntity(targetEntity)
	}

	if targetUnit != nil {
		targetDistance := model.EntityDistance(g.player, targetUnit.Entity) - targetUnit.CollisionRadius() - g.player.CollisionRadius()
		distanceMeters := targetDistance * model.METERS_PER_UNIT

		g.targetStatus.SetUnitDistance(distanceMeters)
	}

	g.targetStatus.SetTargetReticle(g.targetReticle)
	g.targetStatus.SetUnit(targetUnit)
	g.targetStatus.Draw(sBounds, hudOpts)
}

func (g *Game) drawNavStatus(hudOpts *render.DrawHudOptions) {
	if g.navStatus == nil || g.player.navPoint == nil || g.player.Target() != nil {
		return
	}

	marginX, marginY := hudOpts.MarginX, hudOpts.MarginY
	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	statusScale := g.navStatus.Scale() * g.renderScale * g.hudScale
	statusWidth, statusHeight := int(statusScale*float64(hudW)/5), int(statusScale*float64(hudH)/5)

	sX, sY := marginX, g.height-statusHeight-marginY
	sBounds := image.Rect(
		sX, sY, sX+statusWidth, sY+statusHeight,
	)

	navPoint := g.player.navPoint.NavPoint
	pPos, nPos := g.player.Pos(), navPoint.Pos()
	navLine := geom.Line{
		X1: pPos.X, Y1: pPos.Y,
		X2: nPos.X, Y2: nPos.Y,
	}
	navDistance := navLine.Distance() * model.METERS_PER_UNIT

	g.navStatus.SetNavDistance(navDistance)
	g.navStatus.SetNavPoint(navPoint)
	g.navStatus.Draw(sBounds, hudOpts)
}

func (g *Game) drawArmament(hudOpts *render.DrawHudOptions) {
	if g.armament == nil {
		return
	}

	marginX, marginY := hudOpts.MarginX, hudOpts.MarginY
	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	armamentScale := g.armament.Scale() * g.renderScale * g.hudScale
	armamentWidth, armamentHeight := int(armamentScale*float64(hudW)/3), int(armamentScale*float64(3*hudH)/8)
	aX, aY := g.width-armamentWidth-marginX, marginY
	aBounds := image.Rect(
		aX, aY, aX+armamentWidth, aY+armamentHeight,
	)

	weaponOrGroupIndex := g.player.selectedWeapon
	if g.player.fireMode == model.GROUP_FIRE {
		weaponOrGroupIndex = g.player.selectedGroup
	}
	g.armament.SetWeaponGroups(g.player.weaponGroups)
	g.armament.SetSelectedWeapon(weaponOrGroupIndex, g.player.fireMode)
	g.armament.Draw(aBounds, hudOpts)
}

func (g *Game) drawCompass(hudOpts *render.DrawHudOptions) {
	if g.compass == nil {
		return
	}

	marginX, marginY := hudOpts.MarginX, hudOpts.MarginY
	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	compassScale := g.compass.Scale() * g.renderScale * g.hudScale
	compassWidth, compassHeight := int(compassScale*float64(3*hudW)/10), int(compassScale*float64(hudH)/21)
	cX, cY := float64(g.width)/2-float64(compassWidth)/2, float64(marginY)
	cBounds := image.Rect(
		int(cX), int(cY), int(cX)+compassWidth, int(cY)+compassHeight,
	)

	playerPos := g.player.Pos()

	if g.player.Target() == nil {
		g.compass.SetTargetEnabled(false)
	} else {
		targetPos := g.player.Target().Pos()
		tLine := geom.Line{
			X1: playerPos.X, Y1: playerPos.Y,
			X2: targetPos.X, Y2: targetPos.Y,
		}
		tAngle := tLine.Angle()

		g.compass.SetTargetEnabled(true)
		g.compass.SetTargetHeading(tAngle)
	}

	if g.player.navPoint == nil {
		g.compass.SetNavEnabled(false)
	} else {
		navPos := g.player.navPoint.Pos()
		tLine := geom.Line{
			X1: playerPos.X, Y1: playerPos.Y,
			X2: navPos.X, Y2: navPos.Y,
		}
		nAngle := tLine.Angle()

		g.compass.SetNavEnabled(true)
		g.compass.SetNavHeading(nAngle)
	}

	g.compass.Draw(cBounds, hudOpts, g.player.Heading(), g.player.TurretAngle())
}

func (g *Game) drawAltimeter(hudOpts *render.DrawHudOptions) {
	if g.altimeter == nil {
		return
	}

	marginX, marginY := hudOpts.MarginX, hudOpts.MarginY
	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	// convert Z position to meters of altitude
	altitude := g.player.PosZ() * model.METERS_PER_UNIT

	altScale := g.altimeter.Scale() * g.renderScale * g.hudScale
	altWidth, altHeight := int(altScale*float64(hudW)/24), int(altScale*float64(3*hudH)/12)
	aX, aY := float64(marginX), float64(g.height)/2-float64(altHeight)/2-float64(marginY)
	aBounds := image.Rect(
		int(aX), int(aY), int(aX)+altWidth, int(aY)+altHeight,
	)
	g.altimeter.Draw(aBounds, hudOpts, altitude, g.player.Pitch())
}

func (g *Game) drawHeatIndicator(hudOpts *render.DrawHudOptions) {
	if g.heat == nil {
		return
	}

	marginX, marginY := hudOpts.MarginX, hudOpts.MarginY
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
	g.heat.Draw(hBounds, hudOpts, heat, maxHeat, dissipationPerSec)
}

func (g *Game) drawThrottle(hudOpts *render.DrawHudOptions) {
	if g.throttle == nil {
		return
	}

	marginX, marginY := hudOpts.MarginX, hudOpts.MarginY
	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	// convert velocity from units per tick to kilometers per hour
	kphVelocity := g.player.Velocity() * model.VELOCITY_TO_KPH
	kphVelocityZ := g.player.VelocityZ() * model.VELOCITY_TO_KPH
	kphTgtVelocity := g.player.TargetVelocity() * model.VELOCITY_TO_KPH
	kphMax := g.player.MaxVelocity() * model.VELOCITY_TO_KPH

	throttleScale := g.throttle.Scale() * g.renderScale * g.hudScale
	throttleWidth, throttleHeight := int(throttleScale*float64(hudW)/8), int(throttleScale*float64(3*hudH)/8)
	tBounds := image.Rect(
		g.width-throttleWidth-marginX, g.height-throttleHeight-marginY,
		g.width-marginX, g.height-marginY,
	)
	g.throttle.Draw(tBounds, hudOpts, kphVelocity, kphTgtVelocity, kphVelocityZ, kphMax, kphMax/2)
}

func (g *Game) drawJumpJetIndicator(hudOpts *render.DrawHudOptions) {
	if g.jets == nil {
		return
	}

	if g.player == nil || g.player.Unit.JumpJets() == 0 {
		return
	}

	marginX, marginY := hudOpts.MarginX, hudOpts.MarginY
	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	jDuration := g.player.Unit.JumpJetDuration()
	jMaxDuration := g.player.Unit.MaxJumpJetDuration()

	jetsScale := g.jets.Scale() * g.renderScale * g.hudScale
	jetsWidth, jetsHeight := int(jetsScale*float64(hudW)/12), int(jetsScale*float64(3*hudH)/18)
	hX, hY := float64(g.width)/5+2*float64(marginX), float64(g.height-jetsHeight-marginY)
	jBounds := image.Rect(
		int(hX), int(hY), int(hX)+jetsWidth, int(hY)+jetsHeight,
	)
	g.jets.Draw(jBounds, hudOpts, jDuration, jMaxDuration)
}

func (g *Game) drawRadar(hudOpts *render.DrawHudOptions) {
	if g.radar == nil {
		return
	}

	marginX, marginY := hudOpts.MarginX, hudOpts.MarginY
	hudW, hudH := g.width-marginX*2, g.height-marginY*2

	radarScale := g.radar.Scale() * g.renderScale * g.hudScale
	radarWidth, radarHeight := int(radarScale*float64(hudW)/3), int(radarScale*float64(hudH)/3)
	rX, rY := marginX, marginY
	radarBounds := image.Rect(
		rX, rY, rX+radarWidth, rY+radarHeight,
	)

	// find all units and nav points within range to draw as blips
	maxDistanceMeters := 1000.0 // TODO: set in Radar object and game config
	maxDistanceUnits := maxDistanceMeters / model.METERS_PER_UNIT

	radarBlips := make([]*render.RadarBlip, 0, 128)
	rNavPoints := make([]*render.RadarNavPoint, 0, len(g.mission.NavPoints))

	playerPos := g.player.Pos()
	playerAngle := g.player.Heading()
	playerTarget := g.player.Target()

	var playerNav *model.NavPoint
	if g.player.navPoint != nil {
		playerNav = g.player.navPoint.NavPoint
	}

	// discover nav points that are in range
	navCount := 0
	for _, nav := range g.mission.NavPoints {
		navPos := nav.Pos()
		navLine := geom.Line{
			X1: playerPos.X, Y1: playerPos.Y,
			X2: navPos.X, Y2: navPos.Y,
		}

		navIsTarget := playerNav == nav
		navDistance := navLine.Distance()
		if navDistance > maxDistanceUnits {
			if navIsTarget {
				// if current nav point out of radar range, draw just outside edge
				navDistance = maxDistanceUnits + 1
			} else {
				continue
			}
		}

		// determine angle of unit relative from player heading
		relAngle := playerAngle - navLine.Angle()
		rNav := &render.RadarNavPoint{
			NavPoint: nav, Distance: navDistance, Angle: relAngle, IsTarget: navIsTarget,
		}

		rNavPoints = append(rNavPoints, rNav)
		navCount++
	}

	// discover blips that are in range
	blipCount := 0
	for _, spriteMap := range g.sprites.sprites {
		spriteMap.Range(func(k, _ interface{}) bool {
			spriteInterface := k.(raycaster.Sprite)
			entity := getEntityFromInterface(spriteInterface)
			unit := model.EntityUnit(entity)
			if unit == nil {
				return true
			}

			unitPos := unit.Pos()
			unitLine := geom.Line{
				X1: playerPos.X, Y1: playerPos.Y,
				X2: unitPos.X, Y2: unitPos.Y,
			}

			unitIsTarget := playerTarget == entity
			unitDistance := unitLine.Distance()
			if unitDistance > maxDistanceUnits {
				if unitIsTarget {
					// if current target out of radar range, draw just outside edge
					unitDistance = maxDistanceUnits + 1
				} else {
					return true
				}
			}

			// determine angle of unit relative from player heading
			relAngle := playerAngle - unitLine.Angle()
			blip := &render.RadarBlip{
				Unit: unit, Distance: unitDistance, Angle: relAngle, IsTarget: unitIsTarget,
			}

			radarBlips = append(radarBlips, blip)
			blipCount++
			return true
		})
	}

	g.radar.SetNavPoints(rNavPoints[:navCount])
	g.radar.SetRadarBlips(radarBlips[:blipCount])

	cameraViewDegrees := g.fovDegrees / g.camera.FovDepth()
	g.radar.Draw(radarBounds, hudOpts, g.player.Pos(), g.player.Heading(), g.player.TurretAngle(), cameraViewDegrees)
}

func (g *Game) drawCrosshairs(hudOpts *render.DrawHudOptions) {
	if g.crosshairs == nil {
		return
	}

	cScale := g.crosshairs.Scale() * g.renderScale * g.hudScale
	cWidth, cHeight := cScale*float64(g.crosshairs.Width()), cScale*float64(g.crosshairs.Height())
	cX, cY := float64(g.width)/2-cWidth/2, float64(g.height)/2-cHeight/2

	crosshairBounds := image.Rect(
		int(cX), int(cY), int(cX+cWidth), int(cY+cHeight),
	)

	g.crosshairs.Draw(crosshairBounds, hudOpts)
}

func (g *Game) drawTargetReticle(hudOpts *render.DrawHudOptions) {
	if g.targetReticle == nil || g.player.Target() == nil {
		return
	}

	s := g.getSpriteFromEntity(g.player.Target())
	if s == nil {
		return
	}

	targetBounds := s.ScreenRect()
	if targetBounds == nil {
		return
	}

	g.targetReticle.Draw(*targetBounds, hudOpts)
}

func (g *Game) drawNavReticle(hudOpts *render.DrawHudOptions) {
	if g.navReticle == nil || g.player.Target() != nil || g.player.navPoint == nil {
		return
	}

	s := g.player.navPoint
	if s == nil {
		return
	}

	navBounds := s.ScreenRect()
	if navBounds == nil {
		return
	}

	g.navReticle.Draw(*navBounds, hudOpts)
}
