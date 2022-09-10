package game

import (
	"fmt"
	"math"
	"sync"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go"
)

type SpriteHandler struct {
	sprites map[SpriteType]*sync.Map
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
		sprites: make(map[SpriteType]*sync.Map, TotalSpriteTypes),
	}
	s.sprites[MechSpriteType] = &sync.Map{}
	s.sprites[MapSpriteType] = &sync.Map{}
	s.sprites[ProjectileSpriteType] = &sync.Map{}
	s.sprites[EffectSpriteType] = &sync.Map{}

	return s
}

func (s *SpriteHandler) addMapSprite(sprite *model.Sprite) {
	s.sprites[MapSpriteType].Store(sprite, struct{}{})
}

func (s *SpriteHandler) deleteMapSprite(sprite *model.Sprite) {
	s.sprites[MapSpriteType].Delete(sprite)
}

func (s *SpriteHandler) addMechSprite(mech *model.MechSprite) {
	s.sprites[MechSpriteType].Store(mech, struct{}{})
}

func (s *SpriteHandler) deleteMechSprite(mech *model.MechSprite) {
	s.sprites[MechSpriteType].Delete(mech)
}

func (s *SpriteHandler) addProjectile(projectile *model.Projectile) {
	s.sprites[ProjectileSpriteType].Store(projectile, struct{}{})
}

func (s *SpriteHandler) deleteProjectile(projectile *model.Projectile) {
	s.sprites[ProjectileSpriteType].Delete(projectile)
}

func (s *SpriteHandler) addEffect(effect *model.Effect) {
	s.sprites[EffectSpriteType].Store(effect, struct{}{})
}

func (s *SpriteHandler) deleteEffect(effect *model.Effect) {
	s.sprites[EffectSpriteType].Delete(effect)
}

func (g *Game) getRaycastSprites() []raycaster.Sprite {
	raycastSprites := make([]raycaster.Sprite, 0, 512)

	count := 0
	for _, spriteMap := range g.sprites.sprites {
		spriteMap.Range(func(k, _ interface{}) bool {
			spriteInterface := k.(raycaster.Sprite)
			sprite := getSpriteFromInterface(spriteInterface)
			// for now this is sufficient, but for much larger amounts of sprites may need goroutines to divide up the work
			// only include map sprites within fast approximation of render distance
			doSprite := g.renderDistance < 0 ||
				(math.Abs(sprite.Position().X-g.player.Position().X) <= g.renderDistance &&
					math.Abs(sprite.Position().Y-g.player.Position().Y) <= g.renderDistance)
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

	return raycastSprites[:count]
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

func getEntityFromInterface(sInterface raycaster.Sprite) model.Entity {
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
