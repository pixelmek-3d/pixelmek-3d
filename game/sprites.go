package game

import (
	"fmt"
	"math"
	"sync"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go"
)

type SpriteHandler struct {
	sprites  map[SpriteType]map[raycaster.Sprite]struct{}
	mutexMap map[SpriteType]*sync.RWMutex
}

type SpriteType int

const (
	MapSpriteType SpriteType = iota
	MechSpriteType
	ProjectileSpriteType
	EffectSpriteType
	TotalSpriteTypes
)

func NewSpriteHandler() *SpriteHandler {
	s := &SpriteHandler{
		sprites:  make(map[SpriteType]map[raycaster.Sprite]struct{}, TotalSpriteTypes),
		mutexMap: make(map[SpriteType]*sync.RWMutex),
	}
	s.sprites[MechSpriteType] = make(map[raycaster.Sprite]struct{}, 128)
	s.sprites[MapSpriteType] = make(map[raycaster.Sprite]struct{}, 512)
	s.sprites[ProjectileSpriteType] = make(map[raycaster.Sprite]struct{}, 1024)
	s.sprites[EffectSpriteType] = make(map[raycaster.Sprite]struct{}, 1024)

	for spriteType := range s.sprites {
		// TODO: consider trying sync Map, it may be faster for read/delete, but slower for writes?
		// create map mutex to lock/unlock maps for concurrent safety
		s.mutexMap[spriteType] = &sync.RWMutex{}
	}

	return s
}

func (s *SpriteHandler) totalSprites() int {
	total := 0
	for _, spriteMap := range s.sprites {
		total += len(spriteMap)
	}

	return total
}

func (s *SpriteHandler) addMapSprite(sprite *model.Sprite) {
	s.mutexMap[MapSpriteType].Lock()
	s.sprites[MapSpriteType][sprite] = struct{}{}
	s.mutexMap[MapSpriteType].Unlock()
}

func (s *SpriteHandler) deleteMapSprite(sprite *model.Sprite) {
	s.mutexMap[MapSpriteType].Lock()
	delete(s.sprites[MapSpriteType], sprite)
	s.mutexMap[MapSpriteType].Unlock()
}

func (s *SpriteHandler) addMechSprite(mech *model.MechSprite) {
	s.mutexMap[MechSpriteType].Lock()
	s.sprites[MechSpriteType][mech] = struct{}{}
	s.mutexMap[MechSpriteType].Unlock()
}

func (s *SpriteHandler) deleteMechSprite(mech *model.MechSprite) {
	s.mutexMap[MechSpriteType].Lock()
	delete(s.sprites[MechSpriteType], mech)
	s.mutexMap[MechSpriteType].Unlock()
}

func (s *SpriteHandler) addProjectile(projectile *model.Projectile) {
	s.mutexMap[ProjectileSpriteType].Lock()
	s.sprites[ProjectileSpriteType][projectile] = struct{}{}
	s.mutexMap[ProjectileSpriteType].Unlock()
}

func (s *SpriteHandler) deleteProjectile(projectile *model.Projectile) {
	s.mutexMap[ProjectileSpriteType].Lock()
	delete(s.sprites[ProjectileSpriteType], projectile)
	s.mutexMap[ProjectileSpriteType].Unlock()
}

func (s *SpriteHandler) addEffect(effect *model.Effect) {
	s.mutexMap[EffectSpriteType].Lock()
	s.sprites[EffectSpriteType][effect] = struct{}{}
	s.mutexMap[EffectSpriteType].Unlock()
}

func (s *SpriteHandler) deleteEffect(effect *model.Effect) {
	s.mutexMap[EffectSpriteType].Lock()
	delete(s.sprites[EffectSpriteType], effect)
	s.mutexMap[EffectSpriteType].Unlock()
}

func (g *Game) getRaycastSprites() []raycaster.Sprite {
	numSprites := g.sprites.totalSprites() + len(g.clutter.sprites)
	raycastSprites := make([]raycaster.Sprite, numSprites)

	index := 0

	for _, spriteMap := range g.sprites.sprites {
		for spriteInterface := range spriteMap {
			sprite := getSpriteFromInterface(spriteInterface)
			// for now this is sufficient, but for much larger amounts of sprites may need goroutines to divide up the work
			// only include map sprites within fast approximation of render distance
			doSprite := g.renderDistance < 0 ||
				(math.Abs(sprite.Position.X-g.player.Position.X) <= g.renderDistance &&
					math.Abs(sprite.Position.Y-g.player.Position.Y) <= g.renderDistance)
			if doSprite {
				raycastSprites[index] = spriteInterface
				index += 1
			}
		}
	}
	for clutter := range g.clutter.sprites {
		raycastSprites[index] = clutter
		index += 1
	}

	return raycastSprites[:index]
}

func getSpriteFromInterface(sInterface raycaster.Sprite) *model.Sprite {
	switch interfaceType := sInterface.(type) {
	case *model.Sprite:
		return sInterface.(*model.Sprite)
	case *model.MechSprite:
		return sInterface.(*model.MechSprite).Sprite
	case *model.Projectile:
		return sInterface.(*model.Projectile).Sprite
	case *model.Effect:
		return sInterface.(*model.Effect).Sprite
	default:
		panic(fmt.Errorf("unable to get model.Sprite from type %v", interfaceType))
	}
}

func getEntityFromInterface(sInterface raycaster.Sprite) *model.Entity {
	switch interfaceType := sInterface.(type) {
	case *model.Sprite:
		return sInterface.(*model.Sprite).Entity
	case *model.MechSprite:
		return sInterface.(*model.MechSprite).Entity
	case *model.Projectile:
		return sInterface.(*model.Projectile).Entity
	case *model.Effect:
		return sInterface.(*model.Effect).Entity
	default:
		panic(fmt.Errorf("unable to get model.Entity from type %v", interfaceType))
	}
}
