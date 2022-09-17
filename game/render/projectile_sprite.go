package render

import (
	"image/color"

	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type ProjectileSprite struct {
	*Sprite
	ImpactEffect EffectSprite
}

func NewProjectile(
	projectile *model.Projectile, scale float64, img *ebiten.Image, mapColor color.RGBA,
) *ProjectileSprite {
	p := &ProjectileSprite{
		Sprite:       NewSprite(projectile, scale, img, mapColor),
		ImpactEffect: EffectSprite{},
	}

	return p
}

func NewAnimatedProjectile(
	projectile *model.Projectile, scale float64, img *ebiten.Image, mapColor color.RGBA, columns, rows, animationRate int,
) *ProjectileSprite {
	p := &ProjectileSprite{
		Sprite:       NewAnimatedSprite(projectile, scale, img, mapColor, columns, rows, animationRate),
		ImpactEffect: EffectSprite{},
	}

	return p
}

func (p *ProjectileSprite) Clone() *ProjectileSprite {
	pClone := &ProjectileSprite{}
	sClone := &Sprite{}
	eClone := &model.Projectile{}

	copier.Copy(pClone, p)
	copier.Copy(sClone, p.Sprite)
	copier.Copy(eClone, p.Entity)

	pClone.Sprite = sClone
	pClone.Sprite.Entity = eClone

	return pClone
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

func (p *ProjectileSprite) ZeroLifespan() {
	p.Entity.(*model.Projectile).ZeroLifespan()
}

func (p *ProjectileSprite) SpawnProjectile(x, y, z, angle, pitch, velocity float64, spawnedBy model.Entity) *ProjectileSprite {
	pSpawn := p.Clone()

	pSpawn.SetPos(&geom.Vector2{X: x, Y: y})
	pSpawn.SetPosZ(z)
	pSpawn.SetAngle(angle)
	pSpawn.SetPitch(pitch)

	// convert velocity from distance/second to distance per tick
	pSpawn.SetVelocity(velocity / float64(ebiten.MaxTPS()))

	// keep track of what spawned it
	pSpawn.SetParent(spawnedBy)

	return pSpawn
}

func (p *ProjectileSprite) SpawnEffect(x, y, z, angle, pitch float64) *EffectSprite {
	e := p.ImpactEffect.Clone()
	e.SetPos(&geom.Vector2{X: x, Y: y})
	e.SetPosZ(z)
	e.SetAngle(angle)
	e.SetPitch(pitch)

	// keep track of what spawned it
	e.SetParent(p.Parent())

	return e
}
