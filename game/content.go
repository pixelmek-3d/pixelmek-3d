package game

import (
	"fmt"
	"image/color"
	"math"
	"path/filepath"
	"strings"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/sprites"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	"github.com/harbdog/raycaster-go/geom"

	log "github.com/sirupsen/logrus"
)

var (
	projectileSpriteByWeapon = make(map[string]*sprites.ProjectileSprite)
)

// loadContent loads all map texture and static sprite resources
func (g *Game) loadContent() {
	// initialize clutter handler
	g.clutter = NewClutterHandler()

	// load static sprites
	g.loadMapSprites()

	// generate nav point sprites
	g.loadNavSprites()

	// load non-static mission sprites
	g.loadMissionSprites()

	// load special effects
	g.loadSpecialEffects()

	// load HUD display elements
	g.loadHUD()
}

// loadMapSprites generates static map sprites
func (g *Game) loadMapSprites() {
	for _, s := range g.mission.Map().Sprites {
		if len(s.Image) == 0 {
			continue
		}

		scale := s.Height / model.METERS_PER_UNIT

		spriteImg := g.tex.TextureImage(s.Image)
		if spriteImg == nil {
			spriteImg = resources.GetSpriteFromFile(s.Image)
			g.tex.SetTextureImage(s.Image, spriteImg)
		}
		sWidth, sHeight := spriteImg.Bounds().Dx(), spriteImg.Bounds().Dy()

		for _, position := range s.Positions {
			// convert collisionRadius/height pixel values to grid format
			x, y, z := position[0], position[1], s.ZPosition

			collisionRadius, collisionHeight := convertOffsetFromPx(
				s.CollisionPxRadius, s.CollisionPxHeight, sWidth, sHeight, scale,
			)

			hitPoints := math.MaxFloat64
			if s.HitPoints != 0 {
				hitPoints = s.HitPoints
			}

			sprite := sprites.NewSprite(
				model.BasicCollisionEntity(x, y, z, s.Anchor.SpriteAnchor, collisionRadius, collisionHeight, hitPoints),
				scale,
				spriteImg,
			)

			g.sprites.AddMapSprite(sprite)
		}
	}
}

// loadNavSprites generates nav point sprites
func (g *Game) loadNavSprites() {
	if g.mission == nil {
		return
	}

	navSize := resources.TexSize / 2
	var nColor *color.NRGBA
	if g.hudUseCustomColor {
		nColor = g.hudRGBA
	}

	for _, navPoint := range g.mission.NavPoints {
		navImage := sprites.GenerateNavImage(navPoint, navSize, g.fonts.HUDFont, nColor)
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
		mech := g.createUnitSprite(modelMech).(*sprites.MechSprite)
		g.sprites.AddMechSprite(mech)
	}

	for _, missionVehicle := range g.mission.Vehicles {
		modelVehicle, err := createMissionUnitModel[model.Vehicle](g, missionVehicle)
		if err != nil {
			log.Errorf("error creating mission vehicle: %v", err)
			continue
		}
		vehicle := g.createUnitSprite(modelVehicle).(*sprites.VehicleSprite)
		g.sprites.AddVehicleSprite(vehicle)
	}

	for _, missionInfantry := range g.mission.Infantry {
		modelInfantry, err := createMissionUnitModel[model.Infantry](g, missionInfantry)
		if err != nil {
			log.Errorf("error creating mission infantry: %v", err)
			continue
		}
		infantry := g.createUnitSprite(modelInfantry).(*sprites.InfantrySprite)
		g.sprites.AddInfantrySprite(infantry)
	}

	for _, missionVTOL := range g.mission.VTOLs {
		modelVTOL, err := createMissionFlyingUnitModel[model.VTOL](g, missionVTOL)
		if err != nil {
			log.Errorf("error creating mission VTOL: %v", err)
			continue
		}
		vtol := g.createUnitSprite(modelVTOL).(*sprites.VTOLSprite)
		g.sprites.AddVTOLSprite(vtol)
	}

	for _, missionEmplacement := range g.mission.Emplacements {
		modelEmplacement, err := createMissionStaticUnitModel[model.Emplacement](g, missionEmplacement)
		if err != nil {
			log.Errorf("error creating mission emplacement: %v", err)
			continue
		}
		emplacement := g.createUnitSprite(modelEmplacement).(*sprites.EmplacementSprite)
		g.sprites.AddEmplacementSprite(emplacement)
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
	m := model.NewMech(mechResource)
	g.loadUnitWeapons(m, mechResource.Armament, m.PixelWidth(), m.PixelHeight(), m.PixelScale())
	g.loadUnitAmmo(m, mechResource.Ammo)
	return m
}

func (g *Game) createModelVehicleFromResource(vehicleResource *model.ModelVehicleResource) *model.Vehicle {
	m := model.NewVehicle(vehicleResource)
	g.loadUnitWeapons(m, vehicleResource.Armament, m.PixelWidth(), m.PixelHeight(), m.PixelScale())
	g.loadUnitAmmo(m, vehicleResource.Ammo)
	return m
}

func (g *Game) createModelVTOLFromResource(vtolResource *model.ModelVTOLResource) *model.VTOL {
	m := model.NewVTOL(vtolResource)
	g.loadUnitWeapons(m, vtolResource.Armament, m.PixelWidth(), m.PixelHeight(), m.PixelScale())
	g.loadUnitAmmo(m, vtolResource.Ammo)
	return m
}

func (g *Game) createModelInfantryFromResource(infantryResource *model.ModelInfantryResource) *model.Infantry {
	m := model.NewInfantry(infantryResource)
	g.loadUnitWeapons(m, infantryResource.Armament, m.PixelWidth(), m.PixelHeight(), m.PixelScale())
	g.loadUnitAmmo(m, infantryResource.Ammo)
	return m
}

func (g *Game) createModelEmplacementFromResource(emplacementResource *model.ModelEmplacementResource) *model.Emplacement {
	m := model.NewEmplacement(emplacementResource)
	g.loadUnitWeapons(m, emplacementResource.Armament, m.PixelWidth(), m.PixelHeight(), m.PixelScale())
	g.loadUnitAmmo(m, emplacementResource.Ammo)
	return m
}

func (g *Game) loadUnitWeapons(unit model.Unit, armamentList []*model.ModelResourceArmament, unitWidthPx, unitHeightPx int, unitScale float64) {
	// TODO: refactor to load weapons in model package
	projectileSpriteTemplates := g.sprites.ProjectileSpriteTemplates

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
			projectileImg := resources.GetSpriteFromFile(projectileRelPath)
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
				effectImg := resources.GetSpriteFromFile(effectRelPath)

				projectileImpactAudioFiles := make([]string, 1)
				projectileImpactAudioFiles = append(projectileImpactAudioFiles, pResource.ImpactEffect.Audio)
				projectileImpactAudioFiles = append(projectileImpactAudioFiles, pResource.ImpactEffect.RandAudio...)

				eSpriteTemplate := sprites.NewAnimatedEffect(eResource, effectImg, 1)
				pSpriteTemplate := sprites.NewAnimatedProjectile(
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
			projectileImg := resources.GetSpriteFromFile(projectileRelPath)
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
				effectImg := resources.GetSpriteFromFile(effectRelPath)

				projectileImpactAudioFiles := make([]string, 1)
				projectileImpactAudioFiles = append(projectileImpactAudioFiles, pResource.ImpactEffect.Audio)
				projectileImpactAudioFiles = append(projectileImpactAudioFiles, pResource.ImpactEffect.RandAudio...)

				eSpriteTemplate := sprites.NewAnimatedEffect(eResource, effectImg, 1)
				pSpriteTemplate := sprites.NewAnimatedProjectile(
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
			projectileImg := resources.GetSpriteFromFile(projectileRelPath)
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
				effectImg := resources.GetSpriteFromFile(effectRelPath)

				projectileImpactAudioFiles := make([]string, 1)
				projectileImpactAudioFiles = append(projectileImpactAudioFiles, pResource.ImpactEffect.Audio)
				projectileImpactAudioFiles = append(projectileImpactAudioFiles, pResource.ImpactEffect.RandAudio...)

				eSpriteTemplate := sprites.NewAnimatedEffect(eResource, effectImg, 1)
				pSpriteTemplate := sprites.NewAnimatedProjectile(
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
	// TODO: refactor to load ammo in model package
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

func projectileSpriteForWeapon(w model.Weapon) *sprites.ProjectileSprite {
	wKey := model.TechBaseString(w.Tech()) + "_" + w.Name()
	return projectileSpriteByWeapon[wKey]
}

func setProjectileSpriteForWeapon(w model.Weapon, p *sprites.ProjectileSprite) {
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
