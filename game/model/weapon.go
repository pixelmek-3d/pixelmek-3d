package model

import (
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
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
	Tech() TechBase
	Type() WeaponType
	Tonnage() float64
	Damage() float64
	Heat() float64
	Distance() float64
	Velocity() float64
	Cooldown() float64
	MaxCooldown() float64
	DecreaseCooldown(float64)
	TriggerCooldown()

	Offset() *geom.Vector2
	SpawnProjectile(x, y, z, angle, pitch float64, spawnedBy Entity) *Projectile

	Clone() Weapon
	Parent() Entity
}

type EnergyWeapon struct {
	Resource   *ModelEnergyWeaponResource
	name       string
	short      string
	tech       TechBase
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

func NewEnergyWeapon(r *ModelEnergyWeaponResource, collisionRadius, collisionHeight float64, offset *geom.Vector2, parent Entity) (*EnergyWeapon, Projectile) {
	w := &EnergyWeapon{
		Resource: r,
		name:     r.Name,
		short:    r.ShortName,
		tech:     r.Tech.TechBase,
		tonnage:  r.Tonnage,
		damage:   r.Damage,
		heat:     r.Heat,
		distance: r.Distance,
		velocity: r.Velocity,
		cooldown: 0,
		offset:   offset,
		parent:   parent,
	}

	// convert velocity from meters/second to unit distance per tick
	pVelocity := (w.velocity / METERS_PER_UNIT) * SECONDS_PER_TICK

	// convert distance and velocity to number of ticks for lifespan
	pLifespan := w.distance * (1 / w.velocity) * TICKS_PER_SECOND

	p := *NewProjectile(r.Projectile, w.damage, pVelocity, pLifespan, collisionRadius, collisionHeight, parent)
	w.projectile = p
	return w, p
}

func (w *EnergyWeapon) Clone() Weapon {
	wClone := &EnergyWeapon{}
	pClone := w.projectile.Clone().(*Projectile)

	copier.Copy(wClone, w)
	w.projectile = *pClone

	return wClone
}

func (w *EnergyWeapon) SpawnProjectile(x, y, z, angle, pitch float64, spawnedBy Entity) *Projectile {
	pSpawn := w.projectile.Clone().(*Projectile)

	pSpawn.SetPos(&geom.Vector2{X: x, Y: y})
	pSpawn.SetPosZ(z)
	pSpawn.SetAngle(angle)
	pSpawn.SetPitch(pitch)

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

func (w *EnergyWeapon) Tech() TechBase {
	return w.tech
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

func (w *EnergyWeapon) MaxCooldown() float64 {
	return w.Resource.Cooldown
}

func (w *EnergyWeapon) DecreaseCooldown(decreaseBy float64) {
	if w.cooldown > 0 && decreaseBy > 0 {
		w.cooldown -= decreaseBy
	}
	if w.cooldown < 0 {
		w.cooldown = 0
	}
}

func (w *EnergyWeapon) TriggerCooldown() {
	w.cooldown = w.MaxCooldown()
}

func (w *EnergyWeapon) Offset() *geom.Vector2 {
	return w.offset
}

func (w *EnergyWeapon) Parent() Entity {
	return w.parent
}
