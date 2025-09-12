package game

import (
	"fmt"
	"sort"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/sprites"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type proximitySprite struct {
	sprite   *sprites.Sprite
	distance float64
}

type proximityUnit struct {
	unit     model.Unit
	distance float64
}

func (g *Game) createUnitSprite(unit model.Unit) raycaster.Sprite {
	switch interfaceType := unit.(type) {
	case *model.Mech:
		u := unit.(*model.Mech)
		uKey := u.Resource.File
		unitSprite, found := g.sprites.MechSpriteTemplates[uKey]
		if !found {
			relPath := fmt.Sprintf("%s/%s", model.MechResourceType, u.Resource.Image)
			img := resources.GetSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = sprites.NewMechSprite(u, scale, img)
			g.sprites.MechSpriteTemplates[uKey] = unitSprite
		}
		return unitSprite.Clone(u)

	case *model.Vehicle:
		u := unit.(*model.Vehicle)
		uKey := u.Resource.File
		unitSprite, found := g.sprites.VehicleSpriteTemplates[uKey]
		if !found {
			relPath := fmt.Sprintf("%s/%s", model.VehicleResourceType, u.Resource.Image)
			img := resources.GetSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = sprites.NewVehicleSprite(u, scale, img)
			g.sprites.VehicleSpriteTemplates[uKey] = unitSprite
		}
		return unitSprite.Clone(u)

	case *model.VTOL:
		u := unit.(*model.VTOL)
		uKey := u.Resource.File
		unitSprite, found := g.sprites.VTOLSpriteTemplates[uKey]
		if !found {
			relPath := fmt.Sprintf("%s/%s", model.VTOLResourceType, u.Resource.Image)
			img := resources.GetSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = sprites.NewVTOLSprite(u, scale, img)
			g.sprites.VTOLSpriteTemplates[uKey] = unitSprite
		}
		return unitSprite.Clone(u)

	case *model.Infantry:
		u := unit.(*model.Infantry)
		uKey := u.Resource.File
		unitSprite, found := g.sprites.InfantrySpriteTemplates[uKey]
		if !found {
			relPath := fmt.Sprintf("%s/%s", model.InfantryResourceType, u.Resource.Image)
			img := resources.GetSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = sprites.NewInfantrySprite(u, scale, img)
			g.sprites.InfantrySpriteTemplates[uKey] = unitSprite
		}
		return unitSprite.Clone(u)

	case *model.Emplacement:
		u := unit.(*model.Emplacement)
		uKey := u.Resource.File
		unitSprite, found := g.sprites.EmplacementSpriteTemplates[uKey]
		if !found {
			relPath := fmt.Sprintf("%s/%s", model.EmplacementResourceType, u.Resource.Image)
			img := resources.GetSpriteFromFile(relPath)
			scale := convertHeightToScale(u.Resource.Height, img.Bounds().Dy(), u.Resource.HeightPxGap)

			unitSprite = sprites.NewEmplacementSprite(u, scale, img)
			g.sprites.EmplacementSpriteTemplates[uKey] = unitSprite
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
	g.sprites.Range(func(k, _ interface{}) bool {
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
	for _, spriteType := range g.sprites.SpriteTypes() {
		g.sprites.RangeByType(spriteType, func(k, _ interface{}) bool {
			if !isInteractiveType(spriteType) {
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
	for _, spriteType := range g.sprites.SpriteTypes() {
		g.sprites.RangeByType(spriteType, func(k, _ interface{}) bool {
			if !isInteractiveType(spriteType) {
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

func getSpriteFromInterface(sInterface raycaster.Sprite) *sprites.Sprite {
	if sInterface == nil {
		return nil
	}

	sType := sprites.GetSpriteType(sInterface)
	switch sType {
	case sprites.MapSpriteType:
		return sInterface.(*sprites.Sprite)
	case sprites.MechSpriteType:
		return sInterface.(*sprites.MechSprite).Sprite
	case sprites.VehicleSpriteType:
		return sInterface.(*sprites.VehicleSprite).Sprite
	case sprites.VTOLSpriteType:
		return sInterface.(*sprites.VTOLSprite).Sprite
	case sprites.InfantrySpriteType:
		return sInterface.(*sprites.InfantrySprite).Sprite
	case sprites.EmplacementSpriteType:
		return sInterface.(*sprites.EmplacementSprite).Sprite
	case sprites.ProjectileSpriteType:
		return sInterface.(*sprites.ProjectileSprite).Sprite
	case sprites.EffectSpriteType:
		return sInterface.(*sprites.EffectSprite).Sprite
	default:
		panic(fmt.Errorf("unable to get model.Sprite from type %v", sType))
	}
}

func getEntityFromInterface(sInterface raycaster.Sprite) model.Entity {
	sType := sprites.GetSpriteType(sInterface)
	switch sType {
	case sprites.MapSpriteType:
		return sInterface.(*sprites.Sprite).Entity
	case sprites.MechSpriteType:
		return sInterface.(*sprites.MechSprite).Entity
	case sprites.VehicleSpriteType:
		return sInterface.(*sprites.VehicleSprite).Entity
	case sprites.VTOLSpriteType:
		return sInterface.(*sprites.VTOLSprite).Entity
	case sprites.InfantrySpriteType:
		return sInterface.(*sprites.InfantrySprite).Entity
	case sprites.EmplacementSpriteType:
		return sInterface.(*sprites.EmplacementSprite).Entity
	case sprites.ProjectileSpriteType:
		return sInterface.(*sprites.ProjectileSprite).Entity
	case sprites.EffectSpriteType:
		return sInterface.(*sprites.EffectSprite).Entity
	default:
		panic(fmt.Errorf("unable to get model.Entity from type %v", sType))
	}
}

func (g *Game) getSpriteFromEntity(entity model.Entity) *sprites.Sprite {
	var found *sprites.Sprite
	for _, spriteType := range g.sprites.SpriteTypes() {
		g.sprites.RangeByType(spriteType, func(k, _ interface{}) bool {
			if !isInteractiveType(spriteType) {
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

	g.sprites.RangeByType(sprites.MapSpriteType, func(k, _ interface{}) bool {
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
