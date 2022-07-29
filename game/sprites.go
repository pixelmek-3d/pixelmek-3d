package game

import (
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go"
)

type SpriteHandler struct {
	mapSprites  map[*model.Sprite]struct{}
	mechSprites map[*model.MechSprite]struct{}
	projectiles map[*model.Projectile]struct{}
	effects     map[*model.Effect]struct{}
}

func NewSpriteHandler() *SpriteHandler {
	s := &SpriteHandler{}
	s.effects = make(map[*model.Effect]struct{}, 1024)
	s.projectiles = make(map[*model.Projectile]struct{}, 1024)
	s.mapSprites = make(map[*model.Sprite]struct{}, 512)
	s.mechSprites = make(map[*model.MechSprite]struct{}, 128)
	return s
}

func (s *SpriteHandler) addSprite(sprite *model.Sprite) {
	s.mapSprites[sprite] = struct{}{}
}

func (s *SpriteHandler) addMechSprite(mech *model.MechSprite) {
	s.mechSprites[mech] = struct{}{}
}

func (s *SpriteHandler) deleteSprite(sprite *model.Sprite) {
	delete(s.mapSprites, sprite)
}

func (s *SpriteHandler) addProjectile(projectile *model.Projectile) {
	s.projectiles[projectile] = struct{}{}
}

func (s *SpriteHandler) deleteProjectile(projectile *model.Projectile) {
	delete(s.projectiles, projectile)
}

func (s *SpriteHandler) addEffect(effect *model.Effect) {
	s.effects[effect] = struct{}{}
}

func (s *SpriteHandler) deleteEffect(effect *model.Effect) {
	delete(s.effects, effect)
}

func (g *Game) getRaycastSprites() []raycaster.Sprite {
	numSprites := len(g.sprites.mapSprites) + len(g.sprites.mechSprites) + len(g.sprites.projectiles) + len(g.sprites.effects) + len(g.clutter.sprites)
	raycastSprites := make([]raycaster.Sprite, numSprites)
	index := 0

	for sprite := range g.sprites.mapSprites {
		raycastSprites[index] = sprite
		index += 1
	}
	for clutter := range g.clutter.sprites {
		raycastSprites[index] = clutter
		index += 1
	}
	for mech := range g.sprites.mechSprites {
		raycastSprites[index] = mech
		index += 1
	}
	for projectile := range g.sprites.projectiles {
		raycastSprites[index] = projectile.Sprite
		index += 1
	}
	for effect := range g.sprites.effects {
		raycastSprites[index] = effect.Sprite
		index += 1
	}

	return raycastSprites
}
