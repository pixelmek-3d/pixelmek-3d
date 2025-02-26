package game

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"path"
	"path/filepath"
	"strings"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

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
	texFilePath := path.Join(resourcePath, texFile)
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
	resourcePath := path.Join("textures", texFile)
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
	resourcePath := path.Join("sprites", sFile)
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

		scale := s.Height / model.METERS_PER_UNIT

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
				s.CollisionPxRadius, s.CollisionPxHeight, sWidth, sHeight, scale,
			)

			hitPoints := math.MaxFloat64
			if s.HitPoints != 0 {
				hitPoints = s.HitPoints
			}

			sprite := render.NewSprite(
				model.BasicCollisionEntity(x, y, z, s.Anchor.SpriteAnchor, collisionRadius, collisionHeight, hitPoints),
				scale,
				spriteImg,
			)

			g.sprites.addMapSprite(sprite)
		}
	}

	// generate nav point sprites
	g.loadNavSprites()

	// load non-static mission sprites
	g.loadMissionSprites()

	// load special effects
	g.loadSpecialEffects()

	// load HUD display elements
	g.loadHUD()
}

// loadNavSprites generates nav point sprites
func (g *Game) loadNavSprites() {
	if g.mission == nil {
		return
	}

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
	if g.mission == nil {
		return
	}

	for _, missionMech := range g.mission.Mechs {
		modelMech, err := createMissionUnitModel[model.Mech](g, missionMech)
		if err != nil {
			log.Errorf("error creating mission mech: %v", err)
			continue
		}
		mech := g.createUnitSprite(modelMech).(*render.MechSprite)
		g.sprites.addMechSprite(mech)
	}

	for _, missionVehicle := range g.mission.Vehicles {
		modelVehicle, err := createMissionUnitModel[model.Vehicle](g, missionVehicle)
		if err != nil {
			log.Errorf("error creating mission vehicle: %v", err)
			continue
		}
		vehicle := g.createUnitSprite(modelVehicle).(*render.VehicleSprite)
		g.sprites.addVehicleSprite(vehicle)
	}

	for _, missionInfantry := range g.mission.Infantry {
		modelInfantry, err := createMissionUnitModel[model.Infantry](g, missionInfantry)
		if err != nil {
			log.Errorf("error creating mission infantry: %v", err)
			continue
		}
		infantry := g.createUnitSprite(modelInfantry).(*render.InfantrySprite)
		g.sprites.addInfantrySprite(infantry)
	}

	for _, missionVTOL := range g.mission.VTOLs {
		modelVTOL, err := createMissionFlyingUnitModel[model.VTOL](g, missionVTOL)
		if err != nil {
			log.Errorf("error creating mission VTOL: %v", err)
			continue
		}
		vtol := g.createUnitSprite(modelVTOL).(*render.VTOLSprite)
		g.sprites.addVTOLSprite(vtol)
	}

	for _, missionEmplacement := range g.mission.Emplacements {
		modelEmplacement, err := createMissionStaticUnitModel[model.Emplacement](g, missionEmplacement)
		if err != nil {
			log.Errorf("error creating mission emplacement: %v", err)
			continue
		}
		emplacement := g.createUnitSprite(modelEmplacement).(*render.EmplacementSprite)
		g.sprites.addEmplacementSprite(emplacement)
	}
}

func createMissionUnitModel[T model.MissionUnitModels](g *Game, unit model.MissionUnit) (model.Unit, error) {
	var u model.Unit
	mUnit := unit.Unit

	var t T
	switch any(t).(type) {
	case model.Mech:
		r, err := g.resources.GetMechResource(mUnit)
		if err != nil {
			return nil, err
		}
		u = g.createModelMechFromResource(r)
	case model.Vehicle:
		r, err := g.resources.GetVehicleResource(mUnit)
		if err != nil {
			return nil, err
		}
		u = g.createModelVehicleFromResource(r)
	case model.Infantry:
		r, err := g.resources.GetInfantryResource(mUnit)
		if err != nil {
			return nil, err
		}
		u = g.createModelInfantryFromResource(r)
	default:
		return nil, fmt.Errorf("model unit type not implemented: %T", t)
	}

	u.SetID(unit.ID)
	u.SetTeam(unit.Team)
	u.SetPos(&geom.Vector2{X: unit.Position[0], Y: unit.Position[1]})

	rHeading := geom.Radians(unit.Heading)
	u.SetHeading(rHeading)
	u.SetTargetHeading(rHeading)
	u.SetTurretAngle(rHeading)
	u.SetTargetTurretAngle(rHeading)

	u.SetGuardUnit(unit.GuardUnit)
	if len(unit.GuardArea.Position) == 2 && unit.GuardArea.Radius >= 0 &&
		unit.GuardArea.Position[0] > 0 && unit.GuardArea.Position[1] > 0 {

		u.SetGuardArea(unit.GuardArea.Position[0], unit.GuardArea.Position[1], unit.GuardArea.Radius)
		if len(unit.PatrolPath) > 0 {
			// Guard area is mutually exclusive from patrol path since it also uses the path stack
			return nil, fmt.Errorf("[%s] guard area and patrol path are mutually exclusive", u.ID())
		}
	} else {
		u.SetPatrolPath(model.PointsToVector2(unit.PatrolPath))
	}

	return u, nil
}

func createMissionFlyingUnitModel[T model.MissionFlyingUnitModels](g *Game, unit model.MissionFlyingUnit) (model.Unit, error) {
	var u model.Unit
	mUnit := unit.Unit

	var t T
	switch any(t).(type) {
	case model.VTOL:
		r, err := g.resources.GetVTOLResource(mUnit)
		if err != nil {
			return nil, err
		}
		u = g.createModelVTOLFromResource(r)
	default:
		return nil, fmt.Errorf("model flying unit type not implemented: %T", t)
	}

	u.SetID(unit.ID)
	u.SetTeam(unit.Team)
	u.SetPos(&geom.Vector2{X: unit.Position[0], Y: unit.Position[1]})
	u.SetPosZ(unit.ZPosition)

	rHeading := geom.Radians(unit.Heading)
	u.SetHeading(rHeading)
	u.SetTargetHeading(rHeading)
	u.SetTurretAngle(rHeading)
	u.SetTargetTurretAngle(rHeading)

	u.SetGuardUnit(unit.GuardUnit)
	if len(unit.GuardArea.Position) == 2 && unit.GuardArea.Radius >= 0 &&
		unit.GuardArea.Position[0] > 0 && unit.GuardArea.Position[1] > 0 {

		u.SetGuardArea(unit.GuardArea.Position[0], unit.GuardArea.Position[1], unit.GuardArea.Radius)
		if len(unit.PatrolPath) > 0 {
			// Guard area is mutually exclusive from patrol path since it also uses the path stack
			return nil, fmt.Errorf("[%s] guard area and patrol path are mutually exclusive", u.ID())
		}
	} else {
		u.SetPatrolPath(model.PointsToVector2(unit.PatrolPath))
	}

	return u, nil
}

func createMissionStaticUnitModel[T model.MissionStaticUnitModels](g *Game, unit model.MissionStaticUnit) (model.Unit, error) {
	var u model.Unit
	mUnit := unit.Unit

	var t T
	switch any(t).(type) {
	case model.Emplacement:
		r, err := g.resources.GetEmplacementResource(mUnit)
		if err != nil {
			return nil, err
		}
		u = g.createModelEmplacementFromResource(r)
	default:
		return nil, fmt.Errorf("model static unit type not implemented: %T", t)
	}

	u.SetID(unit.ID)
	u.SetTeam(unit.Team)
	u.SetPos(&geom.Vector2{X: unit.Position[0], Y: unit.Position[1]})

	rHeading := geom.Radians(unit.Heading)
	u.SetHeading(rHeading)
	u.SetTargetHeading(rHeading)
	u.SetTurretAngle(rHeading)
	u.SetTargetTurretAngle(rHeading)

	return u, nil
}

func (g *Game) createModelMechFromResource(mechResource *model.ModelMechResource) *model.Mech {
	mechRelPath := fmt.Sprintf("%s/%s", model.MechResourceType, mechResource.Image)
	mechImg := getSpriteFromFile(mechRelPath)

	// need to use the image size to find the unit collision conversion from pixels
	width, height := mechImg.Bounds().Dx(), mechImg.Bounds().Dy()
	width = width / 6 // all mech images are required to be six columns of images in a sheet
	scale := convertHeightToScale(mechResource.Height, height, mechResource.HeightPxGap)
	collisionRadius, collisionHeight := convertOffsetFromPx(
		mechResource.CollisionPxRadius, mechResource.CollisionPxHeight, width, height, scale,
	)

	cockpitPxX, cockpitPxY := mechResource.CockpitPxOffset[0], mechResource.CockpitPxOffset[1]
	cockpitOffX, cockPitOffY := convertOffsetFromPx(cockpitPxX, cockpitPxY, width, height, scale)

	modelMech := model.NewMech(mechResource, collisionRadius, collisionHeight, &geom.Vector2{X: cockpitOffX, Y: cockPitOffY})
	g.loadUnitWeapons(modelMech, mechResource.Armament, width, height, scale)
	g.loadUnitAmmo(modelMech, mechResource.Ammo)

	return modelMech
}

func (g *Game) createModelVehicleFromResource(vehicleResource *model.ModelVehicleResource) *model.Vehicle {
	vehicleRelPath := fmt.Sprintf("%s/%s", model.VehicleResourceType, vehicleResource.Image)
	vehicleImg := getSpriteFromFile(vehicleRelPath)

	// need to use the image size to find the unit collision conversion from pixels
	width, height := vehicleImg.Bounds().Dx(), vehicleImg.Bounds().Dy()
	// handle if image has multiple rows/cols
	if vehicleResource.ImageSheet != nil {
		width = int(float64(width) / float64(vehicleResource.ImageSheet.Columns))
		height = int(float64(height) / float64(vehicleResource.ImageSheet.Rows))
	}

	scale := convertHeightToScale(vehicleResource.Height, height, vehicleResource.HeightPxGap)
	collisionRadius, collisionHeight := convertOffsetFromPx(
		vehicleResource.CollisionPxRadius, vehicleResource.CollisionPxHeight, width, height, scale,
	)

	cockpitPxX, cockpitPxY := vehicleResource.CockpitPxOffset[0], vehicleResource.CockpitPxOffset[1]
	cockpitOffX, cockPitOffY := convertOffsetFromPx(cockpitPxX, cockpitPxY, width, height, scale)

	modelVehicle := model.NewVehicle(vehicleResource, collisionRadius, collisionHeight, &geom.Vector2{X: cockpitOffX, Y: cockPitOffY})
	g.loadUnitWeapons(modelVehicle, vehicleResource.Armament, width, height, scale)
	g.loadUnitAmmo(modelVehicle, vehicleResource.Ammo)

	return modelVehicle
}

func (g *Game) createModelVTOLFromResource(vtolResource *model.ModelVTOLResource) *model.VTOL {
	vtolRelPath := fmt.Sprintf("%s/%s", model.VTOLResourceType, vtolResource.Image)
	vtolImg := getSpriteFromFile(vtolRelPath)

	// need to use the image size to find the unit collision conversion from pixels
	width, height := vtolImg.Bounds().Dx(), vtolImg.Bounds().Dy()
	// handle if image has multiple rows/cols
	if vtolResource.ImageSheet != nil {
		width = int(float64(width) / float64(vtolResource.ImageSheet.Columns))
		height = int(float64(height) / float64(vtolResource.ImageSheet.Rows))
	}

	scale := convertHeightToScale(vtolResource.Height, height, vtolResource.HeightPxGap)
	collisionRadius, collisionHeight := convertOffsetFromPx(
		vtolResource.CollisionPxRadius, vtolResource.CollisionPxHeight, width, height, scale,
	)

	cockpitPxX, cockpitPxY := vtolResource.CockpitPxOffset[0], vtolResource.CockpitPxOffset[1]
	cockpitOffX, cockPitOffY := convertOffsetFromPx(cockpitPxX, cockpitPxY, width, height, scale)

	modelVTOL := model.NewVTOL(vtolResource, collisionRadius, collisionHeight, &geom.Vector2{X: cockpitOffX, Y: cockPitOffY})
	g.loadUnitWeapons(modelVTOL, vtolResource.Armament, width, height, scale)
	g.loadUnitAmmo(modelVTOL, vtolResource.Ammo)

	return modelVTOL
}

func (g *Game) createModelInfantryFromResource(infantryResource *model.ModelInfantryResource) *model.Infantry {
	infantryRelPath := fmt.Sprintf("%s/%s", model.InfantryResourceType, infantryResource.Image)
	infantryImg := getSpriteFromFile(infantryRelPath)

	// need to use the image size to find the unit collision conversion from pixels
	width, height := infantryImg.Bounds().Dx(), infantryImg.Bounds().Dy()
	// handle if image has multiple rows/cols
	if infantryResource.ImageSheet != nil {
		width = int(float64(width) / float64(infantryResource.ImageSheet.Columns))
		height = int(float64(height) / float64(infantryResource.ImageSheet.Rows))
	}

	scale := convertHeightToScale(infantryResource.Height, height, infantryResource.HeightPxGap)
	collisionRadius, collisionHeight := convertOffsetFromPx(
		infantryResource.CollisionPxRadius, infantryResource.CollisionPxHeight, width, height, scale,
	)

	cockpitPxX, cockpitPxY := infantryResource.CockpitPxOffset[0], infantryResource.CockpitPxOffset[1]
	cockpitOffX, cockPitOffY := convertOffsetFromPx(cockpitPxX, cockpitPxY, width, height, scale)

	modelInfantry := model.NewInfantry(infantryResource, collisionRadius, collisionHeight, &geom.Vector2{X: cockpitOffX, Y: cockPitOffY})
	g.loadUnitWeapons(modelInfantry, infantryResource.Armament, width, height, scale)
	g.loadUnitAmmo(modelInfantry, infantryResource.Ammo)

	return modelInfantry
}

func (g *Game) createModelEmplacementFromResource(emplacementResource *model.ModelEmplacementResource) *model.Emplacement {
	emplacementRelPath := fmt.Sprintf("%s/%s", model.EmplacementResourceType, emplacementResource.Image)
	emplacementImg := getSpriteFromFile(emplacementRelPath)

	// need to use the image size to find the unit collision conversion from pixels
	width, height := emplacementImg.Bounds().Dx(), emplacementImg.Bounds().Dy()
	// handle if image has multiple rows/cols
	if emplacementResource.ImageSheet != nil {
		width = int(float64(width) / float64(emplacementResource.ImageSheet.Columns))
		height = int(float64(height) / float64(emplacementResource.ImageSheet.Rows))
	}

	scale := convertHeightToScale(emplacementResource.Height, height, emplacementResource.HeightPxGap)
	collisionRadius, collisionHeight := convertOffsetFromPx(
		emplacementResource.CollisionPxRadius, emplacementResource.CollisionPxHeight, width, height, scale,
	)

	cockpitPxX, cockpitPxY := emplacementResource.CockpitPxOffset[0], emplacementResource.CockpitPxOffset[1]
	cockpitOffX, cockPitOffY := convertOffsetFromPx(cockpitPxX, cockpitPxY, width, height, scale)

	modelEmplacement := model.NewEmplacement(emplacementResource, collisionRadius, collisionHeight, &geom.Vector2{X: cockpitOffX, Y: cockPitOffY})
	g.loadUnitWeapons(modelEmplacement, emplacementResource.Armament, width, height, scale)
	g.loadUnitAmmo(modelEmplacement, emplacementResource.Ammo)

	return modelEmplacement
}

func (g *Game) loadUnitWeapons(unit model.Unit, armamentList []*model.ModelResourceArmament, unitWidthPx, unitHeightPx int, unitScale float64) {
	projectileSpriteTemplates := g.sprites.projectileSpriteTemplates

	for _, armament := range armamentList {
		var weapon model.Weapon
		var projectile model.Projectile

		switch armament.Type.WeaponType {
		case model.ENERGY:
			weaponResource, err := g.resources.GetEnergyWeaponResource(armament.Weapon)
			if err != nil {
				log.Errorf("[%s %s] weapon not found: %s", unit.Name(), unit.Variant(), err)
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

				projectileImpactAudioFiles := make([]string, 1)
				projectileImpactAudioFiles = append(projectileImpactAudioFiles, pResource.ImpactEffect.Audio)
				projectileImpactAudioFiles = append(projectileImpactAudioFiles, pResource.ImpactEffect.RandAudio...)

				eSpriteTemplate := render.NewAnimatedEffect(eResource, effectImg, 1)
				pSpriteTemplate := render.NewAnimatedProjectile(
					&projectile, pResource.Scale, projectileImg, *eSpriteTemplate, projectileImpactAudioFiles,
				)

				projectileSpriteTemplates[pTemplateKey] = pSpriteTemplate
			}

			pSpriteTemplate := projectileSpriteTemplates[pTemplateKey]
			pSprite := pSpriteTemplate.Clone()

			setProjectileSpriteForWeapon(weapon, pSprite)

		case model.MISSILE:
			weaponResource, err := g.resources.GetMissileWeaponResource(armament.Weapon)
			if err != nil {
				log.Errorf("[%s %s] weapon not found: %s", unit.Name(), unit.Variant(), err)
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

				projectileImpactAudioFiles := make([]string, 1)
				projectileImpactAudioFiles = append(projectileImpactAudioFiles, pResource.ImpactEffect.Audio)
				projectileImpactAudioFiles = append(projectileImpactAudioFiles, pResource.ImpactEffect.RandAudio...)

				eSpriteTemplate := render.NewAnimatedEffect(eResource, effectImg, 1)
				pSpriteTemplate := render.NewAnimatedProjectile(
					&projectile, pResource.Scale, projectileImg, *eSpriteTemplate, projectileImpactAudioFiles,
				)

				projectileSpriteTemplates[pTemplateKey] = pSpriteTemplate
			}

			pSpriteTemplate := projectileSpriteTemplates[pTemplateKey]
			pSprite := pSpriteTemplate.Clone()

			setProjectileSpriteForWeapon(weapon, pSprite)

		case model.BALLISTIC:
			weaponResource, err := g.resources.GetBallisticWeaponResource(armament.Weapon)
			if err != nil {
				log.Errorf("[%s %s] weapon not found: %s", unit.Name(), unit.Variant(), err)
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

				projectileImpactAudioFiles := make([]string, 1)
				projectileImpactAudioFiles = append(projectileImpactAudioFiles, pResource.ImpactEffect.Audio)
				projectileImpactAudioFiles = append(projectileImpactAudioFiles, pResource.ImpactEffect.RandAudio...)

				eSpriteTemplate := render.NewAnimatedEffect(eResource, effectImg, 1)
				pSpriteTemplate := render.NewAnimatedProjectile(
					&projectile, pResource.Scale, projectileImg, *eSpriteTemplate, projectileImpactAudioFiles,
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

func (g *Game) loadUnitAmmo(unit model.Unit, ammoList []*model.ModelResourceAmmo) {
	// load stock ammo
	ammo := unit.Ammunition()
	for _, ammoResource := range ammoList {
		ammoType := ammoResource.Type.AmmoType
		switch ammoType {
		case model.AMMO_BALLISTIC:
			var ammoBin *model.AmmoBin
			for _, w := range unit.Armament() {
				// ballistic ammo is specific for weapons of specified caliber
				if ballisticWeapon, ok := w.(*model.BallisticWeapon); ok {
					// trim file extension from weapon resource name so ammo for it can be listed without the extension
					weaponFileBase := strings.TrimSuffix(ballisticWeapon.File(), filepath.Ext(ballisticWeapon.File()))
					if ammoResource.ForWeapon != weaponFileBase {
						continue
					}
					if ammoBin == nil {
						ammoBin = ammo.AddAmmoBin(ammoType, ammoResource.Tons, w)
					}
					w.SetAmmoBin(ammoBin)
				}
			}
			if ammoBin == nil {
				log.Errorf(
					"no ballistic weapons (%s) found for ballistic ammo while initializing unit %s [%s]",
					ammoResource.ForWeapon,
					unit.Name(),
					unit.Variant(),
				)
			}
		case model.AMMO_LRM:
			// ammo is a pool for all LRM weapons, find a representative weapon for
			var ammoBin *model.AmmoBin
			for _, w := range unit.Armament() {
				if w.Classification() == model.MISSILE_LRM {
					if ammoBin == nil {
						ammoBin = ammo.AddAmmoBin(ammoType, ammoResource.Tons, w)
					}
					w.SetAmmoBin(ammoBin)
				}
			}
			if ammoBin == nil {
				log.Errorf(
					"no LRM weapons found for LRM ammo while initializing unit %s [%s]",
					unit.Name(),
					unit.Variant(),
				)
			}
		case model.AMMO_SRM:
			// ammo is a pool for all SRM weapons
			var ammoBin *model.AmmoBin
			for _, w := range unit.Armament() {
				if w.Classification() == model.MISSILE_SRM {
					if missileWeapon, ok := w.(*model.MissileWeapon); ok {
						// make sure it is not Streak SRM
						if missileWeapon.IsLockOnLockRequired() {
							continue
						}
						if ammoBin == nil {
							ammoBin = ammo.AddAmmoBin(ammoType, ammoResource.Tons, w)
						}
						w.SetAmmoBin(ammoBin)
					}
				}
			}
			if ammoBin == nil {
				log.Errorf(
					"no SRM weapons found for SRM ammo while initializing unit %s [%s]",
					unit.Name(),
					unit.Variant(),
				)
			}
		case model.AMMO_STREAK_SRM:
			// ammo is a pool for all Streak SRM weapons
			var ammoBin *model.AmmoBin
			for _, w := range unit.Armament() {
				if w.Classification() == model.MISSILE_SRM {
					if missileWeapon, ok := w.(*model.MissileWeapon); ok {
						// make sure it is Streak SRM
						if !missileWeapon.IsLockOnLockRequired() {
							continue
						}
						if ammoBin == nil {
							ammoBin = ammo.AddAmmoBin(ammoType, ammoResource.Tons, w)
						}
						w.SetAmmoBin(ammoBin)
					}
				}
			}
			if ammoBin == nil {
				log.Errorf(
					"no Streak SRM weapons found for Streak SRM ammo while initializing unit %s [%s]",
					unit.Name(),
					unit.Variant(),
				)
			}
		default:
			log.Errorf(
				"unhandled ammo type value '%v' while initializing unit %s [%s]",
				ammoResource.Type.AmmoType,
				unit.Name(),
				unit.Variant(),
			)
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

func convertHeightToScale(unitHeight float64, imageHeight, heightPxGap int) float64 {
	pxRatio := float64(imageHeight) / float64(imageHeight-heightPxGap)
	return pxRatio * unitHeight / model.METERS_PER_UNIT
}

func convertOffsetFromPx(xPx, yPx float64, width, height int, scaleY float64) (offX float64, offY float64) {
	// scale given based on height, calculate scale for width for images that are not 1:1
	scaleX := scaleY * float64(width) / float64(height)
	offX = (scaleX * xPx) / float64(width)
	offY = (scaleY * yPx) / float64(height)
	return
}
