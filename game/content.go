package game

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"path/filepath"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"
	"github.com/harbdog/pixelmek-3d/game/resources"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"

	log "github.com/sirupsen/logrus"
)

var (
	imageByPath = make(map[string]*ebiten.Image)
	rgbaByPath  = make(map[string]*image.RGBA)

	projectileSpriteByWeapon = make(map[string]*render.ProjectileSprite)
)

func getRGBAFromFile(texFile string) *image.RGBA {
	var rgba *image.RGBA
	resourcePath := "textures"
	texFilePath := filepath.Join(resourcePath, texFile)
	if rgba, ok := rgbaByPath[texFilePath]; ok {
		return rgba
	}

	_, tex, err := resources.NewImageFromFile(texFilePath)
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
	resourcePath := filepath.Join("textures", texFile)
	if eImg, ok := imageByPath[resourcePath]; ok {
		return eImg
	}

	eImg, _, err := resources.NewImageFromFile(resourcePath)
	if err != nil {
		log.Fatal(err)
	}
	if eImg != nil {
		imageByPath[resourcePath] = eImg
	}
	return eImg
}

func getSpriteFromFile(sFile string) *ebiten.Image {
	resourcePath := filepath.Join("sprites", sFile)
	if eImg, ok := imageByPath[resourcePath]; ok {
		return eImg
	}

	eImg, _, err := resources.NewImageFromFile(resourcePath)
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
	g.clutter = NewClutterHandler()
	if len(g.mission.Map().Clutter) > 0 {
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
			sWidth, sHeight := spriteImg.Bounds().Dx(), spriteImg.Bounds().Dy()
			x, y, z := position[0], position[1], s.ZPosition

			collisionRadius, collisionHeight := convertOffsetFromPx(
				s.CollisionPxRadius, s.CollisionPxHeight, sWidth, sHeight, s.Scale,
			)

			hitPoints := math.MaxFloat64
			if s.HitPoints != 0 {
				hitPoints = s.HitPoints
			}

			sprite := render.NewSprite(
				model.BasicCollisionEntity(x, y, z, s.Anchor.SpriteAnchor, collisionRadius, collisionHeight, hitPoints),
				s.Scale, spriteImg,
			)

			g.sprites.addMapSprite(sprite)
		}
	}

	// generate nav point sprites
	g.loadNavSprites()

	// load non-static mission sprites
	g.loadMissionSprites()

	// load HUD display elements
	g.loadHUD()
}

// loadNavSprites generates nav point sprites
func (g *Game) loadNavSprites() {
	navSize := texWidth / 2
	var nColor *color.NRGBA
	if g.hudUseCustomColor {
		nColor = g.hudRGBA
	}

	for _, navPoint := range g.mission.NavPoints {
		navImage := render.GenerateNavImage(navPoint, navSize, g.fonts.HUDFont, nColor)
		navPoint.SetImage(navImage)
	}
}

// loadMissionSprites loads all mission sprite reources
func (g *Game) loadMissionSprites() {
	vehicleSpriteTemplates := g.sprites.vehicleSpriteTemplates
	vtolSpriteTemplates := g.sprites.vtolSpriteTemplates
	infantrySpriteTemplates := g.sprites.infantrySpriteTemplates
	emplacementSpriteTemplates := g.sprites.emplacementSpriteTemplates

	for _, missionMech := range g.mission.Mechs {
		modelMech := g.createModelMech(missionMech.Unit)
		mech := g.createUnitSprite(modelMech).(*render.MechSprite)

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
			modelVehicle := g.createModelVehicle(missionVehicle.Unit)

			vehicleResource := g.resources.GetVehicleResource(missionVehicle.Unit)
			vehicleRelPath := fmt.Sprintf("%s/%s", model.VehicleResourceType, vehicleResource.Image)
			vehicleImg := getSpriteFromFile(vehicleRelPath)

			scale := convertHeightToScale(vehicleResource.Height, vehicleResource.HeightPxRatio)
			vehicleSpriteTemplates[missionVehicle.Unit] = render.NewVehicleSprite(modelVehicle, scale, vehicleImg)
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
			modelVTOL := g.createModelVTOL(missionVTOL.Unit)

			vtolResource := g.resources.GetVTOLResource(missionVTOL.Unit)
			vtolRelPath := fmt.Sprintf("%s/%s", model.VTOLResourceType, vtolResource.Image)
			vtolImg := getSpriteFromFile(vtolRelPath)

			scale := convertHeightToScale(vtolResource.Height, vtolResource.HeightPxRatio)
			vtolSpriteTemplates[missionVTOL.Unit] = render.NewVTOLSprite(modelVTOL, scale, vtolImg)
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
			modelInfantry := g.createModelInfantry(missionInfantry.Unit)

			infantryResource := g.resources.GetInfantryResource(missionInfantry.Unit)
			infantryRelPath := fmt.Sprintf("%s/%s", model.InfantryResourceType, infantryResource.Image)
			infantryImg := getSpriteFromFile(infantryRelPath)

			scale := convertHeightToScale(infantryResource.Height, infantryResource.HeightPxRatio)
			infantrySpriteTemplates[missionInfantry.Unit] = render.NewInfantrySprite(modelInfantry, scale, infantryImg)
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

	for _, missionEmplacement := range g.mission.Emplacements {
		if _, ok := emplacementSpriteTemplates[missionEmplacement.Unit]; !ok {
			modelEmplacement := g.createModelEmplacement(missionEmplacement.Unit)

			emplacementResource := g.resources.GetEmplacementResource(missionEmplacement.Unit)
			emplacementRelPath := fmt.Sprintf("%s/%s", model.EmplacementResourceType, emplacementResource.Image)
			emplacementImg := getSpriteFromFile(emplacementRelPath)

			scale := convertHeightToScale(emplacementResource.Height, emplacementResource.HeightPxRatio)
			emplacementSpriteTemplates[missionEmplacement.Unit] = render.NewEmplacementSprite(modelEmplacement, scale, emplacementImg)
		}

		emplacementTemplate := emplacementSpriteTemplates[missionEmplacement.Unit]
		emplacement := emplacementTemplate.Clone()

		posX, posY := missionEmplacement.Position[0], missionEmplacement.Position[1]
		emplacement.SetPos(&geom.Vector2{X: posX, Y: posY})

		g.sprites.addEmplacementSprite(emplacement)
	}
}

func (g *Game) createModelMech(unit string) *model.Mech {
	mechResource := g.resources.GetMechResource(unit)
	return g.createModelMechFromResource(mechResource)
}

func (g *Game) createModelMechFromResource(mechResource *model.ModelMechResource) *model.Mech {
	mechRelPath := fmt.Sprintf("%s/%s", model.MechResourceType, mechResource.Image)
	mechImg := getSpriteFromFile(mechRelPath)

	// need to use the image size to find the unit collision conversion from pixels
	width, height := mechImg.Bounds().Dx(), mechImg.Bounds().Dy()
	width = width / 6 // all mech images are required to be six columns of images in a sheet
	scale := convertHeightToScale(mechResource.Height, mechResource.HeightPxRatio)
	collisionRadius, collisionHeight := convertOffsetFromPx(
		mechResource.CollisionPxRadius, mechResource.CollisionPxHeight, width, height, scale,
	)

	cockpitPxX, cockpitPxY := mechResource.CockpitPxOffset[0], mechResource.CockpitPxOffset[1]
	cockpitOffX, cockPitOffY := convertOffsetFromPx(cockpitPxX, cockpitPxY, width, height, scale)

	modelMech := model.NewMech(mechResource, collisionRadius, collisionHeight, &geom.Vector2{X: cockpitOffX, Y: cockPitOffY})
	g.loadUnitWeapons(modelMech, mechResource.Armament, width, height, scale)

	return modelMech
}

func (g *Game) createModelVehicle(unit string) *model.Vehicle {
	vehicleResource := g.resources.GetVehicleResource(unit)
	vehicleRelPath := fmt.Sprintf("%s/%s", model.VehicleResourceType, vehicleResource.Image)
	vehicleImg := getSpriteFromFile(vehicleRelPath)

	// need to use the image size to find the unit collision conversion from pixels
	width, height := vehicleImg.Bounds().Dx(), vehicleImg.Bounds().Dy()
	// handle if image has multiple rows/cols
	if vehicleResource.ImageSheet != nil {
		width = int(float64(width) / float64(vehicleResource.ImageSheet.Columns))
		height = int(float64(height) / float64(vehicleResource.ImageSheet.Rows))
	}

	scale := convertHeightToScale(vehicleResource.Height, vehicleResource.HeightPxRatio)
	collisionRadius, collisionHeight := convertOffsetFromPx(
		vehicleResource.CollisionPxRadius, vehicleResource.CollisionPxHeight, width, height, scale,
	)

	cockpitPxX, cockpitPxY := vehicleResource.CockpitPxOffset[0], vehicleResource.CockpitPxOffset[1]
	cockpitOffX, cockPitOffY := convertOffsetFromPx(cockpitPxX, cockpitPxY, width, height, scale)

	modelVehicle := model.NewVehicle(vehicleResource, collisionRadius, collisionHeight, &geom.Vector2{X: cockpitOffX, Y: cockPitOffY})
	g.loadUnitWeapons(modelVehicle, vehicleResource.Armament, width, height, scale)

	return modelVehicle
}

func (g *Game) createModelVTOL(unit string) *model.VTOL {
	vtolResource := g.resources.GetVTOLResource(unit)
	vtolRelPath := fmt.Sprintf("%s/%s", model.VTOLResourceType, vtolResource.Image)
	vtolImg := getSpriteFromFile(vtolRelPath)

	// need to use the image size to find the unit collision conversion from pixels
	width, height := vtolImg.Bounds().Dx(), vtolImg.Bounds().Dy()
	// handle if image has multiple rows/cols
	if vtolResource.ImageSheet != nil {
		width = int(float64(width) / float64(vtolResource.ImageSheet.Columns))
		height = int(float64(height) / float64(vtolResource.ImageSheet.Rows))
	}

	scale := convertHeightToScale(vtolResource.Height, vtolResource.HeightPxRatio)
	collisionRadius, collisionHeight := convertOffsetFromPx(
		vtolResource.CollisionPxRadius, vtolResource.CollisionPxHeight, width, height, scale,
	)

	cockpitPxX, cockpitPxY := vtolResource.CockpitPxOffset[0], vtolResource.CockpitPxOffset[1]
	cockpitOffX, cockPitOffY := convertOffsetFromPx(cockpitPxX, cockpitPxY, width, height, scale)

	modelVTOL := model.NewVTOL(vtolResource, collisionRadius, collisionHeight, &geom.Vector2{X: cockpitOffX, Y: cockPitOffY})
	g.loadUnitWeapons(modelVTOL, vtolResource.Armament, width, height, scale)

	return modelVTOL
}

func (g *Game) createModelInfantry(unit string) *model.Infantry {
	infantryResource := g.resources.GetInfantryResource(unit)
	infantryRelPath := fmt.Sprintf("%s/%s", model.InfantryResourceType, infantryResource.Image)
	infantryImg := getSpriteFromFile(infantryRelPath)

	// need to use the image size to find the unit collision conversion from pixels
	width, height := infantryImg.Bounds().Dx(), infantryImg.Bounds().Dy()
	// handle if image has multiple rows/cols
	if infantryResource.ImageSheet != nil {
		width = int(float64(width) / float64(infantryResource.ImageSheet.Columns))
		height = int(float64(height) / float64(infantryResource.ImageSheet.Rows))
	}

	scale := convertHeightToScale(infantryResource.Height, infantryResource.HeightPxRatio)
	collisionRadius, collisionHeight := convertOffsetFromPx(
		infantryResource.CollisionPxRadius, infantryResource.CollisionPxHeight, width, height, scale,
	)

	cockpitPxX, cockpitPxY := infantryResource.CockpitPxOffset[0], infantryResource.CockpitPxOffset[1]
	cockpitOffX, cockPitOffY := convertOffsetFromPx(cockpitPxX, cockpitPxY, width, height, scale)

	modelInfantry := model.NewInfantry(infantryResource, collisionRadius, collisionHeight, &geom.Vector2{X: cockpitOffX, Y: cockPitOffY})
	g.loadUnitWeapons(modelInfantry, infantryResource.Armament, width, height, scale)

	return modelInfantry
}

func (g *Game) createModelEmplacement(unit string) *model.Emplacement {
	emplacementResource := g.resources.GetEmplacementResource(unit)
	emplacementRelPath := fmt.Sprintf("%s/%s", model.EmplacementResourceType, emplacementResource.Image)
	emplacementImg := getSpriteFromFile(emplacementRelPath)

	// need to use the image size to find the unit collision conversion from pixels
	width, height := emplacementImg.Bounds().Dx(), emplacementImg.Bounds().Dy()
	// handle if image has multiple rows/cols
	if emplacementResource.ImageSheet != nil {
		width = int(float64(width) / float64(emplacementResource.ImageSheet.Columns))
		height = int(float64(height) / float64(emplacementResource.ImageSheet.Rows))
	}

	scale := convertHeightToScale(emplacementResource.Height, emplacementResource.HeightPxRatio)
	collisionRadius, collisionHeight := convertOffsetFromPx(
		emplacementResource.CollisionPxRadius, emplacementResource.CollisionPxHeight, width, height, scale,
	)

	cockpitPxX, cockpitPxY := emplacementResource.CockpitPxOffset[0], emplacementResource.CockpitPxOffset[1]
	cockpitOffX, cockPitOffY := convertOffsetFromPx(cockpitPxX, cockpitPxY, width, height, scale)

	modelEmplacement := model.NewEmplacement(emplacementResource, collisionRadius, collisionHeight, &geom.Vector2{X: cockpitOffX, Y: cockPitOffY})
	g.loadUnitWeapons(modelEmplacement, emplacementResource.Armament, width, height, scale)

	return modelEmplacement
}

func (g *Game) loadUnitWeapons(unit model.Unit, armamentList []*model.ModelResourceArmament, unitWidthPx, unitHeightPx int, unitScale float64) {
	projectileSpriteTemplates := g.sprites.projectileSpriteTemplates

	for _, armament := range armamentList {
		var weapon model.Weapon
		var projectile model.Projectile

		switch armament.Type.WeaponType {
		case model.ENERGY:
			weaponResource := g.resources.GetEnergyWeaponResource(armament.Weapon)
			if weaponResource == nil {
				log.Errorf("[%s %s] weapon not found: %s/%s", unit.Name(), unit.Variant(), model.EnergyResourceType, armament.Weapon)
				continue
			}

			weaponOffX, weaponOffY := convertOffsetFromPx(
				armament.Offset[0], armament.Offset[1], unitWidthPx, unitHeightPx, unitScale,
			)
			weaponOffset := &geom.Vector2{X: weaponOffX, Y: weaponOffY}

			// need to use the projectile image size to find the unit collision conversion from pixels
			pResource := weaponResource.Projectile
			projectileRelPath := fmt.Sprintf("%s/%s", model.ProjectilesResourceType, pResource.Image)
			projectileImg := getSpriteFromFile(projectileRelPath)
			pColumns, pRows := 1, 1
			if pResource.ImageSheet != nil {
				pColumns = pResource.ImageSheet.Columns
				pRows = pResource.ImageSheet.Rows
			}

			pWidth, pHeight := projectileImg.Bounds().Dx(), projectileImg.Bounds().Dy()
			pWidth = pWidth / pColumns
			pHeight = pHeight / pRows
			pCollisionRadius, pCollisionHeight := convertOffsetFromPx(
				pResource.CollisionPxRadius, pResource.CollisionPxHeight, pWidth, pHeight, pResource.Scale,
			)

			// create the weapon and projectile model
			weapon, projectile = model.NewEnergyWeapon(weaponResource, pCollisionRadius, pCollisionHeight, weaponOffset, unit)

			pTemplateKey := model.EnergyResourceType + "_" + armament.Weapon
			if _, ok := projectileSpriteTemplates[pTemplateKey]; !ok {
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

				eSpriteTemplate := render.NewAnimatedEffect(eResource.Scale, effectImg, eColumns, eRows, eAnimationRate, 1)
				pSpriteTemplate := render.NewAnimatedProjectile(
					&projectile, pResource.Scale, projectileImg, *eSpriteTemplate,
				)

				projectileSpriteTemplates[pTemplateKey] = pSpriteTemplate
			}

			pSpriteTemplate := projectileSpriteTemplates[pTemplateKey]
			pSprite := pSpriteTemplate.Clone()

			setProjectileSpriteForWeapon(weapon, pSprite)

		case model.MISSILE:
			weaponResource := g.resources.GetMissileWeaponResource(armament.Weapon)
			if weaponResource == nil {
				log.Errorf("[%s %s] weapon not found: %s/%s", unit.Name(), unit.Variant(), model.MissileResourceType, armament.Weapon)
				continue
			}

			weaponOffX, weaponOffY := convertOffsetFromPx(
				armament.Offset[0], armament.Offset[1], unitWidthPx, unitHeightPx, unitScale,
			)
			weaponOffset := &geom.Vector2{X: weaponOffX, Y: weaponOffY}

			// need to use the projectile image size to find the unit collision conversion from pixels
			pResource := weaponResource.Projectile
			projectileRelPath := fmt.Sprintf("%s/%s", model.ProjectilesResourceType, pResource.Image)
			projectileImg := getSpriteFromFile(projectileRelPath)
			pColumns, pRows := 1, 1
			if pResource.ImageSheet != nil {
				pColumns = pResource.ImageSheet.Columns
				pRows = pResource.ImageSheet.Rows
			}

			pWidth, pHeight := projectileImg.Bounds().Dx(), projectileImg.Bounds().Dy()
			pWidth = pWidth / pColumns
			pHeight = pHeight / pRows
			pCollisionRadius, pCollisionHeight := convertOffsetFromPx(
				pResource.CollisionPxRadius, pResource.CollisionPxHeight, pWidth, pHeight, pResource.Scale,
			)

			// missile tube generated visual needs single pixel offset
			onePxX, onePxY := convertOffsetFromPx(
				1, 1, pWidth, pHeight, pResource.Scale,
			)
			onePxOffset := &geom.Vector2{X: onePxX, Y: onePxY}

			// create the weapon and projectile model
			weapon, projectile = model.NewMissileWeapon(
				weaponResource, pCollisionRadius, pCollisionHeight, weaponOffset, onePxOffset, unit,
			)

			pTemplateKey := model.MissileResourceType + "_" + armament.Weapon
			if _, ok := projectileSpriteTemplates[pTemplateKey]; !ok {
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

				eSpriteTemplate := render.NewAnimatedEffect(eResource.Scale, effectImg, eColumns, eRows, eAnimationRate, 1)
				pSpriteTemplate := render.NewAnimatedProjectile(
					&projectile, pResource.Scale, projectileImg, *eSpriteTemplate,
				)

				projectileSpriteTemplates[pTemplateKey] = pSpriteTemplate
			}

			pSpriteTemplate := projectileSpriteTemplates[pTemplateKey]
			pSprite := pSpriteTemplate.Clone()

			setProjectileSpriteForWeapon(weapon, pSprite)

		case model.BALLISTIC:
			weaponResource := g.resources.GetBallisticWeaponResource(armament.Weapon)
			if weaponResource == nil {
				log.Errorf("[%s %s] weapon not found: %s/%s", unit.Name(), unit.Variant(), model.BallisticResourceType, armament.Weapon)
				continue
			}

			weaponOffX, weaponOffY := convertOffsetFromPx(
				armament.Offset[0], armament.Offset[1], unitWidthPx, unitHeightPx, unitScale,
			)
			weaponOffset := &geom.Vector2{X: weaponOffX, Y: weaponOffY}

			// need to use the projectile image size to find the unit collision conversion from pixels
			pResource := weaponResource.Projectile
			projectileRelPath := fmt.Sprintf("%s/%s", model.ProjectilesResourceType, pResource.Image)
			projectileImg := getSpriteFromFile(projectileRelPath)
			pColumns, pRows := 1, 1
			if pResource.ImageSheet != nil {
				pColumns = pResource.ImageSheet.Columns
				pRows = pResource.ImageSheet.Rows
			}

			pWidth, pHeight := projectileImg.Bounds().Dx(), projectileImg.Bounds().Dy()
			pWidth = pWidth / pColumns
			pHeight = pHeight / pRows
			pCollisionRadius, pCollisionHeight := convertOffsetFromPx(
				pResource.CollisionPxRadius, pResource.CollisionPxHeight, pWidth, pHeight, pResource.Scale,
			)

			// create the weapon and projectile model
			weapon, projectile = model.NewBallisticWeapon(weaponResource, pCollisionRadius, pCollisionHeight, weaponOffset, unit)

			pTemplateKey := model.BallisticResourceType + "_" + armament.Weapon
			if _, ok := projectileSpriteTemplates[pTemplateKey]; !ok {
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

				eSpriteTemplate := render.NewAnimatedEffect(eResource.Scale, effectImg, eColumns, eRows, eAnimationRate, 1)
				pSpriteTemplate := render.NewAnimatedProjectile(
					&projectile, pResource.Scale, projectileImg, *eSpriteTemplate,
				)

				projectileSpriteTemplates[pTemplateKey] = pSpriteTemplate
			}

			pSpriteTemplate := projectileSpriteTemplates[pTemplateKey]
			pSprite := pSpriteTemplate.Clone()

			setProjectileSpriteForWeapon(weapon, pSprite)
		}

		if weapon != nil {
			unit.AddArmament(weapon)
		}
	}
}

func projectileSpriteForWeapon(w model.Weapon) *render.ProjectileSprite {
	wKey := model.TechBaseString(w.Tech()) + "_" + w.Name()
	return projectileSpriteByWeapon[wKey]
}

func setProjectileSpriteForWeapon(w model.Weapon, p *render.ProjectileSprite) {
	wKey := model.TechBaseString(w.Tech()) + "_" + w.Name()
	projectileSpriteByWeapon[wKey] = p
}

func convertHeightToScale(height, pxRatio float64) float64 {
	if pxRatio == 0 {
		// if unset, default to 1.0
		pxRatio = 1
	}
	return pxRatio * height / model.METERS_PER_UNIT
}

func convertOffsetFromPx(xPx, yPx float64, width, height int, scale float64) (offX float64, offY float64) {
	offX = (scale * xPx) / float64(width)
	offY = (scale * yPx) / float64(height)
	return
}
