package game

import (
	"fmt"
	"sort"
	"sync"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/sprites"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type SpriteHandler struct {
	sprites map[SpriteType]*sync.Map

	mechSpriteTemplates        map[string]*sprites.MechSprite
	vehicleSpriteTemplates     map[string]*sprites.VehicleSprite
	vtolSpriteTemplates        map[string]*sprites.VTOLSprite
	infantrySpriteTemplates    map[string]*sprites.InfantrySprite
	emplacementSpriteTemplates map[string]*sprites.EmplacementSprite
	projectileSpriteTemplates  map[string]*sprites.ProjectileSprite
}

type proximitySprite struct {
	sprite   *sprites.Sprite
	distance float64
}

type proximityUnit struct {
	unit     model.Unit
	distance float64
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
		mechSpriteTemplates:        make(map[string]*sprites.MechSprite),
		vehicleSpriteTemplates:     make(map[string]*sprites.VehicleSprite),
		vtolSpriteTemplates:        make(map[string]*sprites.VTOLSprite),
		infantrySpriteTemplates:    make(map[string]*sprites.InfantrySprite),
		emplacementSpriteTemplates: make(map[string]*sprites.EmplacementSprite),
		projectileSpriteTemplates:  make(map[string]*sprites.ProjectileSprite),
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

func (s *SpriteHandler) addMapSprite(sprite *sprites.Sprite) {
	s.sprites[MapSpriteType].Store(sprite, struct{}{})
}

func (s *SpriteHandler) deleteMapSprite(sprite *sprites.Sprite) {
	s.sprites[MapSpriteType].Delete(sprite)
}

func (s *SpriteHandler) addMechSprite(mech *sprites.MechSprite) {
	s.sprites[MechSpriteType].Store(mech, struct{}{})
}

func (s *SpriteHandler) deleteMechSprite(mech *sprites.MechSprite) {
	s.sprites[MechSpriteType].Delete(mech)
}

func (s *SpriteHandler) addVehicleSprite(vehicle *sprites.VehicleSprite) {
	s.sprites[VehicleSpriteType].Store(vehicle, struct{}{})
}

func (s *SpriteHandler) deleteVehicleSprite(vehicle *sprites.VehicleSprite) {
	s.sprites[VehicleSpriteType].Delete(vehicle)
}

func (s *SpriteHandler) addVTOLSprite(vtol *sprites.VTOLSprite) {
	s.sprites[VTOLSpriteType].Store(vtol, struct{}{})
}

func (s *SpriteHandler) deleteVTOLSprite(vtol *sprites.VTOLSprite) {
	s.sprites[VTOLSpriteType].Delete(vtol)
}

func (s *SpriteHandler) addInfantrySprite(infantry *sprites.InfantrySprite) {
	s.sprites[InfantrySpriteType].Store(infantry, struct{}{})
}

func (s *SpriteHandler) deleteInfantrySprite(infantry *sprites.InfantrySprite) {
	s.sprites[InfantrySpriteType].Delete(infantry)
}

func (s *SpriteHandler) addEmplacementSprite(emplacement *sprites.EmplacementSprite) {
	s.sprites[EmplacementSpriteType].Store(emplacement, struct{}{})
}

func (s *SpriteHandler) deleteEmplacementSprite(emplacement *sprites.EmplacementSprite) {
	s.sprites[EmplacementSpriteType].Delete(emplacement)
}

func (s *SpriteHandler) addProjectile(projectile *sprites.ProjectileSprite) {
	s.sprites[ProjectileSpriteType].Store(projectile, struct{}{})
}

func (s *SpriteHandler) deleteProjectile(projectile *sprites.ProjectileSprite) {
	s.sprites[ProjectileSpriteType].Delete(projectile)
}

func (s *SpriteHandler) addEffect(effect *sprites.EffectSprite) {
	s.sprites[EffectSpriteType].Store(effect, struct{}{})
}

func (s *SpriteHandler) deleteEffect(effect *sprites.EffectSprite) {
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
			img := resources.GetSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = sprites.NewMechSprite(u, scale, img)
			g.sprites.mechSpriteTemplates[uKey] = unitSprite
		}
		return unitSprite.Clone(u)

	case *model.Vehicle:
		u := unit.(*model.Vehicle)
		uKey := u.Resource.File
		unitSprite, found := g.sprites.vehicleSpriteTemplates[uKey]
		if !found {
			relPath := fmt.Sprintf("%s/%s", model.VehicleResourceType, u.Resource.Image)
			img := resources.GetSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = sprites.NewVehicleSprite(u, scale, img)
			g.sprites.vehicleSpriteTemplates[uKey] = unitSprite
		}
		return unitSprite.Clone(u)

	case *model.VTOL:
		u := unit.(*model.VTOL)
		uKey := u.Resource.File
		unitSprite, found := g.sprites.vtolSpriteTemplates[uKey]
		if !found {
			relPath := fmt.Sprintf("%s/%s", model.VTOLResourceType, u.Resource.Image)
			img := resources.GetSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = sprites.NewVTOLSprite(u, scale, img)
			g.sprites.vtolSpriteTemplates[uKey] = unitSprite
		}
		return unitSprite.Clone(u)

	case *model.Infantry:
		u := unit.(*model.Infantry)
		uKey := u.Resource.File
		unitSprite, found := g.sprites.infantrySpriteTemplates[uKey]
		if !found {
			relPath := fmt.Sprintf("%s/%s", model.InfantryResourceType, u.Resource.Image)
			img := resources.GetSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = sprites.NewInfantrySprite(u, scale, img)
			g.sprites.infantrySpriteTemplates[uKey] = unitSprite
		}
		return unitSprite.Clone(u)

	case *model.Emplacement:
		u := unit.(*model.Emplacement)
		uKey := u.Resource.File
		unitSprite, found := g.sprites.emplacementSpriteTemplates[uKey]
		if !found {
			relPath := fmt.Sprintf("%s/%s", model.EmplacementResourceType, u.Resource.Image)
			img := resources.GetSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = sprites.NewEmplacementSprite(u, scale, img)
			g.sprites.emplacementSpriteTemplates[uKey] = unitSprite
		}
		return unitSprite.Clone(u)

	default:
		panic(fmt.Errorf("unable to handle model.Unit from type %v", interfaceType))
	}
}

func (g *Game) getRaycastSprites() []raycaster.Sprite {
	raycastSprites := make([]raycaster.Sprite, 0, 512)

	camPos := g.player.CameraPosXY()

	count := 0
	for _, spriteMap := range g.sprites.sprites {
		spriteMap.Range(func(k, _ interface{}) bool {
			spriteInterface := k.(raycaster.Sprite)
			sprite := getSpriteFromInterface(spriteInterface)
			// for now this is sufficient, but for much larger amounts of sprites may need goroutines to divide up the work
			// only include map sprites within fast approximation of render distance
			doSprite := g.renderDistance < 0 || g.player.Target() == sprite.Entity ||
				model.PointInProximity(g.renderDistance, camPos.X, camPos.Y, sprite.Pos().X, sprite.Pos().Copy().Y)
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

	if g.player.DebugCameraTarget() != nil {
		// add player sprite to be raycasted only when camera attached to a target
		raycastSprites = append(raycastSprites, g.player.sprite)
		count++
	}

	return raycastSprites[:count]
}

func (g *Game) getUnitSprites() []*sprites.Sprite {
	sprites := make([]*sprites.Sprite, 0, 64)
	for spriteType := range g.sprites.sprites {
		g.sprites.sprites[spriteType].Range(func(k, _ interface{}) bool {
			if !g.isInteractiveType(spriteType) {
				// only include certain sprite types (skip projectiles, effects, etc.)
				return true
			}

			s := getSpriteFromInterface(k.(raycaster.Sprite))
			if s.IsDestroyed() {
				return true
			}
			sprites = append(sprites, s)
			return true
		})
	}
	return sprites
}

func (g *Game) getSpriteUnits() []model.Unit {
	uSprites := g.getUnitSprites()
	units := make([]model.Unit, 0, len(uSprites))
	for _, s := range uSprites {
		u := s.Unit()
		if u == nil {
			continue
		}
		units = append(units, u)
	}
	return units
}

func (g *Game) getProximityUnitSprites(pos *geom.Vector2, distance float64) []*proximitySprite {
	sprites := make([]*proximitySprite, 0, 64)
	for spriteType := range g.sprites.sprites {
		g.sprites.sprites[spriteType].Range(func(k, _ interface{}) bool {
			if !g.isInteractiveType(spriteType) {
				// only include certain sprite types (skip projectiles, effects, etc.)
				return true
			}
			s := getSpriteFromInterface(k.(raycaster.Sprite))
			if s.IsDestroyed() {
				return true
			}
			sPos := s.Pos()

			// fast proximity check
			if !model.PointInProximity(distance, pos.X, pos.Y, sPos.X, sPos.Y) {
				return true
			}

			// exact distance check
			sDist := geom.Distance(pos.X, pos.Y, sPos.X, sPos.Y)
			if sDist > distance {
				return true
			}

			sprites = append(sprites, &proximitySprite{sprite: s, distance: sDist})
			return true
		})
	}

	if g.player.sprite != nil && !g.aiIgnorePlayer {
		// include player also
		sPos := g.player.sprite.Pos()
		// fast proximity check
		if model.PointInProximity(distance, pos.X, pos.Y, sPos.X, sPos.Y) {
			// exact distance check
			sDist := geom.Distance(pos.X, pos.Y, sPos.X, sPos.Y)
			if sDist <= distance {
				sprites = append(sprites, &proximitySprite{sprite: g.player.sprite, distance: sDist})
			}
		}
	}

	// sort sprites by distance
	sort.Slice(sprites, func(i, j int) bool { return sprites[i].distance < sprites[j].distance })

	return sprites
}

func (g *Game) getProximitySpriteUnits(pos *geom.Vector2, distance float64) []*proximityUnit {
	uSprites := g.getProximityUnitSprites(pos, distance)
	units := make([]*proximityUnit, 0, len(uSprites))
	for _, s := range uSprites {
		u := s.sprite.Unit()
		if u == nil {
			continue
		}
		units = append(units, &proximityUnit{unit: u, distance: s.distance})
	}
	return units
}

func getSpriteType(sInterface raycaster.Sprite) SpriteType {
	switch interfaceType := sInterface.(type) {
	case *sprites.Sprite:
		return MapSpriteType
	case *sprites.MechSprite:
		return MechSpriteType
	case *sprites.VehicleSprite:
		return VehicleSpriteType
	case *sprites.VTOLSprite:
		return VTOLSpriteType
	case *sprites.InfantrySprite:
		return InfantrySpriteType
	case *sprites.EmplacementSprite:
		return EmplacementSpriteType
	case *sprites.ProjectileSprite:
		return ProjectileSpriteType
	case *sprites.EffectSprite:
		return EffectSpriteType
	default:
		panic(fmt.Errorf("unable to get SpriteType from sprite interface type %v", interfaceType))
	}
}

func getSpriteFromInterface(sInterface raycaster.Sprite) *sprites.Sprite {
	if sInterface == nil {
		return nil
	}

	sType := getSpriteType(sInterface)
	switch sType {
	case MapSpriteType:
		return sInterface.(*sprites.Sprite)
	case MechSpriteType:
		return sInterface.(*sprites.MechSprite).Sprite
	case VehicleSpriteType:
		return sInterface.(*sprites.VehicleSprite).Sprite
	case VTOLSpriteType:
		return sInterface.(*sprites.VTOLSprite).Sprite
	case InfantrySpriteType:
		return sInterface.(*sprites.InfantrySprite).Sprite
	case EmplacementSpriteType:
		return sInterface.(*sprites.EmplacementSprite).Sprite
	case ProjectileSpriteType:
		return sInterface.(*sprites.ProjectileSprite).Sprite
	case EffectSpriteType:
		return sInterface.(*sprites.EffectSprite).Sprite
	default:
		panic(fmt.Errorf("unable to get model.Sprite from type %v", sType))
	}
}

func getEntityFromInterface(sInterface raycaster.Sprite) model.Entity {
	sType := getSpriteType(sInterface)
	switch sType {
	case MapSpriteType:
		return sInterface.(*sprites.Sprite).Entity
	case MechSpriteType:
		return sInterface.(*sprites.MechSprite).Entity
	case VehicleSpriteType:
		return sInterface.(*sprites.VehicleSprite).Entity
	case VTOLSpriteType:
		return sInterface.(*sprites.VTOLSprite).Entity
	case InfantrySpriteType:
		return sInterface.(*sprites.InfantrySprite).Entity
	case EmplacementSpriteType:
		return sInterface.(*sprites.EmplacementSprite).Entity
	case ProjectileSpriteType:
		return sInterface.(*sprites.ProjectileSprite).Entity
	case EffectSpriteType:
		return sInterface.(*sprites.EffectSprite).Entity
	default:
		panic(fmt.Errorf("unable to get model.Entity from type %v", sType))
	}
}

func (g *Game) getSpriteFromEntity(entity model.Entity) *sprites.Sprite {
	var found *sprites.Sprite
	for spriteType := range g.sprites.sprites {
		g.sprites.sprites[spriteType].Range(func(k, _ interface{}) bool {
			if !g.isInteractiveType(spriteType) {
				// only include certain sprite types (skip projectiles, effects, etc.)
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

func (g *Game) getMapSpriteFromEntity(entity model.Entity) *sprites.Sprite {
	var found *sprites.Sprite

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
