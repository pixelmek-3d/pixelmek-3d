package model

import (
	"image/color"
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jinzhu/copier"
)

type Projectile struct {
	*Sprite
	Lifespan     float64
	Damage       float64 // TODO: separate out non-visual game model
	ImpactEffect Effect
}

func NewProjectile(
	x, y, scale, lifespan float64, img *ebiten.Image, mapColor color.RGBA, anchor raycaster.SpriteAnchor, collisionRadius, damage float64,
) *Projectile {
	if lifespan < 0 {
		lifespan = math.MaxFloat64
	}
	p := &Projectile{
		Sprite:       NewSprite(x, y, scale, img, mapColor, anchor, collisionRadius),
		Lifespan:     lifespan,
		Damage:       damage,
		ImpactEffect: Effect{},
	}

	return p
}

func NewAnimatedProjectile(
	x, y, scale, lifespan float64, img *ebiten.Image, mapColor color.RGBA, columns, rows, animationRate int,
	anchor raycaster.SpriteAnchor, collisionRadius, damage float64,
) *Projectile {
	if lifespan < 0 {
		lifespan = math.MaxFloat64
	}
	p := &Projectile{
		Sprite:       NewAnimatedSprite(x, y, scale, img, mapColor, columns, rows, animationRate, anchor, collisionRadius),
		Lifespan:     lifespan,
		Damage:       damage,
		ImpactEffect: Effect{},
	}

	return p
}

func (pSpawn *Projectile) SpawnProjectile(x, y, z, angle, pitch, velocity float64, spawnedBy *Entity) *Projectile {
	p := &Projectile{}
	s := &Sprite{}
	copier.Copy(p, pSpawn)
	copier.Copy(s, pSpawn.Sprite)

	p.Sprite = s
	p.Position = &geom.Vector2{X: x, Y: y}
	p.PositionZ = z
	p.Angle = angle
	p.Pitch = pitch

	// convert velocity from distance/second to distance per tick
	p.Velocity = velocity / float64(ebiten.MaxTPS())

	// keep track of what spawned it
	p.Parent = spawnedBy

	return p
}

func (p *Projectile) SpawnEffect(x, y, z, angle, pitch float64) *Effect {
	e := &Effect{}
	s := &Sprite{}
	copier.Copy(e, p.ImpactEffect)
	copier.Copy(s, p.ImpactEffect.Sprite)

	e.Sprite = s
	e.Position = &geom.Vector2{X: x, Y: y}
	e.PositionZ = z
	e.Angle = angle
	e.Pitch = pitch

	// keep track of what spawned it
	e.Parent = p.Parent

	return e
}
