package game

import (
	"fmt"
	"math"
	"sync"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"

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

func getSpriteFromInterface(sInterface raycaster.Sprite) *render.Sprite {
	switch interfaceType := sInterface.(type) {
	case *render.Sprite:
		return sInterface.(*render.Sprite)
	case *render.MechSprite:
		return sInterface.(*render.MechSprite).Sprite
	case *render.ProjectileSprite:
		return sInterface.(*render.ProjectileSprite).Sprite
	case *render.EffectSprite:
		return sInterface.(*render.EffectSprite).Sprite
	default:
		panic(fmt.Errorf("unable to get model.Sprite from type %v", interfaceType))
	}
}

func getEntityFromInterface(sInterface raycaster.Sprite) model.Entity {
	switch interfaceType := sInterface.(type) {
	case *render.Sprite:
		return sInterface.(*render.Sprite).Entity
	case *render.MechSprite:
		return sInterface.(*render.MechSprite).Entity
	case *render.ProjectileSprite:
		return sInterface.(*render.ProjectileSprite).Entity
	case *render.EffectSprite:
		return sInterface.(*render.EffectSprite).Entity
	default:
		panic(fmt.Errorf("unable to get model.Entity from type %v", interfaceType))
	}
}
