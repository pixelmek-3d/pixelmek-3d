package game

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"path/filepath"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/raycaster-go/geom"
)

var (
	imageByPath = make(map[string]*ebiten.Image)
	rgbaByPath  = make(map[string]*image.RGBA)

	projectileSpriteByWeapon = make(map[string]*render.ProjectileSprite)
)

func getRGBAFromFile(texFile string) *image.RGBA {
	var rgba *image.RGBA
	resourcePath := filepath.Join("game", "resources", "textures")
	texFilePath := filepath.Join(resourcePath, texFile)
	if rgba, ok := rgbaByPath[texFilePath]; ok {
		return rgba
	}

	_, tex, err := ebitenutil.NewImageFromFile(texFilePath)
	if err != nil {
		log.Fatal(err)
	}
	if tex != nil {
		rgba = image.NewRGBA(image.Rect(0, 0, texWidth, texWidth))
		// convert into RGBA format
		for x := 0; x < texWidth; x++ {
			for y := 0; y < texWidth; y++ {
				clr := tex.At(x, y).(color.RGBA)
				rgba.SetRGBA(x, y, clr)
			}
		}
	}

	if rgba != nil {
		rgbaByPath[resourcePath] = rgba
	}

	return rgba
}

func getTextureFromFile(texFile string) *ebiten.Image {
	resourcePath := filepath.Join("game", "resources", "textures", texFile)
	if eImg, ok := imageByPath[resourcePath]; ok {
		return eImg
	}

	eImg, _, err := ebitenutil.NewImageFromFile(resourcePath)
	if err != nil {
		log.Fatal(err)
	}
	if eImg != nil {
		imageByPath[resourcePath] = eImg
	}
	return eImg
}

func getSpriteFromFile(sFile string) *ebiten.Image {
	resourcePath := filepath.Join("game", "resources", "sprites", sFile)
	if eImg, ok := imageByPath[resourcePath]; ok {
		return eImg
	}

	eImg, _, err := ebitenutil.NewImageFromFile(resourcePath)
	if err != nil {
		log.Fatal(err)
	}
	if eImg != nil {
		imageByPath[resourcePath] = eImg
	}
	return eImg
}

// loadContent loads all map texture and static sprite resources
func (g *Game) loadContent() {
	g.sprites = NewSpriteHandler()

	// keep a map of textures by name to only load duplicate entries once
	g.tex.texMap = make(map[string]*ebiten.Image, 128)

	// load textured flooring
	if g.mission.Map().Flooring.Default != "" {
		g.tex.floorTexDefault = newFloorTexture(g.mission.Map().Flooring.Default)
	}

	// keep track of floor texture positions by name so they can be matched on later
	var floorTexNames [][]string

	// load texture floor pathing
	if len(g.mission.Map().Flooring.Pathing) > 0 {
		g.tex.floorTexMap = make([][]*FloorTexture, g.mapWidth)
		floorTexNames = make([][]string, g.mapWidth)
		for x := 0; x < g.mapWidth; x++ {
			g.tex.floorTexMap[x] = make([]*FloorTexture, g.mapHeight)
			floorTexNames[x] = make([]string, g.mapHeight)
		}
		// create map grid of path image textures for the X/Y coords indicated
		for _, pathing := range g.mission.Map().Flooring.Pathing {
			tex := newFloorTexture(pathing.Image)

			// create filled rectangle paths
			for _, rect := range pathing.Rects {
				x0, y0, x1, y1 := rect[0][0], rect[0][1], rect[1][0], rect[1][1]
				for x := x0; x <= x1; x++ {
					for y := y0; y <= y1; y++ {
						g.tex.floorTexMap[x][y] = tex
						floorTexNames[x][y] = pathing.Image
					}
				}
			}

			// create line segment paths
			for _, segments := range pathing.Lines {
				var prevPoint *geom.Vector2
				for _, seg := range segments {
					point := &geom.Vector2{X: float64(seg[0]), Y: float64(seg[1])}

					if prevPoint != nil {
						// fill in path for line segment from previous to next point
						line := geom.Line{X1: prevPoint.X, Y1: prevPoint.Y, X2: point.X, Y2: point.Y}

						// use the angle of the line to then find every coordinate along the line path
						angle := line.Angle()
						dist := geom.Distance(line.X1, line.Y1, line.X2, line.Y2)
						for d := 0.0; d <= dist; d += 0.1 {
							nLine := geom.LineFromAngle(line.X1, line.Y1, angle, d)
							g.tex.floorTexMap[int(nLine.X2)][int(nLine.Y2)] = tex
						}
					}

					prevPoint = point
				}
			}
		}
	}

	// load clutter sprites mapped by path
	if len(g.mission.Map().Clutter) > 0 {
		g.clutter = NewClutterHandler()

		for _, clutter := range g.mission.Map().Clutter {
			var clutterImg *ebiten.Image
			if _, ok := g.tex.texMap[clutter.Image]; !ok {
				clutterImg = getSpriteFromFile(clutter.Image)
				g.tex.texMap[clutter.Image] = clutterImg
			}
		}
	}

	// load textures mapped by path
	for _, tex := range g.mission.Map().Textures {
		if tex.Image != "" {
			if _, ok := g.tex.texMap[tex.Image]; !ok {
				g.tex.texMap[tex.Image] = getTextureFromFile(tex.Image)
			}
		}

		if tex.SideX != "" {
			if _, ok := g.tex.texMap[tex.SideX]; !ok {
				g.tex.texMap[tex.SideX] = getTextureFromFile(tex.SideX)
			}
		}

		if tex.SideY != "" {
			if _, ok := g.tex.texMap[tex.SideY]; !ok {
				g.tex.texMap[tex.SideY] = getTextureFromFile(tex.SideY)
			}
		}
	}

	// load static sprites
	for _, s := range g.mission.Map().Sprites {
		if len(s.Image) == 0 {
			continue
		}

		if s.Scale == 0.0 {
			// default unset scale to 1.0
			s.Scale = 1.0
		}

		var spriteImg *ebiten.Image
		if eImg, ok := g.tex.texMap[s.Image]; ok {
			spriteImg = eImg
		} else {
			spriteImg = getSpriteFromFile(s.Image)
			g.tex.texMap[s.Image] = spriteImg
		}

		for _, position := range s.Positions {
			// convert collisionRadius/height pixel values to grid format
			sWidth, sHeight := spriteImg.Size()
			x, y, z := position[0], position[1], s.ZPosition

			collisionRadius, collisionHeight := convertCollisionFromPx(
				s.CollisionPxRadius, s.CollisionPxHeight, sWidth, sHeight, s.Scale,
			)

			hitPoints := math.MaxFloat64
			if s.HitPoints != 0 {
				hitPoints = s.HitPoints
			}

			sprite := render.NewSprite(
				model.BasicCollisionEntity(x, y, z, s.Anchor.SpriteAnchor, collisionRadius, collisionHeight, hitPoints),
				s.Scale, spriteImg, color.RGBA{0, 255, 0, 196},
			)

			g.sprites.addMapSprite(sprite)
		}
	}

	// load non-static mission sprites
	g.loadMissionSprites()

	// load all other game sprites
	g.loadGameSprites()
}

// loadMissionSprites loads all mission sprite reources
func (g *Game) loadMissionSprites() {
	mechSpriteTemplates := make(map[string]*render.MechSprite, len(g.mission.Mechs))
	vehicleSpriteTemplates := make(map[string]*render.VehicleSprite, len(g.mission.Vehicles))
	vtolSpriteTemplates := make(map[string]*render.VTOLSprite, len(g.mission.VTOLs))
	infantrySpriteTemplates := make(map[string]*render.InfantrySprite, len(g.mission.Infantry))

	for _, missionMech := range g.mission.Mechs {
		if _, ok := mechSpriteTemplates[missionMech.Unit]; !ok {
			modelMech := g.createModelMech(missionMech.Unit)

			mechResource := g.resources.GetMechResource(missionMech.Unit)
			mechRelPath := fmt.Sprintf("%s/%s", model.MechResourceType, mechResource.Image)
			mechImg := getSpriteFromFile(mechRelPath)

			mechSpriteTemplates[missionMech.Unit] = render.NewMechSprite(modelMech, mechResource.Scale, mechImg)
		}

		mechTemplate := mechSpriteTemplates[missionMech.Unit]
		mech := mechTemplate.Clone()

		posX, posY := missionMech.Position[0], missionMech.Position[1]
		mech.SetPos(&geom.Vector2{X: posX, Y: posY})

		// TODO: give mission units a bit more of a brain
		if len(missionMech.PatrolPath) > 0 {
			mech.PatrolPath = missionMech.PatrolPath
			mech.SetMechAnimation(render.ANIMATE_STRUT)
			mech.AnimationRate = 3
		} else {
			mech.SetMechAnimation(render.ANIMATE_IDLE)
			mech.AnimationRate = 7
		}

		g.sprites.addMechSprite(mech)
	}

	for _, missionVehicle := range g.mission.Vehicles {
		if _, ok := vehicleSpriteTemplates[missionVehicle.Unit]; !ok {
			vehicleResource := g.resources.GetVehicleResource(missionVehicle.Unit)
			vehicleRelPath := fmt.Sprintf("vehicles/%s", vehicleResource.Image)
			vehicleImg := getSpriteFromFile(vehicleRelPath)

			// need to use the image size to find the unit collision conversion from pixels
			width, height := vehicleImg.Size()
			// handle if image has multiple rows/cols
			if vehicleResource.ImageSheet != nil {
				width = int(float64(width) / float64(vehicleResource.ImageSheet.Columns))
				height = int(float64(width) / float64(vehicleResource.ImageSheet.Rows))
			}

			collisionRadius, collisionHeight := convertCollisionFromPx(
				vehicleResource.CollisionPxRadius, vehicleResource.CollisionPxHeight, width, height, vehicleResource.Scale,
			)

			modelVehicle := model.NewVehicle(vehicleResource, collisionRadius, collisionHeight)
			vehicleSpriteTemplates[missionVehicle.Unit] = render.NewVehicleSprite(modelVehicle, vehicleResource.Scale, vehicleImg)
		}

		vehicleTemplate := vehicleSpriteTemplates[missionVehicle.Unit]
		vehicle := vehicleTemplate.Clone()

		posX, posY := missionVehicle.Position[0], missionVehicle.Position[1]
		vehicle.SetPos(&geom.Vector2{X: posX, Y: posY})

		if len(missionVehicle.PatrolPath) > 0 {
			vehicle.PatrolPath = missionVehicle.PatrolPath
		}

		g.sprites.addVehicleSprite(vehicle)
	}

	for _, missionVTOL := range g.mission.VTOLs {
		if _, ok := vtolSpriteTemplates[missionVTOL.Unit]; !ok {
			vtolResource := g.resources.GetVTOLResource(missionVTOL.Unit)
			vtolRelPath := fmt.Sprintf("vtols/%s", vtolResource.Image)
			vtolImg := getSpriteFromFile(vtolRelPath)

			// need to use the image size to find the unit collision conversion from pixels
			width, height := vtolImg.Size()
			// handle if image has multiple rows/cols
			if vtolResource.ImageSheet != nil {
				width = int(float64(width) / float64(vtolResource.ImageSheet.Columns))
				height = int(float64(width) / float64(vtolResource.ImageSheet.Rows))
			}

			collisionRadius, collisionHeight := convertCollisionFromPx(
				vtolResource.CollisionPxRadius, vtolResource.CollisionPxHeight, width, height, vtolResource.Scale,
			)

			modelVTOL := model.NewVTOL(vtolResource, collisionRadius, collisionHeight)
			vtolSpriteTemplates[missionVTOL.Unit] = render.NewVTOLSprite(modelVTOL, vtolResource.Scale, vtolImg)
		}

		vtolTemplate := vtolSpriteTemplates[missionVTOL.Unit]
		vtol := vtolTemplate.Clone()

		posX, posY, posZ := missionVTOL.Position[0], missionVTOL.Position[1], missionVTOL.ZPosition
		vtol.SetPos(&geom.Vector2{X: posX, Y: posY})
		vtol.SetPosZ(posZ)

		if len(missionVTOL.PatrolPath) > 0 {
			vtol.PatrolPath = missionVTOL.PatrolPath
		}

		g.sprites.addVTOLSprite(vtol)
	}

	for _, missionInfantry := range g.mission.Infantry {
		if _, ok := infantrySpriteTemplates[missionInfantry.Unit]; !ok {
			infantryResource := g.resources.GetInfantryResource(missionInfantry.Unit)
			infantryRelPath := fmt.Sprintf("infantry/%s", infantryResource.Image)
			infantryImg := getSpriteFromFile(infantryRelPath)

			// need to use the image size to find the unit collision conversion from pixels
			width, height := infantryImg.Size()
			// handle if image has multiple rows/cols
			if infantryResource.ImageSheet != nil {
				width = int(float64(width) / float64(infantryResource.ImageSheet.Columns))
				height = int(float64(width) / float64(infantryResource.ImageSheet.Rows))
			}

			collisionRadius, collisionHeight := convertCollisionFromPx(
				infantryResource.CollisionPxRadius, infantryResource.CollisionPxHeight, width, height, infantryResource.Scale,
			)

			modelInfantry := model.NewInfantry(infantryResource, collisionRadius, collisionHeight)
			infantrySpriteTemplates[missionInfantry.Unit] = render.NewInfantrySprite(modelInfantry, infantryResource.Scale, infantryImg)
		}

		infantryTemplate := infantrySpriteTemplates[missionInfantry.Unit]
		infantry := infantryTemplate.Clone()

		posX, posY := missionInfantry.Position[0], missionInfantry.Position[1]
		infantry.SetPos(&geom.Vector2{X: posX, Y: posY})

		if len(missionInfantry.PatrolPath) > 0 {
			infantry.PatrolPath = missionInfantry.PatrolPath
		}

		g.sprites.addInfantrySprite(infantry)
	}
}

// loadGameSprites loads all other game sprite reources
func (g *Game) loadGameSprites() {
	// load crosshairs
	crosshairsSheet := getSpriteFromFile("hud/crosshairs_sheet.png")
	g.crosshairs = render.NewCrosshairs(crosshairsSheet, 1.0, 20, 10, 190)

	reticleSheet := getSpriteFromFile("hud/target_reticle.png")
	g.reticle = render.NewTargetReticle(1.0, reticleSheet)
}

func (g *Game) createModelMech(unit string) *model.Mech {
	mechResource := g.resources.GetMechResource(unit)
	mechRelPath := fmt.Sprintf("%s/%s", model.MechResourceType, mechResource.Image)
	mechImg := getSpriteFromFile(mechRelPath)

	// need to use the image size to find the unit collision conversion from pixels
	width, height := mechImg.Size()
	width = width / 6 // all mech images are required to be six columns of images in a sheet
	collisionRadius, collisionHeight := convertCollisionFromPx(
		mechResource.CollisionPxRadius, mechResource.CollisionPxHeight, width, height, mechResource.Scale,
	)

	modelMech := model.NewMech(mechResource, collisionRadius, collisionHeight)

	for _, armament := range mechResource.Armament {
		var weapon model.Weapon
		var projectile model.Projectile

		switch armament.Type.WeaponType {
		case model.ENERGY:
			weaponResource := g.resources.GetEnergyWeaponResource(armament.Weapon)
			weaponOffset := &geom.Vector2{X: armament.Offset[0], Y: armament.Offset[1]}

			// need to use the projectile image size to find the unit collision conversion from pixels
			pResource := weaponResource.Projectile
			projectileRelPath := fmt.Sprintf("%s/%s", model.ProjectilesResourceType, pResource.Image)
			projectileImg := getSpriteFromFile(projectileRelPath)
			pColumns, pRows := 1, 1
			if pResource.ImageSheet != nil {
				pColumns = pResource.ImageSheet.Columns
				pRows = pResource.ImageSheet.Rows
			}

			pWidth, pHeight := projectileImg.Size()
			pWidth = pWidth / pColumns
			pHeight = pHeight / pRows
			pCollisionRadius, pCollisionHeight := convertCollisionFromPx(
				pResource.CollisionPxRadius, pResource.CollisionPxHeight, pWidth, pHeight, pResource.Scale,
			)

			// create the weapon and projectile model
			weapon, projectile = model.NewEnergyWeapon(weaponResource, pCollisionRadius, pCollisionHeight, weaponOffset, modelMech)

			// create the projectile and effect sprite templates

			eResource := weaponResource.Projectile.ImpactEffect
			effectRelPath := fmt.Sprintf("%s/%s", model.EffectsResourceType, eResource.Image)
			effectImg := getSpriteFromFile(effectRelPath)
			eColumns, eRows, eAnimationRate := 1, 1, 1
			if eResource.ImageSheet != nil {
				eColumns = eResource.ImageSheet.Columns
				eRows = eResource.ImageSheet.Rows
				eAnimationRate = eResource.ImageSheet.AnimationRate
			}

			// TODO: check for existing resource first
			eSprite := render.NewAnimatedEffect(eResource.Scale, effectImg, eColumns, eRows, eAnimationRate, 1)

			// TODO: check for existing resource first
			pSprite := render.NewAnimatedProjectile(
				&projectile, pResource.Scale, projectileImg, color.RGBA{}, *eSprite,
			)
			setProjectileSpriteForWeapon(weapon, pSprite)
		}

		if weapon != nil {
			modelMech.AddArmament(weapon)
		}
	}

	return modelMech
}

func projectileSpriteForWeapon(w model.Weapon) *render.ProjectileSprite {
	wKey := model.TechBaseString(w.Tech()) + "_" + w.Name()
	return projectileSpriteByWeapon[wKey]
}

func setProjectileSpriteForWeapon(w model.Weapon, p *render.ProjectileSprite) {
	wKey := model.TechBaseString(w.Tech()) + "_" + w.Name()
	projectileSpriteByWeapon[wKey] = p
}

func convertCollisionFromPx(collisionPxRadius, collisionPxHeight float64, width, height int, scale float64) (collisionRadius float64, collisionHeight float64) {
	collisionRadius = (scale * collisionPxRadius) / float64(width)
	collisionHeight = (scale * collisionPxHeight) / float64(height)
	return
}
