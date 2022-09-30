package model

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
)

type WeaponType int

const (
	ENERGY WeaponType = iota
	BALLISTIC
	MISSILE
)

type Weapon interface {
	Name() string
	ShortName() string
	Type() WeaponType
	Tonnage() float64
	Damage() float64
	Heat() float64
	Distance() float64
	Velocity() float64
	Cooldown() float64
	Offset() *geom.Vector2
	SpawnProjectile(x, y, z, angle, pitch float64, spawnedBy Entity) *Projectile
	Parent() Entity
}

type EnergyWeapon struct {
	Resource   *ModelEnergyWeaponResource
	name       string
	short      string
	tonnage    float64
	damage     float64
	heat       float64
	distance   float64
	velocity   float64
	cooldown   float64
	offset     *geom.Vector2
	projectile Projectile
	parent     Entity
}

func NewEnergyWeapon(r *ModelEnergyWeaponResource, collisionRadius, collisionHeight float64, offset *geom.Vector2, parent Entity) *EnergyWeapon {
	w := &EnergyWeapon{
		Resource: r,
		name:     r.Name,
		short:    r.ShortName,
		tonnage:  r.Tonnage,
		damage:   r.Damage,
		heat:     r.Heat,
		distance: r.Distance,
		velocity: r.Velocity,
		cooldown: r.Cooldown,
		offset:   offset,
		parent:   parent,
	}
	p := NewProjectile(r.Projectile, w.damage, w.velocity, r.Distance, collisionRadius, collisionHeight, parent)
	w.projectile = *p
	return w
}

func (w *EnergyWeapon) SpawnProjectile(x, y, z, angle, pitch float64, spawnedBy Entity) *Projectile {
	pSpawn := w.projectile.Clone()

	pSpawn.SetPos(&geom.Vector2{X: x, Y: y})
	pSpawn.SetPosZ(z)
	pSpawn.SetAngle(angle)
	pSpawn.SetPitch(pitch)

	// convert velocity from distance/second to distance per tick
	pSpawn.SetVelocity(w.velocity / float64(ebiten.MaxTPS()))

	// keep track of what spawned it
	pSpawn.SetParent(spawnedBy)

	return pSpawn
}

func (w *EnergyWeapon) Name() string {
	return w.name
}

func (w *EnergyWeapon) ShortName() string {
	return w.short
}

func (w *EnergyWeapon) Type() WeaponType {
	return ENERGY
}

func (w *EnergyWeapon) Tonnage() float64 {
	return w.tonnage
}

func (w *EnergyWeapon) Damage() float64 {
	return w.damage
}

func (w *EnergyWeapon) Heat() float64 {
	return w.heat
}

func (w *EnergyWeapon) Distance() float64 {
	return w.distance
}

func (w *EnergyWeapon) Velocity() float64 {
	return w.velocity
}

func (w *EnergyWeapon) Cooldown() float64 {
	return w.cooldown
}

func (w *EnergyWeapon) Offset() *geom.Vector2 {
	return w.offset
}

func (w *EnergyWeapon) Parent() Entity {
	return w.parent
}
