package game

import (
	"fmt"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"
	"github.com/harbdog/pixelmek-3d/game/resources/effects"
	"github.com/harbdog/raycaster-go/geom"
)

var (
	bloodEffects     map[string]*render.EffectSprite
	explosionEffects map[string]*render.EffectSprite
	fireEffects      map[string]*render.EffectSprite
	smokeEffects     map[string]*render.EffectSprite
)

func init() {
	bloodEffects = make(map[string]*render.EffectSprite)
	explosionEffects = make(map[string]*render.EffectSprite)
	fireEffects = make(map[string]*render.EffectSprite)
	smokeEffects = make(map[string]*render.EffectSprite)
}

func (g *Game) loadSpecialEffects() {
	// load the blood effect sprite templates
	g._loadEffectSpritesFromResourceList(effects.Blood, bloodEffects)

	// load the explosion effect sprite templates
	g._loadEffectSpritesFromResourceList(effects.Explosions, explosionEffects)

	// load the fire effect sprite templates
	g._loadEffectSpritesFromResourceList(effects.Fires, fireEffects)

	// load the smoke effect sprite templates
	g._loadEffectSpritesFromResourceList(effects.Smokes, smokeEffects)
}

func (g *Game) _loadEffectSpritesFromResourceList(
	resourceMap map[string]*model.ModelEffectResource, spriteMap map[string]*render.EffectSprite,
) {
	for key, fx := range resourceMap {
		if _, ok := spriteMap[key]; ok {
			continue
		}
		// load the blood effect sprite template
		effectRelPath := fmt.Sprintf("%s/%s", model.EffectsResourceType, fx.Image)
		effectImg := getSpriteFromFile(effectRelPath)

		eSpriteTemplate := render.NewAnimatedEffect(fx, effectImg, 1)
		spriteMap[key] = eSpriteTemplate
	}
}

func (g *Game) spawnGenericDestroyEffects(s *render.Sprite) (duration int) {
	x, y, z := s.Pos().X, s.Pos().Y, s.PosZ()
	r, h := s.CollisionRadius(), s.CollisionHeight()

	numFx := 7 // TODO: alter number of effects based on sprite dimensions
	for i := 0; i < numFx; i++ {
		xFx := x + randFloat(-r/2, r/2)
		yFx := y + randFloat(-r/2, r/2)
		zFx := z + randFloat(h/8, h)

		fireFx := g.randFireEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(fireFx)

		smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(smokeFx)

		fxDuration := fireFx.AnimationDuration()
		if fxDuration > duration {
			duration = fxDuration
		}
	}
	return
}

func (g *Game) spawnMechDestroyEffects(s *render.MechSprite) (duration int) {
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

	// only play one explosion audio track at a time
	playedOneAudio := false

	for i := 0; i < numFx; i++ {
		xFx := x + randFloat(-r, r)
		yFx := y + randFloat(-r, r)
		zFx := z + randFloat(h/4, h)

		explosionFx := g.randExplosionEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(explosionFx)

		smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(smokeFx)

		if !playedOneAudio {
			g.audio.PlayEffectAudio(g, explosionFx)
			playedOneAudio = true
		}

		fxDuration := explosionFx.AnimationDuration()
		if fxDuration > duration {
			duration = fxDuration
		}
	}
	return
}

func (g *Game) spawnInfantryDestroyEffects(s *render.InfantrySprite) (duration int) {
	x, y, z := s.Pos().X, s.Pos().Y, s.PosZ()
	r, h := s.CollisionRadius(), s.CollisionHeight()

	numFx := 4 // TODO: alter number of effects based on sprite dimensions
	for i := 0; i < numFx; i++ {
		xFx := x + randFloat(-r, r)
		yFx := y + randFloat(-r, r)
		zFx := z + randFloat(0, h)

		bloodFx := g.randBloodEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(bloodFx)

		fxDuration := bloodFx.AnimationDuration()
		if fxDuration > duration {
			duration = fxDuration
		}
	}
	return
}

func (g *Game) randBloodEffect(x, y, z, angle, pitch float64) *render.EffectSprite {
	// return random blood effect
	randKey := effects.RandBloodKey()
	e := bloodEffects[randKey].Clone()
	e.SetPos(&geom.Vector2{X: x, Y: y})
	e.SetPosZ(z)
	return e
}

func (g *Game) randExplosionEffect(x, y, z, angle, pitch float64) *render.EffectSprite {
	// return random explosion effect
	randKey := effects.RandExplosionKey()
	e := explosionEffects[randKey].Clone()
	e.SetPos(&geom.Vector2{X: x, Y: y})
	e.SetPosZ(z)

	// TODO: give small negative Z velocity so it falls with the unit being destroyed?
	return e
}

func (g *Game) randFireEffect(x, y, z, angle, pitch float64) *render.EffectSprite {
	// return random fire effect
	randKey := effects.RandFireKey()
	e := fireEffects[randKey].Clone()
	e.SetPos(&geom.Vector2{X: x, Y: y})
	e.SetPosZ(z)
	return e
}

func (g *Game) randSmokeEffect(x, y, z, angle, pitch float64) *render.EffectSprite {
	// return random smoke effect
	randKey := effects.RandSmokeKey()
	e := smokeEffects[randKey].Clone()
	e.SetPos(&geom.Vector2{X: x, Y: y})
	e.SetPosZ(z)

	// give Z velocity so it rises
	e.SetVelocityZ(0.003) // TODO: define velocity of rise in resource model

	// TODO: give some small angle/pitch so it won't rise perfectly vertically (fake wind)
	return e
}
