package render

import (
	"image/color"
	"math"

	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type ProjectileSprite struct {
	*Sprite
	Lifespan     float64
	Damage       float64 // TODO: separate out non-visual game model
	ImpactEffect EffectSprite
}

func NewProjectile(
	modelEntity model.Entity, x, y, scale, lifespan float64, img *ebiten.Image, mapColor color.RGBA, damage float64,
) *ProjectileSprite {
	if lifespan < 0 {
		lifespan = math.MaxFloat64
	}
	p := &ProjectileSprite{
		Sprite:       NewSprite(modelEntity, x, y, scale, img, mapColor),
		Lifespan:     lifespan,
		Damage:       damage,
		ImpactEffect: EffectSprite{},
	}

	return p
}

func NewAnimatedProjectile(
	modelEntity model.Entity, x, y, scale, lifespan float64, img *ebiten.Image, mapColor color.RGBA, columns, rows, animationRate int, damage float64,
) *ProjectileSprite {
	if lifespan < 0 {
		lifespan = math.MaxFloat64
	}
	p := &ProjectileSprite{
		Sprite:       NewAnimatedSprite(modelEntity, x, y, scale, img, mapColor, columns, rows, animationRate),
		Lifespan:     lifespan,
		Damage:       damage,
		ImpactEffect: EffectSprite{},
	}

	return p
}

func (p *ProjectileSprite) Clone() *ProjectileSprite {
	pClone := &ProjectileSprite{}
	sClone := &Sprite{}
	eClone := &model.BasicEntity{}

	copier.Copy(pClone, p)
	copier.Copy(sClone, p.Sprite)
	copier.Copy(eClone, p.Entity)

	pClone.Sprite = sClone
	pClone.Sprite.Entity = eClone

	return pClone
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
