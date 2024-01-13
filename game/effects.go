package game

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources/effects"
)

var (
	jumpJetEffect     *render.EffectSprite
	attachedJJEffects map[*render.Sprite]*render.EffectSprite

	bloodEffects     map[string]*render.EffectSprite
	explosionEffects map[string]*render.EffectSprite
	fireEffects      map[string]*render.EffectSprite
	smokeEffects     map[string]*render.EffectSprite
)

func init() {
	attachedJJEffects = make(map[*render.Sprite]*render.EffectSprite)
	bloodEffects = make(map[string]*render.EffectSprite)
	explosionEffects = make(map[string]*render.EffectSprite)
	fireEffects = make(map[string]*render.EffectSprite)
	smokeEffects = make(map[string]*render.EffectSprite)
}

func (g *Game) loadSpecialEffects() {
	// load the jump jet effect sprite template
	jumpJetImg := _getEffectImageFromResource(effects.JumpJet)
	jumpJetEffect = render.NewAnimatedEffect(effects.JumpJet, jumpJetImg, math.MaxInt)
	for s := range attachedJJEffects {
		delete(attachedJJEffects, s)
	}

	// load the blood effect sprite templates
	_loadEffectSpritesFromResourceList(effects.Blood, bloodEffects)

	// load the explosion effect sprite templates
	_loadEffectSpritesFromResourceList(effects.Explosions, explosionEffects)

	// load the fire effect sprite templates
	_loadEffectSpritesFromResourceList(effects.Fires, fireEffects)

	// load the smoke effect sprite templates
	_loadEffectSpritesFromResourceList(effects.Smokes, smokeEffects)
}

func _getEffectImageFromResource(r *model.ModelEffectResource) *ebiten.Image {
	effectRelPath := fmt.Sprintf("%s/%s", model.EffectsResourceType, r.Image)
	return getSpriteFromFile(effectRelPath)
}

func _loadEffectSpritesFromResourceList(
	resourceMap map[string]*model.ModelEffectResource, spriteMap map[string]*render.EffectSprite,
) {
	for key, fx := range resourceMap {
		if _, ok := spriteMap[key]; ok {
			continue
		}
		// load the blood effect sprite template
		eSpriteTemplate := render.NewAnimatedEffect(fx, _getEffectImageFromResource(fx), 1)
		spriteMap[key] = eSpriteTemplate
	}
}

func (g *Game) spawnJumpJetEffect(s *render.Sprite) {
	_, found := attachedJJEffects[s]
	if found {
		// do not spawn another effect
		return
	}

	jumpFx := jumpJetEffect.Clone()

	jumpFx.SetScale(s.Scale())
	jumpFx.AttachedTo = s
	jumpFx.AttachedDepth = 0.01

	g.sprites.addEffect(jumpFx)

	// illuminate source sprite unit jump jetting
	s.SetIlluminationPeriod(5000, 0.35)

	// keep track of effect so only one is attached and can be deleted later
	attachedJJEffects[s] = jumpFx
}

func (g *Game) removeJumpJetEffect(s *render.Sprite) {
	jumpFx, found := attachedJJEffects[s]
	if found {
		g.sprites.deleteEffect(jumpFx)
		delete(attachedJJEffects, s)
	}
}

func (g *Game) spawnGenericDestroyEffects(s *render.Sprite, spawnFires bool) (duration int) {
	x, y, z := s.Pos().X, s.Pos().Y, s.PosZ()
	r, h := s.CollisionRadius(), s.CollisionHeight()

	numFx := 7 // TODO: alter number of effects based on sprite dimensions
	if !spawnFires {
		numFx = int(math.Ceil(float64(numFx) * 1.25))
	}
	for i := 0; i < numFx; i++ {
		xFx := x + randFloat(-r/2, r/2)
		yFx := y + randFloat(-r/2, r/2)
		zFx := z + randFloat(h/8, h)

		if spawnFires {
			fireFx := g.randFireEffect(xFx, yFx, zFx, s.Heading(), 0)
			g.sprites.addEffect(fireFx)

			fxDuration := fireFx.AnimationDuration()
			if fxDuration > duration {
				duration = fxDuration
			}
		}

		smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(smokeFx)
		if !spawnFires {
			// when not spawning fires, no duration implied
			duration = 0
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

	for i := 0; i < numFx; i++ {
		xFx := x + randFloat(-r, r)
		yFx := y + randFloat(-r, r)
		zFx := z + randFloat(h/4, h)

		explosionFx := g.randExplosionEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(explosionFx)

		smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(smokeFx)

		if i == 0 {
			// only play one audio track at a time
			g.audio.PlayEffectAudio(g, explosionFx)
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

func (g *Game) spawnVehicleDestroyEffects(s *render.VehicleSprite) (duration int) {
	x, y, z := s.Pos().X, s.Pos().Y, s.PosZ()
	r, h := s.CollisionRadius(), s.CollisionHeight()

	numFx := 5 // TODO: alter number of effects based on sprite dimensions
	for i := 0; i < numFx; i++ {
		xFx := x + randFloat(-r, r)
		yFx := y + randFloat(-r, r)
		zFx := z + randFloat(0, h)

		explosionFx := g.randExplosionEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(explosionFx)

		smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(smokeFx)

		if i == 0 || i == numFx/2 {
			// only play two audio tracks for now since they are played at once
			g.audio.PlayEffectAudio(g, explosionFx)
		}

		fxDuration := explosionFx.AnimationDuration()
		if fxDuration > duration {
			duration = fxDuration
		}
	}
	return
}

func (g *Game) spawnVTOLDestroyEffects(s *render.VTOLSprite, spawnExplosions bool) (duration int) {
	x, y, z := s.Pos().X, s.Pos().Y, s.PosZ()
	r, h := s.CollisionRadius(), s.CollisionHeight()

	numFx := 5 // TODO: alter number of effects based on sprite dimensions
	for i := 0; i < numFx; i++ {
		xFx := x + randFloat(-r/2, r/2)
		yFx := y + randFloat(-r/2, r/2)
		zFx := z + randFloat(-h/2, h/2)

		if spawnExplosions {
			explosionFx := g.randExplosionEffect(xFx, yFx, zFx, s.Heading(), 0)
			g.sprites.addEffect(explosionFx)
			if i == 0 || i == numFx/2 {
				// only play two audio tracks for now since they are played at once
				g.audio.PlayEffectAudio(g, explosionFx)
			}

			fxDuration := explosionFx.AnimationDuration()
			if fxDuration > duration {
				duration = fxDuration
			}
		}

		smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(smokeFx)
		if !spawnExplosions {
			fxDuration := smokeFx.AnimationDuration()
			if fxDuration > duration {
				duration = fxDuration
			}
		}
	}
	return
}

func (g *Game) spawnEmplacementDestroyEffects(s *render.EmplacementSprite) (duration int) {
	x, y, z := s.Pos().X, s.Pos().Y, s.PosZ()
	r, h := s.CollisionRadius(), s.CollisionHeight()

	numFx := 5 // TODO: alter number of effects based on sprite dimensions
	for i := 0; i < numFx; i++ {
		xFx := x + randFloat(-r/2, r/2)
		yFx := y + randFloat(-r/2, r/2)
		zFx := z + randFloat(h/12, h)

		explosionFx := g.randExplosionEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(explosionFx)

		smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.addEffect(smokeFx)

		if i == 0 || i == numFx/2 {
			// only play two audio tracks for now since they are played at once
			g.audio.PlayEffectAudio(g, explosionFx)
		}

		fxDuration := explosionFx.AnimationDuration()
		if fxDuration > duration {
			duration = fxDuration
		}
	}

	numFireFx := 5
	for i := 0; i < numFireFx; i++ {
		// for emplacements, also add some random fires
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
