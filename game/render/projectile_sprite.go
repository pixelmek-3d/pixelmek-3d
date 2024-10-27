package render

import (
	"math"
	"math/rand"
	"path"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type ProjectileSprite struct {
	*Sprite
	ImpactAudioFiles []string
	ImpactEffect     EffectSprite
	Projectile       *model.Projectile
}

func NewAnimatedProjectile(
	projectile *model.Projectile, scale float64, img *ebiten.Image, impactEffect EffectSprite, impactAudioFiles []string,
) *ProjectileSprite {
	var p *Sprite
	sheet := projectile.Resource.ImageSheet

	if sheet == nil {
		p = NewSprite(
			projectile, scale, img,
		)
	} else {
		p = NewAnimatedSprite(projectile, scale, img, sheet.Columns, sheet.Rows, sheet.AnimationRate)
		if len(sheet.AngleFacingRow) > 0 {
			facingMap := make(map[float64]int, len(sheet.AngleFacingRow))
			for degrees, index := range sheet.AngleFacingRow {
				rads := geom.Radians(degrees)
				facingMap[rads] = index
			}
			p.SetTextureFacingMap(facingMap)
		}
	}

	// projectiles cannot be focused upon by player reticle
	p.focusable = false

	// projectiles self illuminate so they do not get dimmed in night conditions
	p.illumination = 5000

	s := &ProjectileSprite{
		Sprite:           p,
		ImpactAudioFiles: make([]string, len(impactAudioFiles)),
		ImpactEffect:     impactEffect,
		Projectile:       projectile,
	}

	for _, audioFile := range impactAudioFiles {
		if len(audioFile) > 0 {
			audioFile = path.Join("audio/sfx/impacts", audioFile)
			s.ImpactAudioFiles = append(s.ImpactAudioFiles, audioFile)
		}
	}

	return s
}

func (p *ProjectileSprite) Clone() *ProjectileSprite {
	pClone := &ProjectileSprite{}
	sClone := &Sprite{}
	eClone := p.Entity.Clone()

	copier.Copy(pClone, p)
	copier.Copy(sClone, p.Sprite)

	pClone.Sprite = sClone
	pClone.Sprite.Entity = eClone

	return pClone
}

func (p *ProjectileSprite) ImpactAudio() string {
	numAudioFiles := len(p.ImpactAudioFiles)
	if numAudioFiles == 1 {
		return p.ImpactAudioFiles[0]
	} else if numAudioFiles > 1 {
		return p.ImpactAudioFiles[rand.Intn(numAudioFiles)]
	}
	return ""
}

func (p *ProjectileSprite) Damage() float64 {
	return p.Entity.(*model.Projectile).Damage()
}

func (p *ProjectileSprite) Lifespan() float64 {
	return p.Entity.(*model.Projectile).Lifespan()
}

func (p *ProjectileSprite) DecreaseLifespan(decreaseBy float64) float64 {
	return p.Entity.(*model.Projectile).DecreaseLifespan(decreaseBy)
}

func (p *ProjectileSprite) Destroy() {
	p.Entity.(*model.Projectile).Destroy()
}

func (p *ProjectileSprite) SpawnEffect(x, y, z, angle, pitch float64) *EffectSprite {
	e := p.ImpactEffect.Clone()
	e.SetPos(&geom.Vector2{X: x, Y: y})
	e.SetPosZ(z)
	e.SetHeading(angle)
	e.SetPitch(pitch)

	// keep track of what spawned it
	e.SetParent(p.Parent())

	return e
}

func (s *ProjectileSprite) Update(camPos *geom.Vector2) {
	if s.Parent() != nil && model.IsEntityUnit(s.Parent()) && model.EntityUnit(s.Parent()).IsPlayer() {
		// Projectiles spawned by player weapons in the arms could initially use an angled facing
		// instead of directly behind until further away. Facing angle override is used for first several
		// frames to force the angle viewed as directly behind the projectile from player perspective.
		if s.loopCounter < 7 {
			// 180 degrees (Pi) forces perspective of directly behind projectile as travels away from player
			s.camFacingOverride = &facingAngleOverride{angle: math.Pi}
		} else {
			s.camFacingOverride = nil
		}
	}

	s.Sprite.Update(camPos)
}
