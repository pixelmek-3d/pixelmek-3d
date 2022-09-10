package game

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sync"

	"image/color"
	_ "image/png"

	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/spf13/viper"
)

const (
	//--RaycastEngine constants
	//--set constant, texture size to be the wall (and sprite) texture size--//
	texWidth = 256

	// distance to keep away from walls and obstacles to avoid clipping
	// TODO: may want a smaller distance to test vs. sprites
	clipDistance = 0.1
)

// Game - This is the main type for your game.
type Game struct {
	menu   DemoMenu
	paused bool

	//--create slicer and declare slices--//
	tex *TextureHandler

	// window resolution and scaling
	screenWidth  int
	screenHeight int
	renderScale  float64
	fullscreen   bool
	vsync        bool
	fovDegrees   float64
	fovDepth     float64

	//--viewport width / height--//
	width  int
	height int

	player     *model.Player
	crosshairs *model.Crosshairs
	reticle    *model.TargetReticle

	hudScale float64
	hudRGBA  color.RGBA

	//--define camera and renderer--//
	camera *raycaster.Camera

	mouseMode      MouseMode
	mouseX, mouseY int

	// zoom settings
	zoomFovDepth float64

	renderDistance  float64
	clutterDistance float64

	// lighting settings
	lightFalloff       float64
	globalIllumination float64
	minLightRGB        color.NRGBA
	maxLightRGB        color.NRGBA

	// Mission and map
	mission      *model.Mission
	collisionMap []geom.Line

	sprites *SpriteHandler
	clutter *ClutterHandler

	mapWidth, mapHeight int

	debug bool
}

// NewGame - Allows the game to perform any initialization it needs to before starting to run.
// This is where it can query for any required services and load any non-graphic
// related content.  Calling base.Initialize will enumerate through any components
// and initialize them as well.
func NewGame() *Game {
	// initialize Game object
	g := new(Game)

	g.initConfig()

	ebiten.SetWindowTitle("PixelMek 3D")

	// default TPS is 60
	// ebiten.SetMaxTPS(60)

	// use scale to keep the desired window width and height
	g.setResolution(g.screenWidth, g.screenHeight)
	g.setRenderScale(g.renderScale)
	g.setFullscreen(false)
	g.setVsyncEnabled(true)

	// load mission
	var err error
	missionPath := "trial.yaml"
	g.mission, err = model.LoadMission(missionPath)
	if err != nil {
		log.Println("Error loading mission", missionPath)
		log.Println(err)
		exit(1)
	}

	// load texture handler
	g.tex = NewTextureHandler(g.mission.Map())

	g.collisionMap = g.mission.Map().GetCollisionLines(clipDistance)
	worldMap := g.mission.Map().Level(0)
	g.mapWidth = len(worldMap)
	g.mapHeight = len(worldMap[0])

	// init player model
	angleDegrees := 60.0
	g.player = model.NewPlayer(8.5, 3.5, geom.Radians(angleDegrees), 0)
	g.player.SetCollisionRadius(clipDistance) // TODO: get from player mech
	g.player.SetCollisionHeight(0.5)          // TODO: also get from player mech

	// load map and mission content once when first run
	g.loadContent()

	// init mouse movement mode
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	g.mouseMode = MouseModeMove
	g.mouseX, g.mouseY = math.MinInt32, math.MinInt32

	//--init camera and renderer--//
	g.camera = raycaster.NewCamera(g.width, g.height, texWidth, g.mission.Map(), g.tex)
	g.camera.SetRenderDistance(g.renderDistance)

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
	g.camera.SetLightRGB(g.minLightRGB, g.maxLightRGB)

	// initialize camera to player position
	g.updatePlayerCamera(true)
	g.fovDegrees = g.camera.FovAngle() // TODO: store and load from config file
	g.fovDepth = g.camera.FovDepth()

	g.zoomFovDepth = 2.0

	// initialize clutter
	if g.clutter != nil {
		g.clutter.Update(g, true)
	}

	// init menu system
	g.menu = mainMenu()

	return g
}

func (g *Game) initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.SetEnvPrefix("pixelmek")

	userHomePath, _ := os.UserHomeDir()
	if userHomePath != "" {
		userHomePath = userHomePath + "/.pixelmek-3d"
		viper.AddConfigPath(userHomePath)
	}
	viper.AddConfigPath(".")

	// set default config values
	viper.SetDefault("debug", false)

	viper.SetDefault("screen.width", 1024)
	viper.SetDefault("screen.height", 768)
	viper.SetDefault("screen.renderScale", 1.0)
	viper.SetDefault("screen.renderDistance", -1)
	viper.SetDefault("screen.clutterDistance", 10.0)

	viper.SetDefault("hud.scale", 1.0)
	viper.SetDefault("hud.color.red", 100)
	viper.SetDefault("hud.color.green", 255)
	viper.SetDefault("hud.color.blue", 230)
	viper.SetDefault("hud.color.alpha", 255)

	err := viper.ReadInConfig()
	if err != nil && g.debug {
		log.Print(err)
	}

	// get config values
	g.screenWidth = viper.GetInt("screen.width")
	g.screenHeight = viper.GetInt("screen.height")
	g.renderScale = viper.GetFloat64("screen.renderScale")
	g.renderDistance = viper.GetFloat64("screen.renderDistance")
	g.clutterDistance = viper.GetFloat64("screen.clutterDistance")

	g.hudScale = viper.GetFloat64("hud.scale")
	g.hudRGBA = color.RGBA{
		R: uint8(viper.GetUint("hud.color.red")),
		G: uint8(viper.GetUint("hud.color.green")),
		B: uint8(viper.GetUint("hud.color.blue")),
		A: uint8(viper.GetUint("hud.color.alpha")),
	}

	g.debug = viper.GetBool("debug")
}

func (g *Game) SaveConfig() error {
	userConfigPath, _ := os.UserHomeDir()
	if userConfigPath == "" {
		userConfigPath = "./"
	}
	userConfigPath += "/.pixelmek-3d"

	userConfig := userConfigPath + "/config.json"
	log.Print("Saving config file ", userConfig)

	if _, err := os.Stat(userConfigPath); os.IsNotExist(err) {
		err = os.MkdirAll(userConfigPath, os.ModePerm)
		if err != nil {
			log.Print(err)
			return err
		}
	}
	err := viper.WriteConfigAs(userConfig)
	if err != nil {
		log.Print(err)
	}

	return err
}

// Run is the Ebiten Run loop caller
func (g *Game) Run() {
	// On browsers, let's use fullscreen so that this is playable on any browsers.
	// It is planned to ignore the given 'scale' apply fullscreen automatically on browsers (#571).
	if runtime.GOARCH == "js" || runtime.GOOS == "js" {
		ebiten.SetFullscreen(true)
	}

	g.paused = false

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	w, h := int(float64(g.screenWidth)*g.renderScale), int(float64(g.screenHeight)*g.renderScale)
	g.menu.layout(w, h)
	return int(w), int(h)
}

// Update - Allows the game to run logic such as updating the world, gathering input, and playing audio.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// handle input (when paused making sure only to allow input for closing menu so it can be unpaused)
	g.handleInput()

	if !g.paused {
		// Perform logical updates
		g.updateProjectiles()
		g.updateSprites()

		if g.clutter != nil {
			g.clutter.Update(g, false)
		}

		// handle player camera movement
		g.updatePlayerCamera(false)
	}

	// update the menu (if active)
	g.menu.update(g)

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	// Put projectiles together with sprites for raycasting both as sprites
	raycastSprites := g.getRaycastSprites()

	// Update camera (calculate raycast)
	g.camera.Update(raycastSprites)

	// Render raycast to screen
	g.camera.Draw(screen)

	// draw target reticle
	g.drawTargetReticle(screen)

	// draw crosshairs
	g.drawCrosshairs(screen)

	// draw menu (if active)
	g.menu.draw(screen)

	// draw FPS/TPS counter debug display
	fps := fmt.Sprintf("FPS: %f\nTPS: %f/%v", ebiten.CurrentFPS(), ebiten.CurrentTPS(), ebiten.MaxTPS())
	ebitenutil.DebugPrint(screen, fps)
}

func (g *Game) setFullscreen(fullscreen bool) {
	g.fullscreen = fullscreen
	ebiten.SetFullscreen(fullscreen)
}

func (g *Game) setResolution(screenWidth, screenHeight int) {
	g.screenWidth, g.screenHeight = screenWidth, screenHeight
	ebiten.SetWindowSize(screenWidth, screenHeight)
	g.setRenderScale(g.renderScale)
}

func (g *Game) setRenderScale(renderScale float64) {
	g.renderScale = renderScale
	g.width = int(math.Floor(float64(g.screenWidth) * g.renderScale))
	g.height = int(math.Floor(float64(g.screenHeight) * g.renderScale))
	if g.camera != nil {
		g.camera.SetViewSize(g.width, g.height)
	}
}

func (g *Game) setVsyncEnabled(enableVsync bool) {
	g.vsync = enableVsync
	if enableVsync {
		ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)
	} else {
		ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	}
}

func (g *Game) setFovAngle(fovDegrees float64) {
	g.fovDegrees = fovDegrees
	g.camera.SetFovAngle(fovDegrees, 1.0)
}

// Move player by move speed in the forward/backward direction
func (g *Game) Move(mSpeed float64) {
	playerPosition := g.player.Position()
	moveLine := geom.LineFromAngle(playerPosition.X, playerPosition.Y, g.player.Angle(), mSpeed)
	g.updatePlayerPosition(moveLine.X2, moveLine.Y2)
}

// Move player by strafe speed in the left/right direction
func (g *Game) Strafe(sSpeed float64) {
	strafeAngle := geom.HalfPi
	if sSpeed < 0 {
		strafeAngle = -strafeAngle
	}
	playerPosition := g.player.Position()
	strafeLine := geom.LineFromAngle(playerPosition.X, playerPosition.Y, g.player.Angle()-strafeAngle, math.Abs(sSpeed))
	g.updatePlayerPosition(strafeLine.X2, strafeLine.Y2)
}

// Rotate player heading angle by rotation speed
func (g *Game) Rotate(rSpeed float64) {
	angle := g.player.Angle() + rSpeed

	pi2 := geom.Pi2
	if angle >= pi2 {
		angle = pi2 - angle
	} else if angle <= -pi2 {
		angle = angle + pi2
	}

	g.player.SetAngle(angle)
	g.player.Moved = true
}

// Update player pitch angle by pitch speed
func (g *Game) Pitch(pSpeed float64) {
	// current raycasting method can only allow up to 45 degree pitch in either direction
	g.player.SetPitch(geom.Clamp(pSpeed+g.player.Pitch(), -math.Pi/8, math.Pi/4))
	g.player.Moved = true
}

func (g *Game) Stand() {
	g.player.CameraZ = 0.5
	g.player.Moved = true
}

func (g *Game) IsStanding() bool {
	return g.player.CameraZ == 0.5
}

func (g *Game) Jump() {
	g.player.CameraZ = 0.9
	g.player.Moved = true
}

func (g *Game) Crouch() {
	g.player.CameraZ = 0.3
	g.player.Moved = true
}

func (g *Game) Prone() {
	g.player.CameraZ = 0.1
	g.player.Moved = true
}

func (g *Game) fireWeapon() {
	p := g.player.TestProjectile
	if p == nil {
		return
	}

	if g.player.TestCooldown > 0 {
		return
	}

	// spawning projectile at offsets from player's center point of view
	pAngle, pPitch := g.player.Angle(), g.player.Pitch()

	// firing test projectiles
	pVelocity := 16.0

	pX, pY, pZ := g.weaponPosition3D(0, -0.1)
	projectile := p.SpawnProjectile(pX, pY, pZ, pAngle, pPitch, pVelocity, g.player.Entity)
	if projectile != nil {
		g.sprites.addProjectile(projectile)
		g.player.TestCooldown = 10
	}

	pX, pY, pZ = g.weaponPosition3D(-0.1, -0.2)
	projectile = p.SpawnProjectile(pX, pY, pZ, pAngle, pPitch, pVelocity, g.player.Entity)
	if projectile != nil {
		g.sprites.addProjectile(projectile)
		g.player.TestCooldown = 10
	}

	pX, pY, pZ = g.weaponPosition3D(0.1, -0.2)
	projectile = p.SpawnProjectile(pX, pY, pZ, pAngle, pPitch, pVelocity, g.player.Entity)
	if projectile != nil {
		g.sprites.addProjectile(projectile)
		g.player.TestCooldown = 10
	}
}

func (g *Game) fireTestWeaponAtPlayer() {
	// Just for testing!
	p := g.player.TestProjectile
	if p == nil {
		return
	}

	if g.player.TestCooldown > 0 {
		return
	}

	g.sprites.sprites[MechSpriteType].Range(func(k, _ interface{}) bool {
		// firing test projectile at player
		m := k.(*model.MechSprite)
		pVelocity := 16.0

		playerPosition := g.player.Position()
		mechPosition := m.Position()
		pX, pY, pZ := mechPosition.X, mechPosition.Y, m.PositionZ()+0.4

		pLine := geom3d.Line3d{X1: pX, Y1: pY, Z1: pZ, X2: playerPosition.X, Y2: playerPosition.Y, Z2: randFloat(0.1, 0.7)}
		pHeading, pPitch := pLine.Heading(), pLine.Pitch()
		projectile := p.SpawnProjectile(pX, pY, pZ, pHeading, pPitch, pVelocity, m.Entity)
		if projectile != nil {
			g.sprites.addProjectile(projectile)
			g.player.TestCooldown = 10
		}

		return true
	})
}

// weaponPosition3D gets the X, Y and Z axis offsets needed for weapon projectile spawned from a 2-D sprite reference
func (g *Game) weaponPosition3D(weaponOffX, weaponOffY float64) (float64, float64, float64) {
	playerPosition := g.player.Position()
	wX, wY, wZ := playerPosition.X, playerPosition.Y, g.player.CameraZ+weaponOffY

	if weaponOffX == 0 {
		// no X/Y position adjustments needed
		return wX, wY, wZ
	}

	// calculate X,Y based on player orientation angle perpendicular to angle of view
	offAngle := g.player.Angle() + math.Pi/2

	// create line segment using offset angle and X offset to determine 3D position offset of X/Y
	offLine := geom.LineFromAngle(0, 0, offAngle, weaponOffX)
	wX, wY = wX+offLine.X2, wY+offLine.Y2

	return wX, wY, wZ
}

// Update camera to match player position and orientation
func (g *Game) updatePlayerCamera(forceUpdate bool) {
	if !g.player.Moved && !forceUpdate {
		// only update camera position if player moved or forceUpdate set
		return
	}

	// reset player moved flag to only update camera when necessary
	g.player.Moved = false

	g.camera.SetPosition(g.player.Position().Copy())
	g.camera.SetPositionZ(g.player.CameraZ)
	g.camera.SetHeadingAngle(g.player.Angle())
	g.camera.SetPitchAngle(g.player.Pitch())
}

func (g *Game) updatePlayerPosition(newX, newY float64) {
	// Update player position
	newPos, isCollision, collisions := g.getValidMove(g.player.Entity, newX, newY, g.player.PositionZ(), true)
	if !newPos.Equals(g.player.Position()) {
		g.player.SetPosition(newPos)
		g.player.Moved = true
	}

	if isCollision && len(collisions) > 0 {
		// apply damage to the first sprite entity that was hit
		collisionEntity := collisions[0]

		collisionDamage := 1.0 // TODO: determine collision damage based on player mech and speed
		collisionEntity.entity.DamageHitPoints(collisionDamage)
		fmt.Printf("collided for %0.1f (HP: %0.1f)\n", collisionDamage, collisionEntity.entity.HitPoints())
	}
}

func (g *Game) updateProjectiles() {
	// Update animated projectile movement
	if g.player.TestCooldown > 0 {
		g.player.TestCooldown--
	}

	// perform concurrent projectile updates
	var wg sync.WaitGroup

	g.sprites.sprites[ProjectileSpriteType].Range(func(k, _ interface{}) bool {
		p := k.(*model.Projectile)
		p.Lifespan--
		if p.Lifespan <= 0 {
			g.sprites.deleteProjectile(p)
			return true
		}

		wg.Add(1)
		go g.asyncProjectileUpdate(p, &wg)

		return true
	})

	// Update animated effects
	g.sprites.sprites[EffectSpriteType].Range(func(k, _ interface{}) bool {
		e := k.(*model.Effect)
		e.Update(g.player.Position())
		if e.LoopCounter() >= e.LoopCount {
			g.sprites.deleteEffect(e)
		}

		return true
	})

	wg.Wait()
}

func (g *Game) asyncProjectileUpdate(p *model.Projectile, wg *sync.WaitGroup) {
	defer wg.Done()

	if p.Velocity() != 0 {
		pPosition := p.Position()
		trajectory := geom3d.Line3dFromAngle(pPosition.X, pPosition.Y, p.PositionZ(), p.Angle(), p.Pitch(), p.Velocity())
		xCheck := trajectory.X2
		yCheck := trajectory.Y2
		zCheck := trajectory.Z2

		newPos, isCollision, collisions := g.getValidMove(p.Entity, xCheck, yCheck, zCheck, false)
		if isCollision || p.PositionZ() <= 0 {
			// for testing purposes, projectiles instantly get deleted when collision occurs
			//g.sprites.deleteProjectile(p)
			p.Lifespan = -1

			var collisionEntity *EntityCollision
			if len(collisions) > 0 {
				// apply damage to the first sprite entity that was hit
				collisionEntity = collisions[0]

				if collisionEntity.entity == g.player.Entity {
					// TODO: visual response to player being hit
					println("ouch!")
				} else {
					collisionEntity.entity.DamageHitPoints(p.Damage)
					fmt.Printf("hit for %0.1f (HP: %0.1f)\n", p.Damage, collisionEntity.entity.HitPoints())
				}
			}

			// make a sprite/wall getting hit by projectile cause some visual effect
			if p.ImpactEffect.Sprite != nil {
				if collisionEntity != nil {
					// use the first collision point to place effect at
					newPos = collisionEntity.collision
				}

				// TODO: give impact effect optional ability to have some velocity based on the projectile movement upon impact if it didn't hit a wall
				effect := p.SpawnEffect(newPos.X, newPos.Y, p.PositionZ(), p.Angle(), p.Pitch())

				g.sprites.addEffect(effect)
			}

		} else {
			p.SetPosition(newPos)
			p.SetPositionZ(zCheck)
		}
	}
	p.Update(g.player.Position())
}

func (g *Game) updateSprites() {
	// Update for animated sprite movement
	g.sprites.sprites[MapSpriteType].Range(func(k, _ interface{}) bool {
		s := k.(*model.Sprite)
		if s.HitPoints() <= 0 {
			// TODO: implement sprite destruction animation
			g.sprites.deleteMapSprite(s)
		}

		g.updateSpritePosition(s)
		s.Update(g.player.Position())

		return true
	})

	// Updates for animated mech sprite movement
	g.sprites.sprites[MechSpriteType].Range(func(k, _ interface{}) bool {
		s := k.(*model.MechSprite)
		// TODO: implement mech armor and structure instead of direct HP
		if s.HitPoints() <= 0 {
			// TODO: implement mech destruction animation
			g.sprites.deleteMechSprite(s)
		}

		g.updateMechPosition(s)
		s.Update(g.player.Position())

		return true
	})
}

func (g *Game) updateMechPosition(s *model.MechSprite) {
	// TODO: give mechs a bit more of a brain than this
	sPosition := s.Position()
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
		} else {
			// keep movements towards current patrol point
			s.SetAngle(angle)
		}
	}

	if s.Velocity() != 0 {
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Angle(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PositionZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetAngle(randFloat(-math.Pi, math.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPosition(newPos)
		}
	}
}

func (g *Game) updateSpritePosition(s *model.Sprite) {
	if s.Velocity() != 0 {
		sPosition := s.Position()
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Angle(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PositionZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetAngle(randFloat(-math.Pi, math.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPosition(newPos)
		}
	}
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func exit(rc int) {
	os.Exit(rc)
}
