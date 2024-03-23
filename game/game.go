package game

import (
	"fmt"
	"image"
	"math"
	"os"
	"sort"
	"time"

	"image/color"
	_ "image/png"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"

	input "github.com/quasilyte/ebitengine-input"
	log "github.com/sirupsen/logrus"
)

const (
	title = "PixelMek 3D"
	//--RaycastEngine constants
	//--set constant, texture size to be the wall (and sprite) texture size--//
	texWidth = 256

	// distance to keep away from walls and obstacles to avoid clipping
	// TODO: may want a smaller distance to test vs. sprites
	clipDistance = 0.1
)

// Game - This is the main type for your game.
type Game struct {
	scene  Scene
	menu   Menu
	paused bool

	resources   *model.ModelResources
	audio       *AudioHandler
	input       *input.Handler
	inputSystem input.System

	//--create slicer and declare slices--//
	tex                *TextureHandler
	initRenderFloorTex bool

	// window resolution and scaling
	screenWidth  int
	screenHeight int
	renderScale  float64
	fullscreen   bool
	vsync        bool
	opengl       bool
	fovDegrees   float64
	fovDepth     float64

	//--viewport width / height--//
	width  int
	height int

	player    *Player
	playerHUD map[HUDElementType]HUDElement
	fonts     *render.FontHandler

	hudEnabled        bool
	hudFont           string
	hudScale          float64
	hudRGBA           *color.NRGBA
	hudUseCustomColor bool

	//--define camera and rendering screens--//
	camera        *raycaster.Camera
	rayScreen     *ebiten.Image
	renderScreen  *ebiten.Image
	overlayScreen *ebiten.Image

	crtShader bool

	mouseMode      MouseMode
	mouseX, mouseY int

	// zoom settings
	zoomFovDepth float64

	renderDistance  float64
	clutterDistance float64

	// lighting settings
	lightFalloff       float64
	globalIllumination float64
	minLightRGB        *color.NRGBA
	maxLightRGB        *color.NRGBA

	lightAmpEngaged bool

	// Mission and map
	mapWidth, mapHeight int
	mission             *model.Mission
	collisionMap        []*geom.Line

	sprites                *SpriteHandler
	clutter                *ClutterHandler
	collisonSpriteTypes    map[SpriteType]struct{}
	interactiveSpriteTypes map[SpriteType]struct{}
	delayedProjectiles     map[*DelayedProjectileSpawn]struct{}

	// Gameplay
	objectives *ObjectivesHandler

	// control options
	throttleDecay bool

	osType     osType
	debug      bool
	fpsEnabled bool
}

type osType int

const (
	osTypeDesktop osType = iota
	osTypeBrowser
)

type TargetCycleType int

const (
	TARGET_NEXT TargetCycleType = iota
	TARGET_PREVIOUS
	TARGET_NEAREST
)

// NewGame - Allows the game to perform any initialization it needs to before starting to run.
// This is where it can query for any required services and load any non-graphic
// related content.  Calling base.Initialize will enumerate through any components
// and initialize them as well.
func NewGame() *Game {
	// initialize Game object
	g := new(Game)
	g.initConfig()
	g.initControls()

	if g.opengl {
		os.Setenv("EBITENGINE_GRAPHICS_LIBRARY", "opengl")
	}

	// initialize resources file handler
	resources.InitFS()

	// initialize fonts
	var err error
	g.fonts, err = render.NewFontHandler(g.hudFont)
	if err != nil {
		log.Error("Error loading font handler:", err)
		exit(1)
	}

	// initialize audio and background music
	g.audio = NewAudioHandler()
	g.audio.StartMenuMusic()

	g.initInteractiveTypes()
	g.initCollisionTypes()
	g.initCombatVariables()

	ebiten.SetWindowTitle(title)
	ebiten.SetTPS(int(model.TICKS_PER_SECOND))

	// use scale to keep the desired window width and height
	g.setResolution(g.screenWidth, g.screenHeight)
	g.setRenderScale(g.renderScale)
	g.setFullscreen(g.fullscreen)
	g.setVsyncEnabled(g.vsync)

	g.resources, err = model.LoadModelResources()
	if err != nil {
		log.Error("Error loading models:", err)
		exit(1)
	}

	// init texture and sprite handlers
	g.tex = NewTextureHandler(nil)
	g.tex.renderFloorTex = g.initRenderFloorTex
	g.sprites = NewSpriteHandler()

	// setup initial scene
	g.scene = NewIntroScene(g)

	// set window icon
	_, icon, err := resources.NewImageFromFile("icons/pixelmek_icon.png")
	if err != nil {
		log.Error(err)
	}
	if icon != nil {
		ebiten.SetWindowIcon([]image.Image{icon})
	}

	return g
}

func (g *Game) initMission() {
	if g.mission == nil {
		panic("g.mission must be set before initMission!")
	}

	// reload texture handler
	if g.tex != nil {
		g.initRenderFloorTex = g.tex.renderFloorTex
	}
	g.tex = NewTextureHandler(g.mission.Map())
	g.tex.renderFloorTex = g.initRenderFloorTex

	// clear mission sprites
	g.sprites.clear()

	g.collisionMap = g.mission.Map().GetCollisionLines(clipDistance)
	worldMap := g.mission.Map().Level(0)
	g.mapWidth = len(worldMap)
	g.mapHeight = len(worldMap[0])

	// load map and mission content
	g.loadContent()

	// initialize objectives
	g.objectives = NewObjectivesHandler(g, g.mission.Objectives)

	// init player at DZ
	pX, pY, pDegrees := g.mission.DropZone.Position[0], g.mission.DropZone.Position[1], g.mission.DropZone.Heading
	g.player.SetPos(&geom.Vector2{X: pX, Y: pY})
	g.player.SetHeading(geom.Radians(pDegrees))

	// init player as powered off but booting up
	g.player.SetPowered(model.POWER_ON)

	// init player armament for display
	if armament := g.GetHUDElement(HUD_ARMAMENT); armament != nil {
		armament.(*render.Armament).SetWeapons(g.player.Armament())
	}

	// initial mouse position to establish delta
	g.mouseX, g.mouseY = math.MinInt32, math.MinInt32

	//--init camera and renderer--//
	g.camera = raycaster.NewCamera(g.width, g.height, texWidth, g.mission.Map(), g.tex)
	g.camera.SetRenderDistance(g.renderDistance)
	g.camera.SetAlwaysSetSpriteScreenRect(true)

	if len(g.mission.Map().FloorBox.Image) > 0 {
		g.camera.SetFloorTexture(getTextureFromFile(g.mission.Map().FloorBox.Image))
	}
	if len(g.mission.Map().SkyBox.Image) > 0 {
		g.camera.SetSkyTexture(getTextureFromFile(g.mission.Map().SkyBox.Image))
	}

	// init camera lighting from map settings
	g.lightFalloff = g.mission.Map().Lighting.Falloff
	g.globalIllumination = g.mission.Map().Lighting.Illumination
	g.minLightRGB, g.maxLightRGB = g.mission.Map().Lighting.LightRGB()

	g.camera.SetLightFalloff(g.lightFalloff)
	g.camera.SetGlobalIllumination(g.globalIllumination)
	g.camera.SetLightRGB(*g.minLightRGB, *g.maxLightRGB)

	// initialize camera to player position
	g.updatePlayerCamera(true)
	g.setFovAngle(g.fovDegrees)
	g.fovDepth = g.camera.FovDepth()

	g.zoomFovDepth = 2.0

	// initialize clutter
	if g.clutter != nil {
		g.clutter.Update(g, true)
	}
}

// Run is the Ebiten Run loop caller
func (g *Game) Run() {
	g.paused = false

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	w, h := g.screenWidth, g.screenHeight
	return w, h
}

// Update - Allows the game to run logic such as updating the world, gathering input, and playing audio.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	g.inputSystem.Update()
	return g.scene.Update()
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	g.scene.Draw(screen)
}

// Gets the inner screen rect for UI space to account for ultra-wide resolutions
func (g *Game) uiRect() image.Rectangle {
	minUiAspectRatio, maxUiAspectRatio := 1.0, 1.5
	screenW, screenH := float64(g.screenWidth), float64(g.screenHeight)
	screenAspectRatio := screenW / screenH

	var paddingX, paddingY, uiWidth, uiHeight int

	if screenAspectRatio > maxUiAspectRatio {
		// ultra-wide aspect, constrict HUD width based on screen height
		paddingY = int(screenH * 0.02)
		uiHeight = int(screenH) - paddingY*2

		uiWidth = int(screenH * maxUiAspectRatio)
		//paddingX = menuWidth * 0.02
	} else if screenAspectRatio < minUiAspectRatio {
		// tall vertical aspect, constrict HUD height based on screen width
		paddingX = int(screenW * 0.02)
		uiWidth = int(screenW) - paddingX*2

		uiHeight = int(screenW / minUiAspectRatio)
		//paddingY = menuHeight * 0.02
	} else {
		// use current aspect ratio
		paddingX, paddingY = int(screenW*0.02), int(screenH*0.02)
		uiWidth, uiHeight = int(screenW)-paddingX*2, int(screenH)-paddingY*2
	}

	uiX, uiY := (g.screenWidth-uiWidth)/2, (g.screenHeight-uiHeight)/2
	return image.Rect(
		uiX, uiY,
		uiX+uiWidth, uiY+uiHeight,
	)
}

// Move player by move speed in the forward/backward direction
func (g *Game) Move(mSpeed float64) {
	playerPosition := g.player.Pos()
	moveLine := geom.LineFromAngle(playerPosition.X, playerPosition.Y, g.player.Heading(), mSpeed)
	g.updatePlayerPosition(moveLine.X2, moveLine.Y2, g.player.PosZ())
}

// Move player by strafe speed in the left/right direction
func (g *Game) Strafe(sSpeed float64) {
	strafeAngle := geom.HalfPi
	if sSpeed < 0 {
		strafeAngle = -strafeAngle
	}
	playerPosition := g.player.Pos()
	strafeLine := geom.LineFromAngle(playerPosition.X, playerPosition.Y, g.player.Heading()-strafeAngle, math.Abs(sSpeed))
	g.updatePlayerPosition(strafeLine.X2, strafeLine.Y2, g.player.PosZ())
}

// Move player by vertical speed in the up/down direction
func (g *Game) VerticalMove(vSpeed float64) {
	pos := g.player.Pos()
	newPosZ := g.player.PosZ() + vSpeed
	g.updatePlayerPosition(pos.X, pos.Y, newPosZ)
}

// Rotate player heading angle by rotation speed
func (g *Game) Rotate(rSpeed float64) {
	angle := model.ClampAngle(g.player.Heading() + rSpeed)
	g.player.SetHeading(angle)
	g.player.moved = true
}

// Rotate player turret angle, relative to body heading, by rotation speed
func (g *Game) RotateTurret(rSpeed float64) {
	if !g.player.HasTurret() {
		return
	}
	if g.player.Powered() != model.POWER_ON {
		// disallow turret rotation when shutdown
		return
	}

	angle := g.player.TurretAngle() + rSpeed

	// currently restricting turret rotation to only 90 degrees,
	if angle > geom.HalfPi {
		angle = geom.HalfPi
	} else if angle < -geom.HalfPi {
		angle = -geom.HalfPi
	}

	g.player.SetTurretAngle(angle)
	g.player.moved = true
}

// Update player pitch angle by pitch speed
func (g *Game) Pitch(pSpeed float64) {
	if g.player.Powered() != model.POWER_ON && g.player.ejectionPod == nil {
		// disallow turret pitch when shutdown
		return
	}
	// current raycasting method can only allow up to 45 degree pitch in either direction
	g.player.SetPitch(geom.Clamp(pSpeed+g.player.Pitch(), -geom.Pi/8, geom.Pi/4))
	g.player.moved = true
}

func (g *Game) InProgress() bool {
	return g.objectives.Status() == OBJECTIVES_IN_PROGRESS
}

func (g *Game) updateObjectives() {
	if g.InProgress() {
		g.objectives.Update(g)

		switch g.objectives.Status() {
		case OBJECTIVES_FAILED:
			// end mission as failure
			log.Debugf("one or more objectives failed")
		case OBJECTIVES_COMPLETED:
			// end mission as success
			log.Debugf("all objectives completed")
		}
	}
}

func (g *Game) updatePlayer() {
	if g.player.IsDestroyed() {
		justEjected := g.player.Eject(g)
		if justEjected {
			g.spawnPlayerDestroyEffects()
			g.player.sprite.SetDestroyCounter(int(model.TICKS_PER_SECOND / 3))
		} else {
			// keep playing destruction effects until the counter runs out
			if g.player.sprite.DestroyCounter() > 0 {
				fxDuration := g.spawnPlayerDestroyEffects()
				if fxDuration > 0 {
					counter := g.player.sprite.DestroyCounter() - 1
					g.player.sprite.SetDestroyCounter(counter)
				}
			}

			// make ejection pod thrust sound
			jetThrust := g.audio.sfx.mainSources[AUDIO_JUMP_JET]
			if !jetThrust.player.IsPlaying() {
				jetThrust.Play()
			}
		}

		g.player.moved = true
		return
	}

	if g.player.Update() {

		// TODO: refactor to use same update position function as sprites (g.updateMechPosition, etc.)

		position, posZ := g.player.Pos(), g.player.PosZ()
		velocity, velocityZ := g.player.Velocity(), g.player.VelocityZ()
		if velocityZ != 0 {
			posZ += velocityZ
		}

		moveHeading := g.player.Heading()
		if g.player.JumpJetsActive() || (posZ > 0 && g.player.JumpJets() > 0) {
			// while jumping, or still in air after jumping, continue from last jump jet active heading and velocity
			moveHeading = g.player.JumpJetHeading()
			velocity = g.player.JumpJetVelocity()
		}
		moveLine := geom.LineFromAngle(position.X, position.Y, moveHeading, velocity)

		newX, newY := moveLine.X2, moveLine.Y2
		g.updatePlayerPosition(newX, newY, posZ)
		g.player.moved = true

		// check for nav point visits
		for _, nav := range g.mission.NavPoints {
			if nav.Visited() {
				continue
			}

			navX, navY := nav.Position[0], nav.Position[1]
			if model.PointInProximity(1.0, newX, newY, navX, navY) {
				nav.SetVisited(true)

				// automatically cycle to next nav point
				if g.player.NavPoint() == nav && nav.Objective() != model.NavDustoffObjective {
					g.navPointCycle(false)
				}
			}
		}
	}

	if g.player.Powered() == model.POWER_ON {
		// make sure engine ambience is playing
		if g.audio.EngineAmbience() != _SFX_HINT_ENGINE {
			g.audio.StartEngineAmbience()
		}
	} else {
		// play power down sequence and make sure engine ambience is not playing
		engAmbience := g.audio.EngineAmbience()
		if engAmbience == _SFX_HINT_ENGINE {
			g.audio.StopEngineAmbience()
			g.audio.PlayPowerOffSequence()
		} else {
			// check if power on sound needs to be started
			switch g.player.Unit.(type) {
			case *model.Mech:
				m := g.player.Unit.(*model.Mech)
				if m.PowerOffTimer <= 0 && m.PowerOnTimer > 0 && engAmbience != _SFX_HINT_POWER_ON {
					// play power on sequence if not already playing
					g.audio.PlayPowerOnSequence()
				}
			}
		}
	}

	if g.player.JumpJets() > 0 {
		if g.player.JumpJetsActive() {
			// make jet thrust sound
			jetThrust := g.audio.sfx.mainSources[AUDIO_JUMP_JET]
			if !jetThrust.player.IsPlaying() {
				jetThrust.Play()
			}
		} else {
			jetThrust := g.audio.sfx.mainSources[AUDIO_JUMP_JET]
			if jetThrust.player.IsPlaying() {
				jetThrust.Pause()
			}
		}
	}

	if g.player.strideStomp && !g.player.JumpJetsActive() {
		// make stompy sound
		switch g.player.strideStompDir {
		case StrideStompLeft:
			stompy := g.audio.sfx.mainSources[AUDIO_STOMP_LEFT]
			stompy.Play()
		case StrideStompRight:
			stompy := g.audio.sfx.mainSources[AUDIO_STOMP_RIGHT]
			stompy.Play()
		case StrideStompBoth:
			lStompy := g.audio.sfx.mainSources[AUDIO_STOMP_LEFT]
			rStompy := g.audio.sfx.mainSources[AUDIO_STOMP_RIGHT]
			lStompy.Play()
			rStompy.Play()
		}

		// clear stomp flag
		g.player.strideStomp = false
	}

	target := g.player.Target()
	if target != nil && target.IsDestroyed() {
		g.player.SetTarget(nil)

	} else if target != nil {
		// only increment lock percent on target if reticle near target area and in weapon range
		s := g.getSpriteFromEntity(target)
		if s != nil {
			acquireLock := false
			crosshairLockSize := int(math.Ceil(float64(g.screenWidth) * 0.05))
			midW, midH := g.screenWidth/2, g.screenHeight/2
			crosshairBounds := image.Rect(
				midW-crosshairLockSize/2, midH-crosshairLockSize/2,
				midW+crosshairLockSize/2, midH+crosshairLockSize/2,
			)
			targetBounds := s.ScreenRect(g.renderScale)
			if targetBounds != nil {
				acquireLock = targetBounds.Overlaps(crosshairBounds)
			}

			targetDistance := model.EntityDistance(g.player, target) - g.player.CollisionRadius() - target.CollisionRadius()
			lockOnRange := 1000.0 / model.METERS_PER_UNIT

			if int(targetDistance) <= int(lockOnRange) {
				// TODO: decrease lock percent delta if further from target
				lockDelta := 0.25 / model.TICKS_PER_SECOND
				if !acquireLock {
					lockDelta = -0.15 / model.TICKS_PER_SECOND
				}

				targetLock := g.player.TargetLock() + lockDelta
				if targetLock > 1.0 {
					targetLock = 1.0
				} else if targetLock < 0 {
					targetLock = 0
				}
				g.player.SetTargetLock(targetLock)
			}
		}
	}
}

func (g *Game) updatePlayerPosition(setX, setY, setZ float64) {
	// Update player position
	newPos, newZ, isCollision, collisions := g.getValidMove(g.player.Unit, setX, setY, setZ, true)
	if !(newPos.Equals(g.player.Pos()) && newZ == g.player.PosZ()) {
		g.player.SetPos(newPos)
		g.player.SetPosZ(newZ)
		g.player.moved = true
	}

	if isCollision && len(collisions) > 0 {
		// apply damage to the first sprite entity that was hit
		collisionEntity := collisions[0]

		collisionDamage := 0.1 // TODO: determine collision damage based on player mech and speed
		collisionEntity.entity.ApplyDamage(collisionDamage)
		if g.debug {
			hp, maxHP := collisionEntity.entity.ArmorPoints()+collisionEntity.entity.StructurePoints(), collisionEntity.entity.MaxArmorPoints()+collisionEntity.entity.MaxStructurePoints()
			log.Debugf("collided for %0.1f (HP: %0.1f/%0.0f)", collisionDamage, hp, maxHP)
		}
	}
}

func (g *Game) navPointCycle(replaceTarget bool) {
	if len(g.mission.NavPoints) == 0 {
		return
	}

	if replaceTarget && g.player.Target() != nil {
		// unset player target so status display can show nav selection
		g.player.SetTarget(nil)
		if g.player.currentNav != nil {
			// on the first time after unset target have it not cycle to next nav
			return
		}
	}

	var newNav *model.NavPoint
	navPoints := g.mission.NavPoints
	currentNav := g.player.currentNav

	for _, n := range navPoints {
		if currentNav == nil {
			newNav = n
			break
		}

		if currentNav.NavPoint == n {
			// allow next loop iteration to select as new nav point
			currentNav = nil
			continue
		}
	}

	if newNav == nil {
		newNav = navPoints[0]
	}

	g.player.currentNav = render.NewNavSprite(newNav, 1.0)
}

func (g *Game) targetCrosshairs() model.Entity {
	newTarget := g.player.convergenceSprite
	if newTarget != nil && !newTarget.IsDestroyed() {
		g.player.SetTarget(newTarget.Entity)
		return newTarget.Entity
	}
	return nil
}

func (g *Game) targetCycle(cycleType TargetCycleType) model.Entity {
	targetables := make([]*render.Sprite, 0, 64)

	for spriteType := range g.sprites.sprites {
		g.sprites.sprites[spriteType].Range(func(k, _ interface{}) bool {
			if !g.isInteractiveType(spriteType) {
				// only cycle on certain sprite types (skip projectiles, effects, etc.)
				return true
			}

			s := getSpriteFromInterface(k.(raycaster.Sprite))
			if s.IsDestroyed() {
				return true
			}

			targetables = append(targetables, s)

			return true
		})
	}

	if len(targetables) == 0 {
		g.player.SetTarget(nil)
		return nil
	}

	// sort by distance to player
	playerPos := g.player.Pos()

	if cycleType == TARGET_PREVIOUS {
		sort.Slice(targetables, func(a, b int) bool {
			sA, sB := targetables[a], targetables[b]
			dA := geom.Distance2(sA.Pos().X, sA.Pos().Y, playerPos.X, playerPos.Y)
			dB := geom.Distance2(sB.Pos().X, sB.Pos().Y, playerPos.X, playerPos.Y)
			return dA > dB
		})
	} else {
		sort.Slice(targetables, func(a, b int) bool {
			sA, sB := targetables[a], targetables[b]
			dA := geom.Distance2(sA.Pos().X, sA.Pos().Y, playerPos.X, playerPos.Y)
			dB := geom.Distance2(sB.Pos().X, sB.Pos().Y, playerPos.X, playerPos.Y)
			return dA < dB
		})
	}

	var newTarget *render.Sprite

	if cycleType != TARGET_NEAREST {
		currentTarget := g.player.Target()
		for _, t := range targetables {
			if currentTarget == nil {
				newTarget = t
				break
			}

			if currentTarget == t.Entity {
				// allow next loop iteration to select as new target
				currentTarget = nil
				continue
			}
		}
	}

	if newTarget == nil {
		newTarget = targetables[0]
	}

	g.player.SetTarget(newTarget.Entity)
	return newTarget.Entity
}

func (g *Game) updateSprites() {
	// Update for animated sprite movement
	for spriteType := range g.sprites.sprites {
		g.sprites.sprites[spriteType].Range(func(k, _ interface{}) bool {

			switch spriteType {
			case MapSpriteType:
				s := k.(*render.Sprite)
				if s.IsDestroyed() {
					destroyCounter := s.DestroyCounter()
					if destroyCounter == 0 {
						// start the destruction process but do not remove yet
						// TODO: when tree is destroyed by projectile, add fire effect (energy and missile only)
						fxDuration := g.spawnGenericDestroyEffects(s, false)
						s.SetDestroyCounter(geom.ClampInt(fxDuration, 1, fxDuration))
					} else if destroyCounter == 1 {
						// delete when the counter is basically done (to differentiate with default int value 0)
						g.sprites.deleteMapSprite(s)
					} else {
						s.Update(g.player.Pos())
						s.SetDestroyCounter(destroyCounter - 1)
					}
					break
				}

				g.updateSpritePosition(s)
				s.Update(g.player.Pos())

			case MechSpriteType:
				s := k.(*render.MechSprite)
				sUnit := model.EntityUnit(s.Entity)
				if s.IsDestroyed() {
					if s.MechAnimation() != render.MECH_ANIMATE_DESTRUCT {
						// play unit destruction animation
						s.SetMechAnimation(render.MECH_ANIMATE_DESTRUCT, false)

						// spawn ejection pod
						g.spawnEjectionPod(s.Sprite)

					} else if s.LoopCounter() >= 1 {
						// delete when animation is over
						g.sprites.deleteMechSprite(s)
					} else {
						s.Update(g.player.Pos())
					}

					if sUnit.JumpJets() > 0 {
						g.removeJumpJetEffect(s.Sprite)
					}

					g.spawnMechDestroyEffects(s)
					break
				}

				mech := s.Mech()
				g.updateMechPosition(s)
				s.Update(g.player.Pos())
				g.updateWeaponCooldowns(sUnit)

				if sUnit.Powered() != model.POWER_ON {
					poweringOn := s.AnimationReversed()
					if mech.PowerOffTimer > 0 &&
						(s.MechAnimation() != render.MECH_ANIMATE_SHUTDOWN || poweringOn) {

						// start shutdown animation since unit is powering off
						s.SetMechAnimation(render.MECH_ANIMATE_SHUTDOWN, false)
					}
					if mech.PowerOffTimer <= 0 && mech.PowerOnTimer > 0 &&
						(s.MechAnimation() != render.MECH_ANIMATE_SHUTDOWN || !poweringOn) {

						// reverse shutdown animation since unit is powering on
						s.SetMechAnimation(render.MECH_ANIMATE_SHUTDOWN, true)

					}
					if s.MechAnimation() != render.MECH_ANIMATE_SHUTDOWN {
						s.SetMechAnimation(render.MECH_ANIMATE_SHUTDOWN, true)
					}
				} else {
					if mech.JumpJetsActive() {
						falling := s.AnimationReversed()
						if s.MechAnimation() != render.MECH_ANIMATE_JUMP_JET || falling {
							s.SetMechAnimation(render.MECH_ANIMATE_JUMP_JET, false)

							// spawn jump jet effect when first starting jump jet
							g.spawnJumpJetEffect(s.Sprite)
						}
					} else if s.VelocityZ() < 0 {
						falling := s.AnimationReversed()
						if s.MechAnimation() != render.MECH_ANIMATE_JUMP_JET || !falling {
							// reverse jump jet animation for falling
							s.SetMechAnimation(render.MECH_ANIMATE_JUMP_JET, true)

							// remove jump jet effect since jump jet no longer active
							g.removeJumpJetEffect(s.Sprite)
						}
					} else if s.Velocity() == 0 && s.VelocityZ() == 0 {
						if s.MechAnimation() != render.MECH_ANIMATE_IDLE {
							s.SetMechAnimation(render.MECH_ANIMATE_IDLE, false)
						}
					} else {
						if s.MechAnimation() != render.MECH_ANIMATE_STRUT {
							s.SetMechAnimation(render.MECH_ANIMATE_STRUT, false)
						}
					}
				}

				if s.StrideStomp() {
					s.ResetStrideStomp()
					pos, posZ := s.Pos(), s.PosZ()
					mechStompFile, err := StompSFXForMech(mech)
					if err == nil {
						g.audio.PlayExternalAudio(g, mechStompFile, pos.X, pos.Y, posZ, 2.5, 0.35)
					}
				}

				if mech.JumpJets() > 0 {
					mechJumpFile, err := JumpJetSFXForMech(mech)
					if err == nil {
						switch {
						case mech.JumpJetsActive() && !s.JetsPlaying:
							s.JetsPlaying = true
							g.audio.PlayEntityAudioLoop(g, mechJumpFile, mech, 5.0, 0.35)
						case !mech.JumpJetsActive() && s.JetsPlaying:
							g.audio.StopEntityAudioLoop(g, mechJumpFile, mech)
							s.JetsPlaying = false
						}
					}
				}

			case VehicleSpriteType:
				s := k.(*render.VehicleSprite)
				if s.IsDestroyed() {
					destroyCounter := s.DestroyCounter()
					if destroyCounter == 0 {
						// start the destruction process but do not remove yet
						fxDuration := g.spawnVehicleDestroyEffects(s)
						s.SetDestroyCounter(fxDuration)
					} else if destroyCounter == 1 {
						// delete when the counter is basically done (to differentiate with default int value 0)
						g.sprites.deleteVehicleSprite(s)
					} else {
						s.Update(g.player.Pos())
						s.SetDestroyCounter(destroyCounter - 1)
					}
					break
				}

				g.updateVehiclePosition(s)
				s.Update(g.player.Pos())
				g.updateWeaponCooldowns(model.EntityUnit(s.Entity))

			case VTOLSpriteType:
				s := k.(*render.VTOLSprite)
				if s.IsDestroyed() {
					// unique VTOL destroy effect where it crashes towards the ground spinning
					destroyCounter := s.DestroyCounter()
					if destroyCounter == 0 {
						// start the destruction process but do not remove yet
						g.spawnVTOLDestroyEffects(s, true)
						s.SetVelocity(0)
						s.SetVelocityZ(0)

						// use the destroy counter to determine which effects to spawn
						s.SetDestroyCounter(1)
					} else if s.PosZ() <= 0 {
						// instantly delete if it gets below the ground
						g.sprites.deleteVTOLSprite(s)
						break
					} else {
						// spawn only smoke effects
						g.spawnVTOLDestroyEffects(s, false)
					}

					// fall towards the ground
					velocityZ := s.VelocityZ()
					s.SetVelocityZ(velocityZ - model.GRAVITY_UNITS_PTT)

					// put in a tailspin
					heading := s.Heading()
					s.SetHeading(model.ClampAngle(heading + (geom.Pi2 / model.TICKS_PER_SECOND)))

					hasCollision := g.updateSpritePosition(s.Sprite)
					if hasCollision {
						// instantly remove on collision with some more explosions
						g.spawnVTOLDestroyEffects(s, true)
						g.sprites.deleteVTOLSprite(s)
						break
					}

					s.Update(g.player.Pos())
					break
				}

				g.updateVTOLPosition(s)
				s.Update(g.player.Pos())
				g.updateWeaponCooldowns(model.EntityUnit(s.Entity))

			case InfantrySpriteType:
				s := k.(*render.InfantrySprite)
				if s.IsDestroyed() {
					// infantry are destroyed immediately
					// TODO: if an infantry unit has death animation prior to deletion
					g.spawnInfantryDestroyEffects(s)
					g.sprites.deleteInfantrySprite(s)
					break
				}

				g.updateInfantryPosition(s)
				s.Update(g.player.Pos())
				g.updateWeaponCooldowns(model.EntityUnit(s.Entity))

			case EmplacementSpriteType:
				s := k.(*render.EmplacementSprite)
				if s.IsDestroyed() {
					destroyCounter := s.DestroyCounter()
					if destroyCounter == 0 {
						// start the destruction process but do not remove yet
						fxDuration := g.spawnEmplacementDestroyEffects(s)
						s.SetDestroyCounter(fxDuration)
					} else if destroyCounter == 1 {
						// delete when the counter is basically done (to differentiate with default int value 0)
						g.sprites.deleteEmplacementSprite(s)
					} else {
						s.Update(g.player.Pos())
						s.SetDestroyCounter(destroyCounter - 1)
					}
					break
				}

				g.updateEmplacementPosition(s)
				s.Update(g.player.Pos())
				g.updateWeaponCooldowns(model.EntityUnit(s.Entity))
			}

			return true
		})
	}
}

func (g *Game) updateMechPosition(s *render.MechSprite) {
	if s.Mech().Powered() != model.POWER_ON {
		// TODO: refactor to use same update logic from player shutdown
		s.SetVelocity(0)
		s.SetVelocityZ(0)

		if s.Mech().Heat() < 0.7*s.Mech().MaxHeat() {
			s.Mech().SetPowered(model.POWER_ON)
		}
		s.Mech().Update()
		return
	}

	// TODO: give mechs a bit more of a brain than this
	sPosition := s.Pos()
	if len(s.PatrolPath) > 0 {
		// make sure there's movement towards the next patrol point
		patrolX, patrolY := s.PatrolPath[s.PatrolPathIndex][0], s.PatrolPath[s.PatrolPathIndex][1]
		patrolLine := geom.Line{X1: sPosition.X, Y1: sPosition.Y, X2: patrolX, Y2: patrolY}

		// TODO: do something about this velocity
		s.SetVelocity(0.02 * s.Scale())

		angle := patrolLine.Angle()
		dist := patrolLine.Distance()

		if dist <= s.Velocity() {
			// start movement towards next patrol point
			s.PatrolPathIndex += 1
			if s.PatrolPathIndex >= len(s.PatrolPath) {
				// loop back towards first patrol point
				s.PatrolPathIndex = 0
			}
			g.updateMechPosition(s)
			return
		} else {
			// keep movements towards current patrol point
			s.SetHeading(angle)
		}
	}

	if s.Mech().Update() {
		// TODO: refactor to use same update function as g.updatePlayer()

		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		posZ, velocityZ := s.PosZ(), s.VelocityZ()
		if velocityZ != 0 {
			posZ += velocityZ
		}

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, newPosZ, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, posZ, false)
		if isCollision {
			// TODO: collision damage against units based on mech and speed

			// if mech is falling to the ground, let it land!
			if velocityZ < 0 && posZ <= 0 && newPosZ == 0 {
				s.SetPosZ(newPosZ)
			}
		} else {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}
	}
}

func (g *Game) updateVehiclePosition(s *render.VehicleSprite) {
	// TODO: give units a bit more of a brain than this
	if s.Vehicle().Powered() != model.POWER_ON {
		// TODO: refactor to use same update logic from player shutdown
		s.SetVelocity(0)
		s.SetVelocityZ(0)

		if s.Vehicle().Heat() < 0.7*s.Vehicle().MaxHeat() {
			s.Vehicle().SetPowered(model.POWER_ON)
		}
		s.Vehicle().Update()
		return
	}

	sPosition := s.Pos()
	if len(s.PatrolPath) > 0 {
		// make sure there's movement towards the next patrol point
		patrolX, patrolY := s.PatrolPath[s.PatrolPathIndex][0], s.PatrolPath[s.PatrolPathIndex][1]
		patrolLine := geom.Line{X1: sPosition.X, Y1: sPosition.Y, X2: patrolX, Y2: patrolY}

		// TODO: do something about this velocity
		s.SetVelocity(0.02 * s.Scale())

		angle := patrolLine.Angle()
		dist := patrolLine.Distance()

		if dist <= s.Velocity() {
			// start movement towards next patrol point
			s.PatrolPathIndex += 1
			if s.PatrolPathIndex >= len(s.PatrolPath) {
				// loop back towards first patrol point
				s.PatrolPathIndex = 0
			}
			g.updateVehiclePosition(s)
		} else {
			// keep movements towards current patrol point
			s.SetHeading(angle)
		}
	}

	if s.Velocity() != 0 {
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, newPosZ, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PosZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetHeading(randFloat(-geom.Pi, geom.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}
	}
}

func (g *Game) updateVTOLPosition(s *render.VTOLSprite) {
	// TODO: give units a bit more of a brain than this
	if s.VTOL().Powered() != model.POWER_ON {
		// TODO: refactor to use same update logic from player shutdown
		s.SetVelocity(0)
		s.SetVelocityZ(0)

		if s.VTOL().Heat() < 0.7*s.VTOL().MaxHeat() {
			s.VTOL().SetPowered(model.POWER_ON)
		}
		s.VTOL().Update()
		return
	}

	sPosition := s.Pos()
	if len(s.PatrolPath) > 0 {
		// make sure there's movement towards the next patrol point
		patrolX, patrolY := s.PatrolPath[s.PatrolPathIndex][0], s.PatrolPath[s.PatrolPathIndex][1]
		patrolLine := geom.Line{X1: sPosition.X, Y1: sPosition.Y, X2: patrolX, Y2: patrolY}

		// TODO: do something about this velocity
		s.SetVelocity(0.02 * s.Scale())

		angle := patrolLine.Angle()
		dist := patrolLine.Distance()

		if dist <= s.Velocity() {
			// start movement towards next patrol point
			s.PatrolPathIndex += 1
			if s.PatrolPathIndex >= len(s.PatrolPath) {
				// loop back towards first patrol point
				s.PatrolPathIndex = 0
			}
			g.updateVTOLPosition(s)
		} else {
			// keep movements towards current patrol point
			s.SetHeading(angle)
		}
	}

	if s.Velocity() != 0 {
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, newPosZ, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PosZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetHeading(randFloat(-geom.Pi, geom.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}
	}
}

func (g *Game) updateInfantryPosition(s *render.InfantrySprite) {
	// TODO: give units a bit more of a brain than this
	if s.Infantry().Powered() != model.POWER_ON {
		// TODO: refactor to use same update logic from player shutdown
		s.SetVelocity(0)
		s.SetVelocityZ(0)

		if s.Infantry().Heat() < 0.7*s.Infantry().MaxHeat() {
			s.Infantry().SetPowered(model.POWER_ON)
		}
		s.Infantry().Update()
		return
	}

	sPosition := s.Pos()
	if len(s.PatrolPath) > 0 {
		// make sure there's movement towards the next patrol point
		patrolX, patrolY := s.PatrolPath[s.PatrolPathIndex][0], s.PatrolPath[s.PatrolPathIndex][1]
		patrolLine := geom.Line{X1: sPosition.X, Y1: sPosition.Y, X2: patrolX, Y2: patrolY}

		// TODO: do something about this velocity
		s.SetVelocity(0.02 * s.Scale())

		angle := patrolLine.Angle()
		dist := patrolLine.Distance()

		if dist <= s.Velocity() {
			// start movement towards next patrol point
			s.PatrolPathIndex += 1
			if s.PatrolPathIndex >= len(s.PatrolPath) {
				// loop back towards first patrol point
				s.PatrolPathIndex = 0
			}
			g.updateInfantryPosition(s)
		} else {
			// keep movements towards current patrol point
			s.SetHeading(angle)
		}
	}

	if s.Velocity() != 0 {
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, newPosZ, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PosZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetHeading(randFloat(-geom.Pi, geom.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}
	}
}

func (g *Game) updateEmplacementPosition(s *render.EmplacementSprite) {
	// TODO: give turrets a bit more of a brain than this
	if s.Emplacement().Powered() != model.POWER_ON {
		// TODO: refactor to use same update logic from player shutdown
		if s.Emplacement().Heat() < 0.7*s.Emplacement().MaxHeat() {
			s.Emplacement().SetPowered(model.POWER_ON)
		}
		return
	}
}

func (g *Game) updateSpritePosition(s *render.Sprite) bool {
	if s.Velocity() != 0 || s.VelocityZ() != 0 {
		sPosition := s.Pos()
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2
		zCheck := s.PosZ() + s.VelocityZ()

		newPos, newPosZ, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, zCheck, false)
		if isCollision {
			return true
		} else {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}
	}
	return false
}

func (g *Game) updateWeaponCooldowns(unit model.Unit) {
	if unit == nil {
		return
	}
	armament := unit.Armament()
	if len(armament) == 0 {
		return
	}

	for _, weapon := range armament {
		weapon.DecreaseCooldown(model.SECONDS_PER_TICK)
	}
}

func (g *Game) randomUnit(unitResourceType string) model.Unit {
	// TODO: make it useful for random unit of any unit type, or within some tonnage range
	switch unitResourceType {
	case model.MechResourceType:
		model.Randish.Seed(time.Now().UnixNano())
		mechResources := g.resources.GetMechResourceList()
		randIndex := model.Randish.Intn(len(mechResources))
		randResource := mechResources[randIndex]
		return g.createModelMechFromResource(randResource)
	default:
		panic(fmt.Errorf("currently unable to handle random model.Unit for resource type %v", unitResourceType))
	}
}

func randFloat(min, max float64) float64 {
	return model.RandFloat64In(min, max)
}

func exit(rc int) {
	os.Exit(rc)
}
