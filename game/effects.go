package game

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	renderFx "github.com/pixelmek-3d/pixelmek-3d/game/render/effects"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/sprites"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources/effects"
	log "github.com/sirupsen/logrus"
)

const (
	ejectName string = "EJECT"
)

var (
	crtShader *renderFx.CRT

	ejectLauncher model.Weapon
	ejectPod      *sprites.ProjectileSprite

	jumpJetEffect     *sprites.EffectSprite
	attachedJJEffects map[*sprites.Sprite]*sprites.EffectSprite

	bloodEffects     map[string]*sprites.EffectSprite
	explosionEffects map[string]*sprites.EffectSprite
	fireEffects      map[string]*sprites.EffectSprite
	smokeEffects     map[string]*sprites.EffectSprite
)

func init() {
	attachedJJEffects = make(map[*sprites.Sprite]*sprites.EffectSprite)
	bloodEffects = make(map[string]*sprites.EffectSprite)
	explosionEffects = make(map[string]*sprites.EffectSprite)
	fireEffects = make(map[string]*sprites.EffectSprite)
	smokeEffects = make(map[string]*sprites.EffectSprite)
}

func (g *Game) loadSpecialEffects() {
	// load shader effect
	_loadShaderEffects()

	// load the ejection pod template
	_loadEjectionPodResource(g)

	// load the jump jet effect sprite template
	_loadJumpJetEffectResource()

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
	return resources.GetSpriteFromFile(effectRelPath)
}

func _loadShaderEffects() {
	crtShader = renderFx.NewCRT()
}

func _loadJumpJetEffectResource() {
	if jumpJetEffect == nil {
		jumpJetImg := _getEffectImageFromResource(effects.JumpJet)
		jumpJetEffect = sprites.NewAnimatedEffect(effects.JumpJet, jumpJetImg, math.MaxInt)
	}
	for s := range attachedJJEffects {
		delete(attachedJJEffects, s)
	}
}

func _loadEjectionPodResource(g *Game) {
	if ejectPod != nil {
		return
	}

	// TODO: refactor to use same func as g.loadUnitWeapons
	weaponResource, err := g.resources.GetMissileWeaponResource("_ejection_pod")
	if err != nil {
		log.Error(err)
	}

	// need to use the projectile image size to find the unit collision conversion from pixels
	pResource := weaponResource.Projectile
	projectileRelPath := fmt.Sprintf("%s/%s", model.ProjectilesResourceType, pResource.Image)
	projectileImg := resources.GetSpriteFromFile(projectileRelPath)
	pColumns, pRows := 1, 1
	if pResource.ImageSheet != nil {
		pColumns = pResource.ImageSheet.Columns
		pRows = pResource.ImageSheet.Rows
	}

	pWidth, pHeight := projectileImg.Bounds().Dx(), projectileImg.Bounds().Dy()
	pWidth = pWidth / pColumns
	pHeight = pHeight / pRows

	// calculate sprite scale based on projectile diameter using collision height pixel size within sprite
	scale := convertProjectileDiameterToScale(pResource.Diameter, pHeight, pResource.CollisionPxHeight)
	pCollisionRadius, pCollisionHeight := convertOffsetFromPx(
		pResource.CollisionPxRadius, pResource.CollisionPxHeight, pWidth, pHeight, scale,
	)

	// create the pod as missile projectile model
	var modelPod model.Projectile
	ejectLauncher, modelPod = model.NewMissileWeapon(
		weaponResource, model.HEAD, pCollisionRadius, pCollisionHeight, &geom.Vector2{}, &geom.Vector2{}, nil,
	)

	// create the projectile and effect sprite templates
	eResource := weaponResource.Projectile.ImpactEffect
	effectRelPath := fmt.Sprintf("%s/%s", model.EffectsResourceType, eResource.Image)
	effectImg := resources.GetSpriteFromFile(effectRelPath)

	projectileImpactAudioFiles := make([]string, 1)
	projectileImpactAudioFiles = append(projectileImpactAudioFiles, pResource.ImpactEffect.Audio)
	projectileImpactAudioFiles = append(projectileImpactAudioFiles, pResource.ImpactEffect.RandAudio...)

	eSpriteTemplate := sprites.NewAnimatedEffect(eResource, effectImg, 1)
	ejectPod = sprites.NewAnimatedProjectile(
		&modelPod, scale, projectileImg, *eSpriteTemplate, projectileImpactAudioFiles,
	)
}

func _loadEffectSpritesFromResourceList(
	resourceMap map[string]*model.ModelEffectResource, spriteMap map[string]*sprites.EffectSprite,
) {
	for key, fx := range resourceMap {
		if _, ok := spriteMap[key]; ok {
			continue
		}
		// load the blood effect sprite template
		eSpriteTemplate := sprites.NewAnimatedEffect(fx, _getEffectImageFromResource(fx), 1)
		spriteMap[key] = eSpriteTemplate
	}
}

func (g *Game) spawnEjectionPod(s *sprites.Sprite) *sprites.ProjectileSprite {
	podSprite := ejectPod.Clone()
	podSprite.SetParent(s.Entity)
	podSprite.Projectile.SetWeapon(ejectLauncher)

	podSprite.SetPos(s.Pos().Copy())
	podSprite.SetPosZ(s.PosZ() + s.CollisionHeight())
	podSprite.SetPitch(geom.HalfPi)

	if s == g.player.sprite {
		podHeading := model.ClampAngle2Pi(g.player.cameraAngle)
		podSprite.SetHeading(podHeading)
	} else {
		podSprite.SetHeading(s.Entity.Heading())
	}

	// let ejection pod accelerate from zero for effect
	podSprite.SetVelocity(0)
	podSprite.Projectile.SetAcceleration(podSprite.Projectile.MaxVelocity() / (2 * model.TICKS_PER_SECOND))

	g.sprites.AddProjectile(podSprite)
	g.audio.PlayLocalWeaponFireAudio(ejectLauncher)

	return podSprite
}

func (g *Game) spawnEjectionPodSmokeEffects(s *sprites.ProjectileSprite) (duration int) {
	x, y, z := s.Pos().X, s.Pos().Y, s.PosZ()
	r, h := s.CollisionRadius(), s.CollisionHeight()

	// TODO: reduce smoke scale and/or opacity?

	// only spawn smoke every few frames
	fxCounter := s.EffectCounter()
	if fxCounter > 0 {
		s.SetEffectCounter(fxCounter - 1)
		return
	}

	xFx := x + randFloat(-r/2, r/2)
	yFx := y + randFloat(-r/2, r/2)
	zFx := z + randFloat(-h/2, h/2)

	smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
	g.sprites.AddEffect(smokeFx)

	fxDuration := smokeFx.AnimationDuration()
	if fxDuration > duration {
		duration = fxDuration
	}

	s.SetEffectCounter(1 + rand.Intn(3))
	return
}

func (g *Game) spawnJumpJetEffect(s *sprites.Sprite) {
	_, found := attachedJJEffects[s]
	if found {
		// do not spawn another effect
		return
	}

	jumpFx := jumpJetEffect.Clone()

	jumpFx.SetScale(s.Scale())
	jumpFx.AttachedTo = s
	jumpFx.AttachedDepth = 0.01

	g.sprites.AddEffect(jumpFx)

	// illuminate source sprite unit jump jetting
	s.SetIlluminationPeriod(5000, 0.35)

	// keep track of effect so only one is attached and can be deleted later
	attachedJJEffects[s] = jumpFx
}

func (g *Game) removeJumpJetEffect(s *sprites.Sprite) {
	jumpFx, found := attachedJJEffects[s]
	if found {
		g.sprites.DeleteEffect(jumpFx)
		delete(attachedJJEffects, s)
	}
}

func (g *Game) spawnGenericDestroyEffects(s *sprites.Sprite, spawnFires bool) (duration int) {
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

		// only spawn effects in front of sprite relative to camera position
		xFx, yFx = g.clampToCameraSpriteView(xFx, yFx, x, y)

		if spawnFires {
			fireFx := g.randFireEffect(xFx, yFx, zFx, s.Heading(), 0)
			g.sprites.AddEffect(fireFx)

			fxDuration := fireFx.AnimationDuration()
			if fxDuration > duration {
				duration = fxDuration
			}
		}

		smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.AddEffect(smokeFx)
		if !spawnFires {
			// when not spawning fires, no duration implied
			duration = 0
		}

	}
	return
}

func (g *Game) spawnPlayerDestroyEffects() (duration int) {
	s := g.player.sprite

	// limit effects to only spawn every few frames to reduce performance impact
	fxCounter := s.EffectCounter()
	if fxCounter > 0 {
		s.SetEffectCounter(fxCounter - 1)
		return
	}

	// TODO: player destruction effects based on player unit type/size
	x, y, z := s.Pos().X, s.Pos().Y, s.PosZ()
	r, h := s.CollisionRadius(), s.CollisionHeight()

	numFx := 2
	for i := 0; i < numFx; i++ {
		xFx := x + randFloat(-r, r)
		yFx := y + randFloat(-r, r)
		zFx := z + randFloat(h/4, h)

		explosionFx := g.randExplosionEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.AddEffect(explosionFx)

		smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.AddEffect(smokeFx)

		if i == 0 {
			// only play one audio track at a time
			g.audio.PlayEffectAudio(g, explosionFx)
		}

		fxDuration := explosionFx.AnimationDuration()
		if fxDuration > duration {
			duration = fxDuration
		}
	}
	s.SetEffectCounter(1 + rand.Intn(4))
	return
}

func (g *Game) spawnMechDestroyEffects(s *sprites.MechSprite) (duration int) {
	// limit effects to only spawn every few frames to reduce performance impact
	fxCounter := s.EffectCounter()
	if fxCounter > 0 {
		s.SetEffectCounter(fxCounter - 1)
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

		// only spawn effects in front of sprite relative to camera position
		xFx, yFx = g.clampToCameraSpriteView(xFx, yFx, x, y)

		explosionFx := g.randExplosionEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.AddEffect(explosionFx)

		smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.AddEffect(smokeFx)

		if i == 0 {
			// only play one audio track at a time
			g.audio.PlayEffectAudio(g, explosionFx)
		}

		fxDuration := explosionFx.AnimationDuration()
		if fxDuration > duration {
			duration = fxDuration
		}
	}
	s.SetEffectCounter(1 + rand.Intn(2))
	return
}

func (g *Game) spawnInfantryDestroyEffects(s *sprites.InfantrySprite) (duration int) {
	x, y, z := s.Pos().X, s.Pos().Y, s.PosZ()
	r, h := s.CollisionRadius(), s.CollisionHeight()

	numFx := 4 // TODO: alter number of effects based on sprite dimensions
	for i := 0; i < numFx; i++ {
		xFx := x + randFloat(-r, r)
		yFx := y + randFloat(-r, r)
		zFx := z + randFloat(0, h)

		// only spawn effects in front of sprite relative to camera position
		xFx, yFx = g.clampToCameraSpriteView(xFx, yFx, x, y)

		bloodFx := g.randBloodEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.AddEffect(bloodFx)

		fxDuration := bloodFx.AnimationDuration()
		if fxDuration > duration {
			duration = fxDuration
		}
	}
	return
}

func (g *Game) spawnVehicleDestroyEffects(s *sprites.VehicleSprite) (duration int) {
	x, y, z := s.Pos().X, s.Pos().Y, s.PosZ()
	r, h := s.CollisionRadius(), s.CollisionHeight()

	numFx := 5 // TODO: alter number of effects based on sprite dimensions
	for i := 0; i < numFx; i++ {
		xFx := x + randFloat(-r, r)
		yFx := y + randFloat(-r, r)
		zFx := z + randFloat(0, h)

		// only spawn effects in front of sprite relative to camera position
		xFx, yFx = g.clampToCameraSpriteView(xFx, yFx, x, y)

		explosionFx := g.randExplosionEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.AddEffect(explosionFx)

		smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.AddEffect(smokeFx)

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

func (g *Game) spawnVTOLDestroyEffects(s *sprites.VTOLSprite, spawnExplosions bool) (duration int) {
	// only spawn smoke every few frames
	fxCounter := s.EffectCounter()

	numFx := 1 // TODO: alter number of effects based on sprite dimensions
	if spawnExplosions {
		numFx = 5
	}

	x, y, z := s.Pos().X, s.Pos().Y, s.PosZ()
	r, h := s.CollisionRadius(), s.CollisionHeight()

	for i := 0; i < numFx; i++ {
		xFx := x + randFloat(-r/2, r/2)
		yFx := y + randFloat(-r/2, r/2)
		zFx := z + randFloat(-h/2, h/2)

		// only spawn effects in front of sprite relative to camera position
		xFx, yFx = g.clampToCameraSpriteView(xFx, yFx, x, y)

		if spawnExplosions {
			explosionFx := g.randExplosionEffect(xFx, yFx, zFx, s.Heading(), 0)
			g.sprites.AddEffect(explosionFx)
			if i == 0 || i == numFx/2 {
				// only play two audio tracks for now since they are played at once
				g.audio.PlayEffectAudio(g, explosionFx)
			}

			fxDuration := explosionFx.AnimationDuration()
			if fxDuration > duration {
				duration = fxDuration
			}
		}

		if fxCounter == 0 {
			smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
			g.sprites.AddEffect(smokeFx)
			if !spawnExplosions {
				fxDuration := smokeFx.AnimationDuration()
				if fxDuration > duration {
					duration = fxDuration
				}
			}
		}
	}

	if fxCounter > 0 {
		s.SetEffectCounter(fxCounter - 1)
	} else {
		s.SetEffectCounter(1 + rand.Intn(3))
	}

	return
}

func (g *Game) spawnEmplacementDestroyEffects(s *sprites.EmplacementSprite) (duration int) {
	x, y, z := s.Pos().X, s.Pos().Y, s.PosZ()
	r, h := s.CollisionRadius(), s.CollisionHeight()

	numFx := 5 // TODO: alter number of effects based on sprite dimensions
	for i := 0; i < numFx; i++ {
		xFx := x + randFloat(-r/2, r/2)
		yFx := y + randFloat(-r/2, r/2)
		zFx := z + randFloat(h/12, h)

		// only spawn effects in front of sprite relative to camera position
		xFx, yFx = g.clampToCameraSpriteView(xFx, yFx, x, y)

		explosionFx := g.randExplosionEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.AddEffect(explosionFx)

		smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.AddEffect(smokeFx)

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
		g.sprites.AddEffect(fireFx)

		smokeFx := g.randSmokeEffect(xFx, yFx, zFx, s.Heading(), 0)
		g.sprites.AddEffect(smokeFx)

		fxDuration := fireFx.AnimationDuration()
		if fxDuration > duration {
			duration = fxDuration
		}
	}

	return
}

func (g *Game) randBloodEffect(x, y, z, angle, pitch float64) *sprites.EffectSprite {
	// return random blood effect
	randKey := effects.RandBloodKey()
	e := bloodEffects[randKey].Clone()
	e.SetPos(&geom.Vector2{X: x, Y: y})
	e.SetPosZ(z)
	return e
}

func (g *Game) randExplosionEffect(x, y, z, angle, pitch float64) *sprites.EffectSprite {
	// return random explosion effect
	randKey := effects.RandExplosionKey()
	e := explosionEffects[randKey].Clone()
	e.SetPos(&geom.Vector2{X: x, Y: y})
	e.SetPosZ(z)

	// TODO: give small negative Z velocity so it falls with the unit being destroyed?
	return e
}

func (g *Game) randFireEffect(x, y, z, angle, pitch float64) *sprites.EffectSprite {
	// return random fire effect
	randKey := effects.RandFireKey()
	e := fireEffects[randKey].Clone()
	e.SetPos(&geom.Vector2{X: x, Y: y})
	e.SetPosZ(z)
	return e
}

func (g *Game) randSmokeEffect(x, y, z, angle, pitch float64) *sprites.EffectSprite {
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
