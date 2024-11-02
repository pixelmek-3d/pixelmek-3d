package game

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
)

type HUDElementType int

const (
	HUD_FPS HUDElementType = iota
	HUD_BANNER
	HUD_ALTIMETER
	HUD_ARMAMENT
	HUD_COMPASS
	HUD_CROSSHAIRS
	HUD_HEAT
	HUD_JETS
	HUD_NAV_RETICLE
	HUD_NAV_STATUS
	HUD_PLAYER_STATUS
	HUD_RADAR
	HUD_FRIENDLY_RETICLE
	HUD_TARGET_RETICLE
	HUD_TARGET_STATUS
	HUD_THROTTLE
	TOTAL_HUD_ELEMENT_TYPES
)

type HUDElement interface {
	Draw(bounds image.Rectangle, hudOpts *render.DrawHudOptions)
	Width() int
	Height() int
	Scale() float64
	SetScale(float64)
}

func (g *Game) GetHUDElement(t HUDElementType) HUDElement {
	if h, ok := g.playerHUD[t]; ok {
		return h
	}
	return nil
}

func (g *Game) initInteractiveTypes() {
	g.interactiveSpriteTypes = map[SpriteType]bool{
		MechSpriteType:        true,
		VehicleSpriteType:     true,
		VTOLSpriteType:        true,
		InfantrySpriteType:    true,
		EmplacementSpriteType: true,
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
	g.playerHUD = make(map[HUDElementType]HUDElement)

	compass := render.NewCompass(g.fonts.HUDFont)
	g.playerHUD[HUD_COMPASS] = compass

	altimeter := render.NewAltimeter(g.fonts.HUDFont)
	g.playerHUD[HUD_ALTIMETER] = altimeter

	heat := render.NewHeatIndicator(g.fonts.HUDFont)
	g.playerHUD[HUD_HEAT] = heat

	jets := render.NewJumpJetIndicator(g.fonts.HUDFont)
	g.playerHUD[HUD_JETS] = jets

	radar := render.NewRadar(g.fonts.HUDFont)
	radar.SetMapLines(g.collisionMap)
	g.playerHUD[HUD_RADAR] = radar

	armament := render.NewArmament(g.fonts.HUDFont)
	g.playerHUD[HUD_ARMAMENT] = armament

	throttle := render.NewThrottle(g.fonts.HUDFont)
	g.playerHUD[HUD_THROTTLE] = throttle

	playerStatus := render.NewUnitStatus(true, g.fonts.HUDFont)
	g.playerHUD[HUD_PLAYER_STATUS] = playerStatus
	targetStatus := render.NewUnitStatus(false, g.fonts.HUDFont)
	g.playerHUD[HUD_TARGET_STATUS] = targetStatus
	navStatus := render.NewNavStatus(g.fonts.HUDFont)
	g.playerHUD[HUD_NAV_STATUS] = navStatus

	crosshairsSheet := getSpriteFromFile("hud/crosshairs_sheet.png")
	crosshairs := render.NewCrosshairs(
		crosshairsSheet, resources.CrosshairsSheet.Columns, resources.CrosshairsSheet.Rows, g.hudCrosshairIndex,
	)
	g.playerHUD[HUD_CROSSHAIRS] = crosshairs

	tgtReticleSheet := getSpriteFromFile("hud/target_reticle.png")
	targetReticle := render.NewTargetReticle(tgtReticleSheet)
	g.playerHUD[HUD_TARGET_RETICLE] = targetReticle

	friendlyReticleSheet := getSpriteFromFile("hud/friendly_reticle.png")
	friendlyReticle := render.NewTargetReticle(friendlyReticleSheet)
	g.playerHUD[HUD_FRIENDLY_RETICLE] = friendlyReticle

	navReticleSheet := getSpriteFromFile("hud/nav_reticle.png")
	navReticle := render.NewNavReticle(navReticleSheet)
	g.playerHUD[HUD_NAV_RETICLE] = navReticle

	banner := render.NewMissionBanner(g.fonts.HUDFont)
	g.playerHUD[HUD_BANNER] = banner

	fps := render.NewFPSIndicator(g.fonts.HUDFont)
	g.playerHUD[HUD_FPS] = fps
}

func (g *Game) resetHUDElementScale() {
	hudElement := g.GetHUDElement(HUD_CROSSHAIRS)
	if hudElement != nil && hudElement.Scale() < 1.0 {
		for _, hudElement := range g.playerHUD {
			hudElement.SetScale(1.0)
		}
	}
}

// drawHUD draws HUD elements on the screen
func (g *Game) drawHUD(screen *ebiten.Image) {
	hudRect := g.uiRect()
	marginX, marginY := hudRect.Dx()/50, hudRect.Dy()/50

	hudOpts := &render.DrawHudOptions{
		Screen:         screen,
		HudRect:        hudRect,
		MarginX:        marginX,
		MarginY:        marginY,
		UseCustomColor: g.hudUseCustomColor,
		Color:          *g.hudRGBA,
	}

	// draw FPS display
	g.drawFPS(hudOpts)

	if !g.hudEnabled {
		return
	}

	// custom HUD elements based on player status
	switch {
	case g.player.ejectionPod != nil:
		// limited HUD for ejection pod: radar, altimeter, compass, mission banner
		g.drawCompass(hudOpts)
		g.drawAltimeter(hudOpts)
		g.drawRadar(hudOpts)
		g.drawMissionBanner(hudOpts)

		// make sure HUD element scale is properly set in case of just coming from power down
		g.resetHUDElementScale()
		return

	case g.player.Powered() == model.POWER_ON:
		// make sure HUD element scale is properly set in case of just coming from power down
		g.resetHUDElementScale()

	default:
		// handle player shutting down or powering up
		isOverHeated := g.player.OverHeated()
		switch unitType := g.player.Unit.(type) {
		case *model.Mech:
			m := g.player.Unit.(*model.Mech)
			switch {
			case m.PowerOffTimer > 0:
				powerTime := model.TICKS_PER_SECOND * model.UNIT_POWER_OFF_SECONDS
				remainTime := float64(m.PowerOffTimer)
				hudPercent := remainTime / powerTime

				for hudType, hudElement := range g.playerHUD {
					switch hudType {
					case HUD_HEAT:
						if m.Heat() > 0 {
							// keep only heat indicator shown while powering down from heat shutdown
							hudElement.SetScale(1.0)
						} else {
							hudElement.SetScale(hudPercent)
						}

					default:
						hudElement.SetScale(hudPercent)
					}
				}

			case m.PowerOnTimer > 0 && !isOverHeated:
				powerTime := model.TICKS_PER_SECOND * model.MECH_POWER_ON_SECONDS
				remainTime := float64(m.PowerOnTimer)
				hudPercent := 1 - (remainTime / powerTime)

				for hudType, hudElement := range g.playerHUD {
					switch hudType {
					case HUD_HEAT:
						if m.Heat() > 0 {
							// keep only heat indicator shown while powering up from heat shutdown
							hudElement.SetScale(1.0)
						} else {
							hudElement.SetScale(hudPercent)
						}

					case HUD_CROSSHAIRS:
						if m.PowerOnTimer > 0 {
							// hide crosshairs until fully powered on
							hudElement.SetScale(0)
						}

					default:
						hudElement.SetScale(hudPercent)
					}
				}

			default:
				if m.Heat() > 0 {
					// keep only heat indicator on while powered down if hot
					g.drawHeatIndicator(hudOpts)
				}
				return
			}

		default:
			panic(fmt.Sprintf("unhandled player HUD power off for unit type %s", unitType))
		}
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

	// draw mission banner
	g.drawMissionBanner(hudOpts)
}

func (g *Game) drawFPS(hudOpts *render.DrawHudOptions) {
	fps := g.GetHUDElement(HUD_FPS).(*render.FPSIndicator)
	if fps == nil || !g.fpsEnabled {
		return
	}

	fpsText := fmt.Sprintf("FPS: %0.1f | TPS: %0.1f/%d", ebiten.ActualFPS(), ebiten.ActualTPS(), ebiten.TPS())
	fps.SetFPSText(fpsText)

	marginY := hudOpts.MarginY
	hudRect := hudOpts.HudRect

	fScale := fps.Scale() * g.hudScale
	fWidth, fHeight := int(fScale*float64(hudRect.Dx())/5), int(fScale*float64(marginY))

	fX, fY := 0, 0
	fBounds := image.Rect(
		fX, fY, fX+fWidth, fY+fHeight,
	)
	fps.Draw(fBounds, hudOpts)
}

func (g *Game) drawPlayerStatus(hudOpts *render.DrawHudOptions) {
	playerStatus := g.GetHUDElement(HUD_PLAYER_STATUS).(*render.UnitStatus)
	if playerStatus == nil {
		return
	}

	hudRect := hudOpts.HudRect
	hudW, hudH := hudRect.Dx(), hudRect.Dy()

	statusScale := playerStatus.Scale() * g.hudScale
	if statusScale == 0 {
		return
	}
	statusWidth, statusHeight := int(statusScale*float64(hudW)/5), int(statusScale*float64(hudH)/5)

	sX, sY := hudRect.Min.X+int(4*float64(hudW)/5-2*float64(statusWidth)/3), hudRect.Min.Y+hudH-statusHeight
	sBounds := image.Rect(
		sX, sY, sX+statusWidth, sY+statusHeight,
	)
	playerStatus.SetUnit(g.player.sprite)
	playerStatus.Draw(sBounds, hudOpts)
}

func (g *Game) drawTargetStatus(hudOpts *render.DrawHudOptions) {
	targetStatus := g.GetHUDElement(HUD_TARGET_STATUS).(*render.UnitStatus)
	if targetStatus == nil || g.player.Target() == nil {
		return
	}

	hudRect := hudOpts.HudRect
	hudW, hudH := hudRect.Dx(), hudRect.Dy()

	statusScale := targetStatus.Scale() * g.hudScale
	if statusScale == 0 {
		return
	}
	statusWidth, statusHeight := int(statusScale*float64(hudW)/5), int(statusScale*float64(hudH)/5)

	sX, sY := hudRect.Min.X, hudRect.Min.Y+hudH-statusHeight
	sBounds := image.Rect(
		sX, sY, sX+statusWidth, sY+statusHeight,
	)

	targetEntity := g.player.Target()
	targetUnit := targetStatus.Unit()
	if targetUnit == nil || targetUnit.Entity != targetEntity {
		targetUnit = g.getSpriteFromEntity(targetEntity)
	}

	if targetUnit != nil {
		targetDistance := model.EntityDistance(g.player, targetUnit.Entity) - targetUnit.CollisionRadius() - g.player.CollisionRadius()
		distanceMeters := targetDistance * model.METERS_PER_UNIT
		targetStatus.SetUnitDistance(distanceMeters)
	}

	targetIsFriendly := g.IsFriendly(g.player, targetEntity)

	if targetUnit == nil || targetIsFriendly || g.player.Powered() != model.POWER_ON {
		// do not show target lock indicator if no target, target is friendly, or player not full powered on
		targetStatus.ShowTargetLock(false)
		targetStatus.SetTargetLock(0)
	} else {
		// determine if lock percent should show
		hasLockOns := false
		for _, w := range g.player.Armament() {
			missileWeapon, isMissile := w.(*model.MissileWeapon)
			if isMissile && missileWeapon.IsLockOn() {
				hasLockOns = true
				break
			}
		}
		targetStatus.ShowTargetLock(hasLockOns)
		targetStatus.SetTargetLock(g.player.TargetLock())
	}

	// show different target reticle if target is friendly
	var targetReticle *render.TargetReticle
	if targetUnit != nil && targetIsFriendly {
		targetReticle = g.GetHUDElement(HUD_FRIENDLY_RETICLE).(*render.TargetReticle)
	} else {
		targetReticle = g.GetHUDElement(HUD_TARGET_RETICLE).(*render.TargetReticle)
	}
	targetReticle.ReticleLeadBounds = nil

	targetStatus.SetTargetReticle(targetReticle)
	targetStatus.SetUnit(targetUnit)
	targetStatus.Draw(sBounds, hudOpts)
}

func (g *Game) drawNavStatus(hudOpts *render.DrawHudOptions) {
	navStatus := g.GetHUDElement(HUD_NAV_STATUS).(*render.NavStatus)
	navPoint := g.player.NavPoint()
	if navStatus == nil || navPoint == nil || g.player.Target() != nil {
		return
	}

	hudRect := hudOpts.HudRect
	hudW, hudH := hudRect.Dx(), hudRect.Dy()

	statusScale := navStatus.Scale() * g.hudScale
	if statusScale == 0 {
		return
	}
	statusWidth, statusHeight := int(statusScale*float64(hudW)/5), int(statusScale*float64(hudH)/5)

	sX, sY := hudRect.Min.X, hudRect.Min.Y+hudH-statusHeight
	sBounds := image.Rect(
		sX, sY, sX+statusWidth, sY+statusHeight,
	)

	pPos, nPos := g.player.Pos(), navPoint.Pos()
	navLine := geom.Line{
		X1: pPos.X, Y1: pPos.Y,
		X2: nPos.X, Y2: nPos.Y,
	}
	navDistance := navLine.Distance() * model.METERS_PER_UNIT

	navStatus.SetNavDistance(navDistance)
	navStatus.SetNavPoint(navPoint)
	navStatus.Draw(sBounds, hudOpts)
}

func (g *Game) drawArmament(hudOpts *render.DrawHudOptions) {
	armament := g.GetHUDElement(HUD_ARMAMENT).(*render.Armament)
	if armament == nil {
		return
	}

	marginX := hudOpts.MarginX
	hudRect := hudOpts.HudRect
	hudW, hudH := hudRect.Dx(), hudRect.Dy()

	armamentScale := armament.Scale() * g.hudScale
	if armamentScale == 0 {
		return
	}
	armamentWidth, armamentHeight := int(armamentScale*float64(hudW)/3), int(armamentScale*float64(3*hudH)/8)
	aX, aY := hudRect.Min.X+hudW-armamentWidth+marginX, hudRect.Min.Y
	aBounds := image.Rect(
		aX, aY, aX+armamentWidth, aY+armamentHeight,
	)

	weaponFireMode := g.player.fireMode
	weaponGroups := g.player.weaponGroups
	weaponOrGroupIndex := g.player.selectedWeapon
	if g.player.fireMode == model.GROUP_FIRE {
		weaponOrGroupIndex = g.player.selectedGroup
	}

	if g.player.debugCameraTarget != nil {
		// override display for debug camera target
		if !armament.IsDebugWeapons() {
			armament.SetDebugWeapons(g.player.debugCameraTarget.Armament())
		}
		weaponFireMode = model.GROUP_FIRE
		weaponGroups = make([][]model.Weapon, 0)
		weaponOrGroupIndex = 0
	} else if armament.IsDebugWeapons() {
		armament.SetWeapons(g.player.Armament())
	}

	armament.SetWeaponGroups(weaponGroups)
	armament.SetSelectedWeapon(weaponOrGroupIndex, weaponFireMode)
	armament.Draw(aBounds, hudOpts)
}

func (g *Game) drawCompass(hudOpts *render.DrawHudOptions) {
	compass := g.GetHUDElement(HUD_COMPASS).(*render.Compass)
	if compass == nil {
		return
	}

	hudRect := hudOpts.HudRect
	hudW, hudH := hudRect.Dx(), hudRect.Dy()

	compassScale := compass.Scale() * g.hudScale
	if compassScale == 0 {
		return
	}
	compassWidth, compassHeight := int(compassScale*float64(3*hudW)/10), int(compassScale*float64(hudH)/21)
	cX, cY := hudRect.Min.X+int(float64(hudW)/2-float64(compassWidth)/2), hudRect.Min.Y
	cBounds := image.Rect(
		cX, cY, cX+compassWidth, cY+compassHeight,
	)

	camPos := g.player.Pos()
	camHeading := g.player.Heading()
	camTurretAngle := g.player.TurretAngle()

	if g.player.debugCameraTarget != nil {
		// override display for debug camera target
		camPos = g.player.debugCameraTarget.Pos()
		camHeading = g.player.debugCameraTarget.Heading()
		camTurretAngle = g.player.debugCameraTarget.TurretAngle()
	}

	if g.player.Target() == nil {
		compass.SetTargetEnabled(false)
	} else {
		targetPos := g.player.Target().Pos()
		tLine := geom.Line{
			X1: camPos.X, Y1: camPos.Y,
			X2: targetPos.X, Y2: targetPos.Y,
		}
		tAngle := tLine.Angle()

		compass.SetTargetEnabled(true)
		compass.SetTargetHeading(tAngle)
		compass.SetTargetFriendly(g.IsFriendly(g.player, g.player.Target()))
	}

	if g.player.currentNav == nil {
		compass.SetNavEnabled(false)
	} else {
		navPos := g.player.currentNav.Pos()
		tLine := geom.Line{
			X1: camPos.X, Y1: camPos.Y,
			X2: navPos.X, Y2: navPos.Y,
		}
		nAngle := tLine.Angle()

		compass.SetNavEnabled(true)
		compass.SetNavHeading(nAngle)
	}

	compass.SetValues(camHeading, camTurretAngle)
	compass.Draw(cBounds, hudOpts)
}

func (g *Game) drawAltimeter(hudOpts *render.DrawHudOptions) {
	altimeter := g.GetHUDElement(HUD_ALTIMETER).(*render.Altimeter)
	if altimeter == nil {
		return
	}

	marginY := hudOpts.MarginY
	hudRect := hudOpts.HudRect
	hudW, hudH := hudRect.Dx(), hudRect.Dy()

	// convert Z position to meters of altitude
	altitude := g.player.PosZ() * model.METERS_PER_UNIT

	if g.player.debugCameraTarget != nil {
		// override display for debug camera target
		altitude = g.player.debugCameraTarget.PosZ() * model.METERS_PER_UNIT
	}

	altScale := altimeter.Scale() * g.hudScale
	if altScale == 0 {
		return
	}
	altWidth, altHeight := int(altScale*float64(hudW)/24), int(altScale*float64(3*hudH)/12)
	aX, aY := hudRect.Min.X, hudRect.Min.Y+int(float64(hudH)/2-float64(altHeight)/2-float64(marginY))
	aBounds := image.Rect(
		aX, aY, aX+altWidth, aY+altHeight,
	)
	altimeter.SetValues(altitude, g.player.Pitch())
	altimeter.Draw(aBounds, hudOpts)
}

func (g *Game) drawHeatIndicator(hudOpts *render.DrawHudOptions) {
	heat := g.GetHUDElement(HUD_HEAT).(*render.HeatIndicator)
	if heat == nil {
		return
	}

	hudRect := hudOpts.HudRect
	hudW, hudH := hudRect.Dx(), hudRect.Dy()

	// convert heat dissipation to seconds
	currHeat, maxHeat := g.player.Heat(), g.player.MaxHeat()
	dissipationPerSec := g.player.HeatDissipation() * model.TICKS_PER_SECOND

	if g.player.debugCameraTarget != nil {
		// override display for debug camera target
		currHeat, maxHeat = g.player.debugCameraTarget.Heat(), g.player.debugCameraTarget.MaxHeat()
		dissipationPerSec = g.player.debugCameraTarget.HeatDissipation() * model.TICKS_PER_SECOND
	}

	heatScale := heat.Scale() * g.hudScale
	if heatScale == 0 {
		return
	}
	heatWidth, heatHeight := int(heatScale*float64(3*hudW)/10), int(heatScale*float64(hudH)/18)
	hX, hY := hudRect.Min.X+int(float64(hudW)/2-float64(heatWidth)/2), hudRect.Min.Y+hudH-heatHeight
	hBounds := image.Rect(
		hX, hY, hX+heatWidth, hY+heatHeight,
	)
	heat.SetValues(currHeat, maxHeat, dissipationPerSec)
	heat.Draw(hBounds, hudOpts)
}

func (g *Game) drawThrottle(hudOpts *render.DrawHudOptions) {
	throttle := g.GetHUDElement(HUD_THROTTLE).(*render.Throttle)
	if throttle == nil {
		return
	}

	hudRect := hudOpts.HudRect
	hudW, hudH := hudRect.Dx(), hudRect.Dy()

	velocity := g.player.Velocity()
	if g.player.JumpJetVelocity() > 0 {
		velocity = g.player.JumpJetVelocity()
	}

	// convert velocity from units per tick to kilometers per hour
	kphVelocity := velocity * model.VELOCITY_TO_KPH
	kphVelocityZ := g.player.VelocityZ() * model.VELOCITY_TO_KPH
	kphTgtVelocity := g.player.TargetVelocity() * model.VELOCITY_TO_KPH
	kphMax := g.player.MaxVelocity() * model.VELOCITY_TO_KPH

	if g.player.debugCameraTarget != nil {
		// override display for debug camera target
		kphVelocity = g.player.debugCameraTarget.Velocity() * model.VELOCITY_TO_KPH
		kphVelocityZ = g.player.debugCameraTarget.VelocityZ() * model.VELOCITY_TO_KPH
		kphTgtVelocity = g.player.debugCameraTarget.TargetVelocity() * model.VELOCITY_TO_KPH
		kphMax = g.player.debugCameraTarget.MaxVelocity() * model.VELOCITY_TO_KPH
	}

	throttleScale := throttle.Scale() * g.hudScale
	if throttleScale == 0 {
		return
	}
	throttleWidth, throttleHeight := int(throttleScale*float64(hudW)/8), int(throttleScale*float64(3*hudH)/8)
	tX, tY := hudRect.Min.X+hudW-throttleWidth, hudRect.Min.Y+hudH-throttleHeight
	tBounds := image.Rect(
		tX, tY,
		tX+throttleWidth, tY+throttleHeight,
	)
	throttle.SetValues(kphVelocity, kphTgtVelocity, kphVelocityZ, kphMax, kphMax/2)
	throttle.Draw(tBounds, hudOpts)
}

func (g *Game) drawJumpJetIndicator(hudOpts *render.DrawHudOptions) {
	jets := g.GetHUDElement(HUD_JETS).(*render.JumpJetIndicator)
	if jets == nil {
		return
	}

	if g.player == nil || g.player.Unit.JumpJets() == 0 {
		return
	}

	marginX := hudOpts.MarginX
	hudRect := hudOpts.HudRect
	hudW, hudH := hudRect.Dx(), hudRect.Dy()

	jDuration := g.player.Unit.JumpJetDuration()
	jMaxDuration := g.player.Unit.MaxJumpJetDuration()

	if g.player.debugCameraTarget != nil {
		// override display for debug camera target
		jDuration = g.player.debugCameraTarget.JumpJetDuration()
		jMaxDuration = g.player.debugCameraTarget.MaxJumpJetDuration()
	}

	jetsScale := jets.Scale() * g.hudScale
	if jetsScale == 0 {
		return
	}
	jetsWidth, jetsHeight := int(jetsScale*float64(hudW)/12), int(jetsScale*float64(3*hudH)/18)
	hX, hY := hudRect.Min.X+int(float64(hudW)/5+2*float64(marginX)), hudRect.Min.Y+hudH-jetsHeight
	jBounds := image.Rect(
		hX, hY, hX+jetsWidth, hY+jetsHeight,
	)
	jets.SetValues(jDuration, jMaxDuration)
	jets.Draw(jBounds, hudOpts)
}

func (g *Game) drawRadar(hudOpts *render.DrawHudOptions) {
	radar := g.GetHUDElement(HUD_RADAR).(*render.Radar)
	if radar == nil {
		return
	}

	hudRect := hudOpts.HudRect
	hudW, hudH := hudRect.Dx(), hudRect.Dy()

	radarScale := radar.Scale() * g.hudScale
	if radarScale == 0 {
		return
	}
	radarWidth, radarHeight := int(radarScale*float64(hudW)/3), int(radarScale*float64(hudH)/3)
	rX, rY := hudRect.Min.X, hudRect.Min.Y
	radarBounds := image.Rect(
		rX, rY, rX+radarWidth, rY+radarHeight,
	)

	// find all units and nav points within range to draw as blips
	maxDistanceMeters := 1000.0 // TODO: set in Radar object and game config
	maxDistanceUnits := maxDistanceMeters / model.METERS_PER_UNIT

	radarBlips := make([]*render.RadarBlip, 0, 128)
	rNavPoints := make([]*render.RadarNavPoint, 0, len(g.mission.NavPoints))

	camPos := g.player.Pos()
	camHeading := g.player.Heading()
	camTurretAngle := g.player.TurretAngle()

	if g.player.debugCameraTarget != nil {
		// override radar location for debug camera target
		camPos = g.player.debugCameraTarget.Pos()
		camHeading = g.player.debugCameraTarget.Heading()
		camTurretAngle = g.player.debugCameraTarget.TurretAngle()

	}

	playerTarget := g.player.Target()
	playerNav := g.player.NavPoint()

	// discover nav points that are in range
	navCount := 0
	for _, nav := range g.mission.NavPoints {
		navPos := nav.Pos()
		navLine := geom.Line{
			X1: camPos.X, Y1: camPos.Y,
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
		relAngle := camHeading - navLine.Angle()
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
				X1: camPos.X, Y1: camPos.Y,
				X2: unitPos.X, Y2: unitPos.Y,
			}

			unitIsFriendly := g.IsFriendly(g.player, entity)
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
			relAngle := camHeading - unitLine.Angle()
			// determine heading of unit relative from player heading
			relHeading := camHeading - unit.Heading()
			relTurretHeading := camHeading - unit.TurretAngle()

			blip := &render.RadarBlip{
				Unit:          unit,
				Distance:      unitDistance,
				Angle:         relAngle,
				Heading:       relHeading,
				TurretHeading: relTurretHeading,
				IsTarget:      unitIsTarget,
				IsFriendly:    unitIsFriendly,
			}

			radarBlips = append(radarBlips, blip)
			blipCount++
			return true
		})
	}

	if g.debug && playerTarget != nil {
		// draw debug nav lines for AI pathing of player target
		var navLines []*geom.Line
		targetUnit := model.EntityUnit(playerTarget)
		if targetUnit != nil {
			unitBehavior := g.ai.UnitAI(targetUnit)
			if unitBehavior.piloting.Len() > 0 {
				navLines = make([]*geom.Line, 0, unitBehavior.piloting.Len())
				prevPathPos := targetUnit.Pos()
				for _, pathPos := range unitBehavior.piloting.destPath {
					navLines = append(navLines, &geom.Line{X1: prevPathPos.X, Y1: prevPathPos.Y, X2: pathPos.X, Y2: pathPos.Y})
					prevPathPos = pathPos
				}
			}
		}
		radar.SetNavLines(navLines)
	}

	cameraViewDegrees := g.fovDegrees / g.camera.FovDepth()
	radar.SetValues(camPos, camHeading, camTurretAngle, cameraViewDegrees)

	radar.SetNavPoints(rNavPoints[:navCount])
	radar.SetRadarBlips(radarBlips[:blipCount])

	radar.Draw(radarBounds, hudOpts)
}

func (g *Game) drawCrosshairs(hudOpts *render.DrawHudOptions) {
	crosshairs := g.GetHUDElement(HUD_CROSSHAIRS).(*render.Crosshairs)
	if crosshairs == nil {
		return
	}

	cScale := crosshairs.Scale() * g.hudScale
	if cScale == 0 {
		return
	}

	hudH := float64(hudOpts.HudRect.Dy())
	cWidth, cHeight := hudH/8, hudH/8
	cX, cY := float64(g.screenWidth)/2-cWidth/2, float64(g.screenHeight)/2-cHeight/2

	crosshairBounds := image.Rect(
		int(cX), int(cY), int(cX+cWidth), int(cY+cHeight),
	)

	deltaAngle := model.AngleDistance(g.player.TurretAngle(), g.player.cameraAngle)
	deltaPitch := model.AngleDistance(g.player.Pitch(), g.player.cameraPitch)

	if g.player.debugCameraTarget != nil {
		// override crosshair location for debug camera target
		deltaAngle = 0
		deltaPitch = 0
	}

	fovHorizontal, fovVertical := g.camera.FovRadians(), g.camera.FovRadiansVertical()
	crosshairs.SetOffsets(deltaAngle, deltaPitch)
	crosshairs.SetFocalAngles(fovHorizontal, fovVertical)

	crosshairs.Draw(crosshairBounds, hudOpts)
}

func (g *Game) drawTargetReticle(hudOpts *render.DrawHudOptions) {
	var targetReticle *render.TargetReticle
	if g.player.Target() != nil && g.IsFriendly(g.player, g.player.Target()) {
		targetReticle = g.GetHUDElement(HUD_FRIENDLY_RETICLE).(*render.TargetReticle)
	} else {
		targetReticle = g.GetHUDElement(HUD_TARGET_RETICLE).(*render.TargetReticle)
	}
	if targetReticle == nil || g.player.Target() == nil {
		return
	}

	s := g.getSpriteFromEntity(g.player.Target())
	if s == nil {
		return
	}

	targetBounds := s.ScreenRect(g.renderScale)
	if targetBounds == nil {
		return
	}

	var targetLeadBounds *image.Rectangle
	if g.player.reticleLead != nil {
		targetLeadBounds = g.player.reticleLead.ScreenRect(g.renderScale)
	}
	targetReticle.ReticleLeadBounds = targetLeadBounds
	targetReticle.Friendly = g.IsFriendly(g.player, g.player.Target())

	targetReticle.Draw(*targetBounds, hudOpts)
}

func (g *Game) drawNavReticle(hudOpts *render.DrawHudOptions) {
	navReticle := g.GetHUDElement(HUD_NAV_RETICLE).(*render.NavReticle)
	if navReticle == nil || g.player.Target() != nil || g.player.currentNav == nil {
		return
	}

	s := g.player.currentNav
	if s == nil {
		return
	}

	navBounds := s.ScreenRect(g.renderScale)
	if navBounds == nil {
		return
	}

	navReticle.Draw(*navBounds, hudOpts)
}

func (g *Game) drawMissionBanner(hudOpts *render.DrawHudOptions) {
	banner := g.GetHUDElement(HUD_BANNER).(*render.MissionBanner)
	if banner == nil {
		return
	}

	bannerText := ""
	if !g.InProgress() {
		if g.objectives.Status() == OBJECTIVES_COMPLETED {
			bannerText = "Mission Successful..."
		} else {
			bannerText = "Mission Failed..."
		}
	}
	if len(bannerText) == 0 {
		return
	}

	banner.SetBannerText(bannerText)

	marginY := hudOpts.MarginY
	hudRect := hudOpts.HudRect

	bScale := banner.Scale() * g.hudScale
	bWidth, bHeight := int(bScale*float64(hudRect.Dx())), 3*int(bScale*float64(marginY))

	bX, bY := hudRect.Min.X, 0
	bBounds := image.Rect(
		bX, bY, bX+bWidth, bY+bHeight,
	)
	banner.Draw(bBounds, hudOpts)
}
