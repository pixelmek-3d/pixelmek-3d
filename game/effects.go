package game

import (
	"fmt"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"
	"github.com/harbdog/pixelmek-3d/game/resources/effects"
	"github.com/harbdog/raycaster-go/geom"
)

var (
	explosionEffects map[string]*render.EffectSprite
	smokeEffects     map[string]*render.EffectSprite
)

func init() {
	explosionEffects = make(map[string]*render.EffectSprite)
	smokeEffects = make(map[string]*render.EffectSprite)
}

func (g *Game) loadSpecialEffects() {
	for key, fx := range effects.Explosions {
		if _, ok := explosionEffects[key]; ok {
			continue
		}
		// load the explosion effect sprite template
		// TODO: load sound effect audio
		effectRelPath := fmt.Sprintf("%s/%s", model.EffectsResourceType, fx.Image)
		effectImg := getSpriteFromFile(effectRelPath)
		eColumns, eRows, eAnimationRate := 1, 1, 1
		if fx.ImageSheet != nil {
			eColumns = fx.ImageSheet.Columns
			eRows = fx.ImageSheet.Rows
			eAnimationRate = fx.ImageSheet.AnimationRate
		}
		eSpriteTemplate := render.NewAnimatedEffect(fx.Scale, effectImg, eColumns, eRows, eAnimationRate, 1)
		explosionEffects[key] = eSpriteTemplate
	}

	for key, fx := range effects.Smokes {
		if _, ok := smokeEffects[key]; ok {
			continue
		}
		// load the smoke effect sprite template
		effectRelPath := fmt.Sprintf("%s/%s", model.EffectsResourceType, fx.Image)
		effectImg := getSpriteFromFile(effectRelPath)
		eColumns, eRows, eAnimationRate := 1, 1, 1
		if fx.ImageSheet != nil {
			eColumns = fx.ImageSheet.Columns
			eRows = fx.ImageSheet.Rows
			eAnimationRate = fx.ImageSheet.AnimationRate
		}
		eSpriteTemplate := render.NewAnimatedEffect(fx.Scale, effectImg, eColumns, eRows, eAnimationRate, 1)
		smokeEffects[key] = eSpriteTemplate
	}
}

func (g *Game) spawnMechDestructEffects(s *render.MechSprite) {
	if s.AnimationFrameCounter() != 0 {
		// do not spawn effects every tick
		return
	}

	x, y, z := s.Pos().X, s.Pos().Y, s.PosZ()
	r, h := s.CollisionRadius(), s.CollisionHeight()

	// use size of mech to determine number of effects to randomly spawn each time
	m := s.Mech()
	numFx := 1
	if m.Class() >= model.MECH_HEAVY {
		numFx = 2
	}

	for i := 0; i < numFx; i++ {
		xFx := x + randFloat(-r, r)
		yFx := y + randFloat(-r, r)
		zFx := z + randFloat(h/4, h)

		explosionFx := g.randExplosionEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(explosionFx)

		smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(smokeFx)
	}
}

func (g *Game) randExplosionEffect(x, y, z, angle, pitch float64) *render.EffectSprite {
	// TODO: return random explosion
	e := explosionEffects["07"].Clone()
	e.SetPos(&geom.Vector2{X: x, Y: y})
	e.SetPosZ(z)

	// TODO: give small negative Z velocity so it falls with the unit being destroyed
	return e
}

func (g *Game) randSmokeEffect(x, y, z, angle, pitch float64) *render.EffectSprite {
	// TODO: return random smoke
	e := smokeEffects["01.5"].Clone()
	e.SetPos(&geom.Vector2{X: x, Y: y})
	e.SetPosZ(z)

	// give Z velocity so it rises
	e.SetVelocityZ(0.003) // TODO: define velocity of rise in resource model

	// TODO: give some small angle/pitch so it won't rise perfectly vertically (fake wind)
	return e
}
