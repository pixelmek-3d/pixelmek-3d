package game

import (
	"fmt"
	"image"
	"math"
	"math/rand"
	"os"
	"sort"

	"image/color"

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
	clipDistance = 0.01
)

// Game - This is the main type for your game.
type Game struct {
	scene  Scene
	menu   Menu
	paused bool

	ai             *AIHandler
	aiIgnorePlayer bool

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
	fullscreen   bool
	vsync        bool
	opengl       bool
	fovDegrees   float64
	fovDepth     float64

	//--raycast rendered width / height--//
	renderScale  float64
	renderWidth  int
	renderHeight int

	player    *Player
	playerHUD map[HUDElementType]HUDElement
	fonts     *render.FontHandler

	hudEnabled        bool
	hudFont           string
	hudScale          float64
	hudRGBA           *color.NRGBA
	hudUseCustomColor bool
	hudCrosshairIndex int

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
	collisonSpriteTypes    map[SpriteType]bool
	interactiveSpriteTypes map[SpriteType]bool
	delayedProjectiles     map[*ProjectileSpawn]struct{}

	// Gameplay
	objectives *ObjectivesHandler

	// control options
	throttleDecay bool

	osType     osType
	benchmark  bool
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

	// initialize common resources
	resources.InitResources()

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
	g.scene = NewSplashScene(g)

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

func (g *Game) Resources() *model.ModelResources {
	return g.resources
}

func (g *Game) SetScene(scene Scene) {
	g.scene = scene
}

func (g *Game) LoadMission(missionFile string) (*model.Mission, error) {
	mission, err := model.LoadMission(missionFile)
	if err != nil {
		return nil, err
	}
	g.mission = mission
	return mission, err
}

func (g *Game) initMission() {
	if g.mission == nil {
		panic("g.mission must be set before initMission!")
	}

	missionMap := g.mission.Map()

	// reload texture handler
	if g.tex != nil {
		g.initRenderFloorTex = g.tex.renderFloorTex
	}
	g.tex = NewTextureHandler(missionMap)
	g.tex.renderFloorTex = g.initRenderFloorTex

	// clear mission sprites
	g.sprites.clear()

	g.collisionMap = missionMap.GenerateWallCollisionLines(clipDistance)
	g.mapWidth, g.mapHeight = missionMap.Size()

	// load map and mission content
	g.loadContent()

	// initialize objectives
	g.objectives = NewObjectivesHandler(g, g.mission.Objectives)

	// init player at DZ
	pX, pY, pDegrees := g.mission.DropZone.Position[0], g.mission.DropZone.Position[1], g.mission.DropZone.Heading
	pHeading := geom.Radians(pDegrees)
	g.player.SetPos(&geom.Vector2{X: pX, Y: pY})
	g.player.SetHeading(pHeading)
	g.player.SetTargetHeading(pHeading)
	g.player.SetTurretAngle(pHeading)
	g.player.cameraAngle = pHeading
	g.player.cameraPitch = 0

	// init player as powered off but booting up
	g.player.SetPowered(model.POWER_ON)

	// init player armament for display
	if armament := g.GetHUDElement(HUD_ARMAMENT); armament != nil {
		armament.(*render.Armament).SetWeapons(g.player.Armament())
	}

	// initial mouse position to establish delta
	g.mouseX, g.mouseY = math.MinInt32, math.MinInt32

	//--init camera and renderer--//
	g.camera = raycaster.NewCamera(g.renderWidth, g.renderHeight, texWidth, g.mission.Map(), g.tex)
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

	// initialize AI
	g.ai = NewAIHandler(g)
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

// Move player by strafe speed in the left/right direction
// func (g *Game) Strafe(sSpeed float64) {
// 	strafeAngle := geom.HalfPi
// 	if sSpeed < 0 {
// 		strafeAngle = -strafeAngle
// 	}
// 	playerPosition := g.player.Pos()
// 	strafeLine := geom.LineFromAngle(playerPosition.X, playerPosition.Y, g.player.Heading()-strafeAngle, math.Abs(sSpeed))
// 	g.updatePlayerPosition(strafeLine.X2, strafeLine.Y2, g.player.PosZ())
// }

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

func (g *Game) updateAI() {
	if g.ai == nil {
		return
	}
	g.ai.Update()
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

	prevPos, prevPosZ := g.player.Pos(), g.player.PosZ()
	g.updateUnitPosition(g.player)
	newPos, newPosZ := g.player.Pos(), g.player.PosZ()

	if !(prevPos.Equals(newPos) && prevPosZ == newPosZ) {
		g.player.moved = true

		// check for nav point visits
		for _, nav := range g.mission.NavPoints {
			if nav.Visited() {
				continue
			}

			navX, navY := nav.Position[0], nav.Position[1]
			if model.PointInProximity(1.0, newPos.X, newPos.Y, navX, navY) {
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
	}

	if target == nil || g.IsFriendly(g.player, target) || g.player.Powered() != model.POWER_ON {
		// clear target lock if no target, friendly target, or player is not fully powered on
		g.player.SetTargetLock(0)
	} else {
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
	newTarget := g.spriteInCrosshairs()
	if newTarget != nil && !newTarget.IsDestroyed() {
		g.player.SetTarget(newTarget.Entity)
		return newTarget.Entity
	} else {
		// unset target if nothing is there
		g.player.SetTarget(nil)
	}
	return nil
}

func (g *Game) spriteInCrosshairs() *render.Sprite {
	cSprite := g.player.convergenceSprite
	if cSprite == nil {
		// check for target in crosshairs bounds if not directly at the single center raycasted pixel
		crosshairs := g.GetHUDElement(HUD_CROSSHAIRS).(*render.Crosshairs)
		if crosshairs == nil {
			return nil
		}

		crosshairRect := crosshairs.Rect().Add(
			image.Point{X: (g.screenWidth / 2) - (crosshairs.Width() / 2), Y: (g.screenHeight / 2) - (crosshairs.Height() / 2)})

		var cSpriteArea int
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

				sBounds := s.ScreenRect(g.renderScale)
				if sBounds == nil {
					return true
				}

				// check if sprite bounds intersects general crosshair area
				intersectRect := crosshairRect.Intersect(*sBounds)
				intersectArea := intersectRect.Dx() * intersectRect.Dy()

				if intersectArea > cSpriteArea {
					cSprite = s
					cSpriteArea = intersectArea
				}
				return true
			})
		}
	}

	return cSprite
}

func (g *Game) targetCycle(cycleType TargetCycleType) model.Entity {
	pSprites := g.getProximityUnitSprites(g.player.Pos(), 1000/model.METERS_PER_UNIT)
	targetables := make([]*render.Sprite, 0, len(pSprites))

	if cycleType == TARGET_PREVIOUS {
		// reverse sort by distance
		sort.Slice(pSprites, func(i, j int) bool { return pSprites[i].distance > pSprites[j].distance })
	}

	for _, p := range pSprites {
		s := p.sprite
		if g.IsFriendly(g.player, s.Entity) {
			// skip friendly units
			continue
		}
		targetables = append(targetables, s)
	}

	if len(targetables) == 0 {
		g.player.SetTarget(nil)
		return nil
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

func (g *Game) LoadUnit(unitResourceType, unitFile string) model.Unit {
	// TODO: make it useful for unit of any unit type
	switch unitResourceType {
	case model.MechResourceType:
		if resource, ok := g.resources.Mechs[unitFile]; ok {
			return g.createModelMechFromResource(resource)
		}
	default:
		panic(fmt.Errorf("currently unable to handle load model.Unit for resource type %v", unitResourceType))
	}
	return nil
}

func (g *Game) RandomUnit(unitResourceType string) model.Unit {
	// TODO: make it useful for random unit of any unit type, or within some tonnage range
	switch unitResourceType {
	case model.MechResourceType:
		mechResources := g.resources.GetMechResourceList()
		randIndex := rand.Intn(len(mechResources))
		randResource := mechResources[randIndex]
		return g.createModelMechFromResource(randResource)
	default:
		panic(fmt.Errorf("currently unable to handle random model.Unit for resource type %v", unitResourceType))
	}
}

func (g *Game) IsFriendly(e1, e2 model.Entity) bool {
	if e1 == nil || e2 == nil {
		return false
	}
	if e1 == g.player || e2 == g.player {
		return e1.Team() < 0 && e2.Team() < 0
	}
	return e1.Team() == e2.Team()
}

func randFloat(min, max float64) float64 {
	return model.RandFloat64In(min, max, nil)
}

func exit(rc int) {
	os.Exit(rc)
}
