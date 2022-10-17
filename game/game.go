package game

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"

	"image/color"
	_ "image/png"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"

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

	resources *model.ModelResources

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

	player     *render.Player
	crosshairs *render.Crosshairs
	reticle    *render.TargetReticle

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

	sprites                *SpriteHandler
	clutter                *ClutterHandler
	collisonSpriteTypes    map[SpriteType]struct{}
	interactiveSpriteTypes map[SpriteType]struct{}

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

	g.initInteractiveTypes()
	g.initCollisionTypes()

	ebiten.SetWindowTitle("PixelMek 3D")
	ebiten.SetMaxTPS(int(model.TICKS_PER_SECOND))

	// use scale to keep the desired window width and height
	g.setResolution(g.screenWidth, g.screenHeight)
	g.setRenderScale(g.renderScale)
	g.setFullscreen(false)
	g.setVsyncEnabled(true)

	var err error
	g.resources, err = model.LoadModelResources()
	if err != nil {
		log.Println("Error loading models:")
		log.Println(err)
		exit(1)
	}

	// load mission
	missionPath := "trial.yaml"
	g.mission, err = model.LoadMission(missionPath)
	if err != nil {
		log.Println("Error loading mission: ", missionPath)
		log.Println(err)
		exit(1)
	}

	// load texture handler
	g.tex = NewTextureHandler(g.mission.Map())

	g.collisionMap = g.mission.Map().GetCollisionLines(clipDistance)
	worldMap := g.mission.Map().Level(0)
	g.mapWidth = len(worldMap)
	g.mapHeight = len(worldMap[0])

	// load map and mission content once when first run
	g.loadContent()

	// init player model
	pX, pY, pDegrees := 8.5, 3.5, 60.0                  // TODO: get from mission
	pMech := g.createModelMech("timberwolf_prime.yaml") // TODO: get from mission, initially?
	g.player = render.NewPlayer(pMech, pX, pY, geom.Radians(pDegrees), 0)
	g.player.SetCollisionRadius(pMech.CollisionRadius())
	g.player.SetCollisionHeight(pMech.CollisionHeight())

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

		// handle player weapon updates
		g.updateWeaponCooldowns(g.player.Entity)

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

	// store raycasted convergence point for next Update
	g.player.ConvergenceDistance = g.camera.GetConvergenceDistance()
	g.player.ConvergencePoint = g.camera.GetConvergencePoint()

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
	playerPosition := g.player.Pos()
	moveLine := geom.LineFromAngle(playerPosition.X, playerPosition.Y, g.player.Angle(), mSpeed)
	g.updatePlayerPosition(moveLine.X2, moveLine.Y2)
}

// Move player by strafe speed in the left/right direction
func (g *Game) Strafe(sSpeed float64) {
	strafeAngle := geom.HalfPi
	if sSpeed < 0 {
		strafeAngle = -strafeAngle
	}
	playerPosition := g.player.Pos()
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
	// weapons test from model
	armament := g.player.Armament()
	if len(armament) == 0 {
		return
	}

	// in case convergence point not set, use player heading and pitch
	pAngle, pPitch := g.player.Angle(), g.player.Pitch()
	convergencePoint := g.player.ConvergencePoint
	// convergenceDistance := g.player.ConvergenceDistance

	for _, weapon := range armament {
		if weapon.Cooldown() > 0 {
			continue
		}

		var projectile *model.Projectile
		if convergencePoint == nil {
			projectile = weapon.SpawnProjectile(pAngle, pPitch, g.player.Entity)
		} else {
			projectile = weapon.SpawnProjectileToward(convergencePoint, g.player.Entity)
		}

		if projectile != nil {
			weapon.TriggerCooldown()

			// TODO: make projectiles spawned by player use their head-on facing angle for the first several frames to avoid
			//       them using a facing that looks weird (like lasers are doing when fired from arm location)

			pTemplate := projectileSpriteForWeapon(weapon)
			pSprite := pTemplate.Clone()
			pSprite.Entity = projectile
			g.sprites.addProjectile(pSprite)

			// use go routine to handle creation of multiple projectiles after time delay
			if weapon.ProjectileCount() > 1 {
				for i := 1; i < weapon.ProjectileCount(); i++ {
					go g.delayedSpawnProjectile(float64(i)*weapon.ProjectileDelay(), weapon, g.player.Entity)
				}
			}
		}
	}
}

func (g *Game) delayedSpawnProjectile(delay float64, w model.Weapon, e model.Entity) {
	// TODO: instead of go routine, just have a countdown that gets checked during Update?
	time.Sleep(time.Duration(delay * float64(time.Second)))

	var projectile *model.Projectile

	convergencePoint := g.player.ConvergencePoint
	if e != g.player.Entity || convergencePoint == nil {
		projectile = w.SpawnProjectile(e.Angle(), e.Pitch(), e)
	} else {
		projectile = w.SpawnProjectileToward(convergencePoint, e)
	}

	if projectile != nil {
		pTemplate := projectileSpriteForWeapon(w)
		pSprite := pTemplate.Clone()
		pSprite.Entity = projectile
		g.sprites.addProjectile(pSprite)
	}
}

func (g *Game) fireTestWeaponAtPlayer() {
	// Just for testing! Firing test projectiles at player
	playerPosition := g.player.Pos()
	for spriteType := range g.sprites.sprites {
		g.sprites.sprites[spriteType].Range(func(k, _ interface{}) bool {
			var pX, pY, pZ float64
			var entity model.Entity

			switch spriteType {
			case MechSpriteType:
				s := k.(*render.MechSprite)
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()+0.4
				entity = s.Entity

			case VehicleSpriteType:
				s := k.(*render.VehicleSprite)
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()+0.2
				entity = s.Entity

			case VTOLSpriteType:
				s := k.(*render.VTOLSprite)
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()
				entity = s.Entity

			case InfantrySpriteType:
				s := k.(*render.InfantrySprite)
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()+0.1
				entity = s.Entity
			}

			if entity == nil {
				return true
			}

			pLine := geom3d.Line3d{X1: pX, Y1: pY, Z1: pZ, X2: playerPosition.X, Y2: playerPosition.Y, Z2: randFloat(0.1, 0.7)}
			pHeading, pPitch := pLine.Heading(), pLine.Pitch()

			// TESTING: needed until turret heading is separated from heading angle so projectiles come from correct postion
			entity.SetAngle(pHeading)
			entity.SetPitch(pPitch)

			for _, weapon := range entity.Armament() {
				if weapon.Cooldown() > 0 {
					continue
				}

				projectile := weapon.SpawnProjectile(pHeading, pPitch, entity)
				if projectile != nil {
					// TODO: add muzzle flash effect on being fired at
					weapon.TriggerCooldown()

					pTemplate := projectileSpriteForWeapon(weapon)
					pSprite := pTemplate.Clone()
					pSprite.Entity = projectile
					g.sprites.addProjectile(pSprite)

					// use go routine to handle creation of multiple projectiles after time delay
					if weapon.ProjectileCount() > 1 {
						for i := 1; i < weapon.ProjectileCount(); i++ {
							go g.delayedSpawnProjectile(float64(i)*weapon.ProjectileDelay(), weapon, entity)
						}
					}
				}
			}

			return true
		})
	}
}

// Update camera to match player position and orientation
func (g *Game) updatePlayerCamera(forceUpdate bool) {
	if !g.player.Moved && !forceUpdate {
		// only update camera position if player moved or forceUpdate set
		return
	}

	// reset player moved flag to only update camera when necessary
	g.player.Moved = false

	g.camera.SetPosition(g.player.Pos().Copy())
	g.camera.SetPositionZ(g.player.CameraZ)
	g.camera.SetHeadingAngle(g.player.Angle())
	g.camera.SetPitchAngle(g.player.Pitch())
}

func (g *Game) updatePlayerPosition(newX, newY float64) {
	// Update player position
	newPos, isCollision, collisions := g.getValidMove(g.player.Entity, newX, newY, g.player.PosZ(), true)
	if !newPos.Equals(g.player.Pos()) {
		g.player.SetPos(newPos)
		g.player.Moved = true
	}

	if isCollision && len(collisions) > 0 {
		// apply damage to the first sprite entity that was hit
		collisionEntity := collisions[0]

		collisionDamage := 1.0 // TODO: determine collision damage based on player mech and speed
		collisionEntity.entity.ApplyDamage(collisionDamage)
		fmt.Printf("collided for %0.1f (HP: %0.1f)\n", collisionDamage, collisionEntity.entity.ArmorPoints())
	}
}

func (g *Game) updateProjectiles() {
	// perform concurrent projectile updates
	var wg sync.WaitGroup

	g.sprites.sprites[ProjectileSpriteType].Range(func(k, _ interface{}) bool {
		p := k.(*render.ProjectileSprite)
		p.DecreaseLifespan(1)
		if p.Lifespan() <= 0 {
			// TODO: have projectiles fade out slowly, maybe do less damage at extreme range?
			g.sprites.deleteProjectile(p)
			return true
		}

		wg.Add(1)
		go g.asyncProjectileUpdate(p, &wg)

		return true
	})

	// Update animated effects
	g.sprites.sprites[EffectSpriteType].Range(func(k, _ interface{}) bool {
		e := k.(*render.EffectSprite)
		e.Update(g.player.Pos())
		if e.LoopCounter() >= e.LoopCount {
			g.sprites.deleteEffect(e)
		}

		return true
	})

	wg.Wait()
}

func (g *Game) asyncProjectileUpdate(p *render.ProjectileSprite, wg *sync.WaitGroup) {
	defer wg.Done()

	if p.Velocity() != 0 {
		pPosition := p.Pos()
		trajectory := geom3d.Line3dFromAngle(pPosition.X, pPosition.Y, p.PosZ(), p.Angle(), p.Pitch(), p.Velocity())
		xCheck := trajectory.X2
		yCheck := trajectory.Y2
		zCheck := trajectory.Z2

		newPos, isCollision, collisions := g.getValidMove(p.Entity, xCheck, yCheck, zCheck, false)
		if isCollision || p.PosZ() <= 0 {
			// for testing purposes, projectiles instantly get deleted when collision occurs
			p.ZeroLifespan()

			var collisionEntity *EntityCollision
			if len(collisions) > 0 {
				// apply damage to the first sprite entity that was hit
				collisionEntity = collisions[0]
				entity := collisionEntity.entity

				if entity == g.player.Entity {
					// TODO: visual response to player being hit
					println("ouch!")
				} else {
					damage := p.Damage()
					entity.ApplyDamage(damage)

					// TODO: visual method for showing damage was done
					hp, maxHP := entity.ArmorPoints()+entity.StructurePoints(), entity.MaxArmorPoints()+entity.MaxStructurePoints()
					percentHP := 100 * (hp / maxHP)
					fmt.Printf("[%0.2f%s] hit for %0.1f (HP: %0.1f/%0.0f)\n", percentHP, "%", damage, hp, maxHP)
				}
			}

			// make a sprite/wall getting hit by projectile cause some visual effect
			if p.ImpactEffect.Sprite != nil {
				if collisionEntity != nil {
					// use the first collision point to place effect at
					newPos = collisionEntity.collision
				}

				// TODO: give impact effect optional ability to have some velocity based on the projectile movement upon impact if it didn't hit a wall
				effect := p.SpawnEffect(newPos.X, newPos.Y, p.PosZ(), p.Angle(), p.Pitch())

				g.sprites.addEffect(effect)
			}

		} else {
			p.SetPos(newPos)
			p.SetPosZ(zCheck)
		}
	}
	p.Update(g.player.Pos())
}

func (g *Game) updateSprites() {
	// Update for animated sprite movement
	for spriteType := range g.sprites.sprites {
		g.sprites.sprites[spriteType].Range(func(k, _ interface{}) bool {

			switch spriteType {
			case MapSpriteType:
				s := k.(*render.Sprite)
				if s.ArmorPoints() <= 0 && s.StructurePoints() <= 0 {
					// TODO: implement sprite destruction animation
					g.sprites.deleteMapSprite(s)
				}

				g.updateSpritePosition(s)
				s.Update(g.player.Pos())

			case MechSpriteType:
				s := k.(*render.MechSprite)
				// TODO: implement mech armor and structure instead of direct HP
				if s.ArmorPoints() <= 0 && s.StructurePoints() <= 0 {
					// TODO: implement unit destruction animation
					g.sprites.deleteMechSprite(s)
				}

				g.updateMechPosition(s)
				s.Update(g.player.Pos())
				g.updateWeaponCooldowns(s.Entity)

			case VehicleSpriteType:
				s := k.(*render.VehicleSprite)
				// TODO: implement vehicle armor and structure instead of direct HP
				if s.ArmorPoints() <= 0 && s.StructurePoints() <= 0 {
					// TODO: implement unit destruction animation
					g.sprites.deleteVehicleSprite(s)
				}

				g.updateVehiclePosition(s)
				s.Update(g.player.Pos())
				g.updateWeaponCooldowns(s.Entity)

			case VTOLSpriteType:
				s := k.(*render.VTOLSprite)
				// TODO: implement vtol armor and structure instead of direct HP
				if s.ArmorPoints() <= 0 && s.StructurePoints() <= 0 {
					// TODO: implement unit destruction animation
					g.sprites.deleteVTOLSprite(s)
				}

				g.updateVTOLPosition(s)
				s.Update(g.player.Pos())
				g.updateWeaponCooldowns(s.Entity)

			case InfantrySpriteType:
				s := k.(*render.InfantrySprite)
				if s.ArmorPoints() <= 0 && s.StructurePoints() <= 0 {
					// TODO: implement unit destruction animation
					g.sprites.deleteInfantrySprite(s)
				}

				g.updateInfantryPosition(s)
				s.Update(g.player.Pos())
				g.updateWeaponCooldowns(s.Entity)
			}

			return true
		})
	}
}

func (g *Game) updateMechPosition(s *render.MechSprite) {
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
		} else {
			// keep movements towards current patrol point
			s.SetAngle(angle)
		}
	}

	if s.Velocity() != 0 {
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Angle(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PosZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetAngle(randFloat(-math.Pi, math.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPos(newPos)
		}
	}
}

func (g *Game) updateVehiclePosition(s *render.VehicleSprite) {
	// TODO: give units a bit more of a brain than this
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
			s.SetAngle(angle)
		}
	}

	if s.Velocity() != 0 {
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Angle(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PosZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetAngle(randFloat(-math.Pi, math.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPos(newPos)
		}
	}
}

func (g *Game) updateVTOLPosition(s *render.VTOLSprite) {
	// TODO: give units a bit more of a brain than this
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
			s.SetAngle(angle)
		}
	}

	if s.Velocity() != 0 {
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Angle(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PosZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetAngle(randFloat(-math.Pi, math.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPos(newPos)
		}
	}
}

func (g *Game) updateInfantryPosition(s *render.InfantrySprite) {
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
			g.updateInfantryPosition(s)
		} else {
			// keep movements towards current patrol point
			s.SetAngle(angle)
		}
	}

	if s.Velocity() != 0 {
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Angle(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PosZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetAngle(randFloat(-math.Pi, math.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPos(newPos)
		}
	}
}

func (g *Game) updateSpritePosition(s *render.Sprite) {
	if s.Velocity() != 0 {
		sPosition := s.Pos()
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Angle(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PosZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetAngle(randFloat(-math.Pi, math.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPos(newPos)
		}
	}
}

func (g *Game) updateWeaponCooldowns(entity model.Entity) {
	if entity == nil {
		return
	}
	armament := entity.Armament()
	if len(armament) == 0 {
		return
	}

	for _, weapon := range armament {
		weapon.DecreaseCooldown(model.SECONDS_PER_TICK)
	}
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func exit(rc int) {
	os.Exit(rc)
}
