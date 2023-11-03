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

func (g *Game) explosionEffect(x, y, z, angle, pitch float64) *render.EffectSprite {
	// TODO: return random explosion
	e := explosionEffects["01"].Clone()
	e.SetPos(&geom.Vector2{X: x, Y: y})
	e.SetPosZ(z)
	return e
}

func (g *Game) smokeEffect(x, y, z, angle, pitch float64) *render.EffectSprite {
	// TODO: return random smoke
	e := smokeEffects["01"].Clone()
	e.SetPos(&geom.Vector2{X: x, Y: y})
	e.SetPosZ(z)

	// TODO: give Z velocity so it rises
	e.SetVelocityZ(0.003) // TODO: define velocity of rise in resource model

	// TODO: give some small angle/pitch so it won't rise perfectly vertically (fake wind)
	return e
}
