package game

import (
	"fmt"
	"math"
	"sync"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"

	"github.com/harbdog/raycaster-go"
)

type SpriteHandler struct {
	sprites map[SpriteType]*sync.Map

	mechSpriteTemplates        map[string]*render.MechSprite
	vehicleSpriteTemplates     map[string]*render.VehicleSprite
	vtolSpriteTemplates        map[string]*render.VTOLSprite
	infantrySpriteTemplates    map[string]*render.InfantrySprite
	emplacementSpriteTemplates map[string]*render.EmplacementSprite
	projectileSpriteTemplates  map[string]*render.ProjectileSprite
}

type SpriteType int

const (
	MapSpriteType SpriteType = iota
	MechSpriteType
	VehicleSpriteType
	VTOLSpriteType
	InfantrySpriteType
	EmplacementSpriteType
	ProjectileSpriteType
	EffectSpriteType
	TotalSpriteTypes
)

func NewSpriteHandler() *SpriteHandler {
	s := &SpriteHandler{
		sprites:                    make(map[SpriteType]*sync.Map, TotalSpriteTypes),
		mechSpriteTemplates:        make(map[string]*render.MechSprite),
		vehicleSpriteTemplates:     make(map[string]*render.VehicleSprite),
		vtolSpriteTemplates:        make(map[string]*render.VTOLSprite),
		infantrySpriteTemplates:    make(map[string]*render.InfantrySprite),
		emplacementSpriteTemplates: make(map[string]*render.EmplacementSprite),
		projectileSpriteTemplates:  make(map[string]*render.ProjectileSprite),
	}
	s.sprites[MechSpriteType] = &sync.Map{}
	s.sprites[VehicleSpriteType] = &sync.Map{}
	s.sprites[VTOLSpriteType] = &sync.Map{}
	s.sprites[InfantrySpriteType] = &sync.Map{}
	s.sprites[EmplacementSpriteType] = &sync.Map{}
	s.sprites[MapSpriteType] = &sync.Map{}
	s.sprites[ProjectileSpriteType] = &sync.Map{}
	s.sprites[EffectSpriteType] = &sync.Map{}

	return s
}

func (s *SpriteHandler) clear() {
	for spriteType := range s.sprites {
		s.sprites[spriteType] = &sync.Map{}
	}
}

func (s *SpriteHandler) addMapSprite(sprite *render.Sprite) {
	s.sprites[MapSpriteType].Store(sprite, struct{}{})
}

func (s *SpriteHandler) deleteMapSprite(sprite *render.Sprite) {
	s.sprites[MapSpriteType].Delete(sprite)
}

func (s *SpriteHandler) addMechSprite(mech *render.MechSprite) {
	s.sprites[MechSpriteType].Store(mech, struct{}{})
}

func (s *SpriteHandler) deleteMechSprite(mech *render.MechSprite) {
	s.sprites[MechSpriteType].Delete(mech)
}

func (s *SpriteHandler) addVehicleSprite(vehicle *render.VehicleSprite) {
	s.sprites[VehicleSpriteType].Store(vehicle, struct{}{})
}

func (s *SpriteHandler) deleteVehicleSprite(vehicle *render.VehicleSprite) {
	s.sprites[VehicleSpriteType].Delete(vehicle)
}

func (s *SpriteHandler) addVTOLSprite(vtol *render.VTOLSprite) {
	s.sprites[VTOLSpriteType].Store(vtol, struct{}{})
}

func (s *SpriteHandler) deleteVTOLSprite(vtol *render.VTOLSprite) {
	s.sprites[VTOLSpriteType].Delete(vtol)
}

func (s *SpriteHandler) addInfantrySprite(infantry *render.InfantrySprite) {
	s.sprites[InfantrySpriteType].Store(infantry, struct{}{})
}

func (s *SpriteHandler) deleteInfantrySprite(infantry *render.InfantrySprite) {
	s.sprites[InfantrySpriteType].Delete(infantry)
}

func (s *SpriteHandler) addEmplacementSprite(emplacement *render.EmplacementSprite) {
	s.sprites[EmplacementSpriteType].Store(emplacement, struct{}{})
}

func (s *SpriteHandler) deleteEmplacementSprite(emplacement *render.EmplacementSprite) {
	s.sprites[EmplacementSpriteType].Delete(emplacement)
}

func (s *SpriteHandler) addProjectile(projectile *render.ProjectileSprite) {
	s.sprites[ProjectileSpriteType].Store(projectile, struct{}{})
}

func (s *SpriteHandler) deleteProjectile(projectile *render.ProjectileSprite) {
	s.sprites[ProjectileSpriteType].Delete(projectile)
}

func (s *SpriteHandler) addEffect(effect *render.EffectSprite) {
	s.sprites[EffectSpriteType].Store(effect, struct{}{})
}

func (s *SpriteHandler) deleteEffect(effect *render.EffectSprite) {
	s.sprites[EffectSpriteType].Delete(effect)
}

func (g *Game) createUnitSprite(unit model.Unit) raycaster.Sprite {
	switch interfaceType := unit.(type) {
	case *model.Mech:
		u := unit.(*model.Mech)
		uKey := u.Resource.File
		unitSprite, found := g.sprites.mechSpriteTemplates[uKey]
		if !found {
			relPath := fmt.Sprintf("%s/%s", model.MechResourceType, u.Resource.Image)
			img := getSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = render.NewMechSprite(u, scale, img)
			g.sprites.mechSpriteTemplates[uKey] = unitSprite
		}
		return unitSprite.Clone(u)

	case *model.Vehicle:
		u := unit.(*model.Vehicle)
		uKey := u.Resource.File
		unitSprite, found := g.sprites.vehicleSpriteTemplates[uKey]
		if !found {
			relPath := fmt.Sprintf("%s/%s", model.VehicleResourceType, u.Resource.Image)
			img := getSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = render.NewVehicleSprite(u, scale, img)
			g.sprites.vehicleSpriteTemplates[uKey] = unitSprite
		}
		return unitSprite.Clone(u)

	case *model.VTOL:
		u := unit.(*model.VTOL)
		uKey := u.Resource.File
		unitSprite, found := g.sprites.vtolSpriteTemplates[uKey]
		if !found {
			relPath := fmt.Sprintf("%s/%s", model.VTOLResourceType, u.Resource.Image)
			img := getSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = render.NewVTOLSprite(u, scale, img)
			g.sprites.vtolSpriteTemplates[uKey] = unitSprite
		}
		return unitSprite.Clone(u)

	case *model.Infantry:
		u := unit.(*model.Infantry)
		uKey := u.Resource.File
		unitSprite, found := g.sprites.infantrySpriteTemplates[uKey]
		if !found {
			relPath := fmt.Sprintf("%s/%s", model.InfantryResourceType, u.Resource.Image)
			img := getSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = render.NewInfantrySprite(u, scale, img)
			g.sprites.infantrySpriteTemplates[uKey] = unitSprite
		}
		return unitSprite.Clone(u)

	case *model.Emplacement:
		u := unit.(*model.Emplacement)
		uKey := u.Resource.File
		unitSprite, found := g.sprites.emplacementSpriteTemplates[uKey]
		if !found {
			relPath := fmt.Sprintf("%s/%s", model.EmplacementResourceType, u.Resource.Image)
			img := getSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = render.NewEmplacementSprite(u, scale, img)
			g.sprites.emplacementSpriteTemplates[uKey] = unitSprite
		}
		return unitSprite.Clone(u)

	default:
		panic(fmt.Errorf("unable to handle model.Unit from type %v", interfaceType))
	}
}

func (g *Game) getRaycastSprites() []raycaster.Sprite {
	raycastSprites := make([]raycaster.Sprite, 0, 512)

	playerPos := g.player.Pos()

	count := 0
	for _, spriteMap := range g.sprites.sprites {
		spriteMap.Range(func(k, _ interface{}) bool {
			spriteInterface := k.(raycaster.Sprite)
			sprite := getSpriteFromInterface(spriteInterface)
			// for now this is sufficient, but for much larger amounts of sprites may need goroutines to divide up the work
			// only include map sprites within fast approximation of render distance
			doSprite := g.renderDistance < 0 || g.player.Target() == sprite.Entity ||
				(math.Abs(sprite.Pos().X-playerPos.X) <= g.renderDistance &&
					math.Abs(sprite.Pos().Y-playerPos.Y) <= g.renderDistance)
			if doSprite {
				raycastSprites = append(raycastSprites, sprite)
				count++
			}
			return true
		})
	}
	for clutter := range g.clutter.sprites {
		raycastSprites = append(raycastSprites, clutter)
		count++
	}

	// add the reticle lead indicator as sprite for raycast location
	if g.player.reticleLead != nil {
		raycastSprites = append(raycastSprites, g.player.reticleLead)
		count++
	}

	// add the currently selected nav point as sprite
	if g.player.currentNav != nil {
		raycastSprites = append(raycastSprites, g.player.currentNav)
		count++
	}

	return raycastSprites[:count]
}

func (g *Game) getSpriteUnits() []model.Unit {
	units := make([]model.Unit, 0, 64)
	for _, spriteMap := range g.sprites.sprites {
		spriteMap.Range(func(k, _ interface{}) bool {
			spriteInterface := k.(raycaster.Sprite)
			entity := getEntityFromInterface(spriteInterface)
			unit := model.EntityUnit(entity)
			if unit != nil {
				units = append(units, unit)
			}
			return true
		})
	}

	return units
}

func getSpriteType(sInterface raycaster.Sprite) SpriteType {
	switch interfaceType := sInterface.(type) {
	case *render.Sprite:
		return MapSpriteType
	case *render.MechSprite:
		return MechSpriteType
	case *render.VehicleSprite:
		return VehicleSpriteType
	case *render.VTOLSprite:
		return VTOLSpriteType
	case *render.InfantrySprite:
		return InfantrySpriteType
	case *render.EmplacementSprite:
		return EmplacementSpriteType
	case *render.ProjectileSprite:
		return ProjectileSpriteType
	case *render.EffectSprite:
		return EffectSpriteType
	default:
		panic(fmt.Errorf("unable to get SpriteType from sprite interface type %v", interfaceType))
	}
}

func getSpriteFromInterface(sInterface raycaster.Sprite) *render.Sprite {
	if sInterface == nil {
		return nil
	}

	sType := getSpriteType(sInterface)
	switch sType {
	case MapSpriteType:
		return sInterface.(*render.Sprite)
	case MechSpriteType:
		return sInterface.(*render.MechSprite).Sprite
	case VehicleSpriteType:
		return sInterface.(*render.VehicleSprite).Sprite
	case VTOLSpriteType:
		return sInterface.(*render.VTOLSprite).Sprite
	case InfantrySpriteType:
		return sInterface.(*render.InfantrySprite).Sprite
	case EmplacementSpriteType:
		return sInterface.(*render.EmplacementSprite).Sprite
	case ProjectileSpriteType:
		return sInterface.(*render.ProjectileSprite).Sprite
	case EffectSpriteType:
		return sInterface.(*render.EffectSprite).Sprite
	default:
		panic(fmt.Errorf("unable to get model.Sprite from type %v", sType))
	}
}

func getEntityFromInterface(sInterface raycaster.Sprite) model.Entity {
	sType := getSpriteType(sInterface)
	switch sType {
	case MapSpriteType:
		return sInterface.(*render.Sprite).Entity
	case MechSpriteType:
		return sInterface.(*render.MechSprite).Entity
	case VehicleSpriteType:
		return sInterface.(*render.VehicleSprite).Entity
	case VTOLSpriteType:
		return sInterface.(*render.VTOLSprite).Entity
	case InfantrySpriteType:
		return sInterface.(*render.InfantrySprite).Entity
	case EmplacementSpriteType:
		return sInterface.(*render.EmplacementSprite).Entity
	case ProjectileSpriteType:
		return sInterface.(*render.ProjectileSprite).Entity
	case EffectSpriteType:
		return sInterface.(*render.EffectSprite).Entity
	default:
		panic(fmt.Errorf("unable to get model.Entity from type %v", sType))
	}
}

func (g *Game) getSpriteFromEntity(entity model.Entity) *render.Sprite {
	var found *render.Sprite
	for spriteType := range g.sprites.sprites {
		g.sprites.sprites[spriteType].Range(func(k, _ interface{}) bool {
			if !g.isInteractiveType(spriteType) {
				// only show on certain sprite types (skip projectiles, effects, etc.)
				return true
			}

			s := getSpriteFromInterface(k.(raycaster.Sprite))
			if entity == s.Entity {
				found = s
				return false // found, stop Range iteration
			}

			return true
		})

		if found != nil {
			return found
		}
	}

	return nil
}

func (g *Game) getMapSpriteFromEntity(entity model.Entity) *render.Sprite {
	var found *render.Sprite

	g.sprites.sprites[MapSpriteType].Range(func(k, _ interface{}) bool {
		s := getSpriteFromInterface(k.(raycaster.Sprite))
		if entity == s.Entity {
			found = s
			return false // found, stop Range iteration
		}

		return true
	})

	if found != nil {
		return found
	}

	return nil
}
