package model

import (
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type MissileWeapon struct {
	Resource   *ModelMissileWeaponResource
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

func NewMissileWeapon(r *ModelMissileWeaponResource, collisionRadius, collisionHeight float64, offset *geom.Vector2, parent Entity) (*MissileWeapon, Projectile) {
	w := &MissileWeapon{
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

	pDamage := w.damage
	if w.ProjectileCount() > 1 {
		// damage per projectile is divided from the total weapon damage
		pDamage /= float64(w.ProjectileCount())
	}

	p := *NewProjectile(r.Projectile, pDamage, pVelocity, pLifespan, collisionRadius, collisionHeight, parent)
	w.projectile = p
	return w, p
}

func (w *MissileWeapon) Clone() Weapon {
	wClone := &MissileWeapon{}
	pClone := w.projectile.Clone().(*Projectile)

	copier.Copy(wClone, w)
	w.projectile = *pClone

	return wClone
}

func (w *MissileWeapon) ProjectileCount() int {
	return w.Resource.ProjectileCount
}

func (w *MissileWeapon) ProjectileDelay() float64 {
	return w.Resource.ProjectileDelay
}

func (w *MissileWeapon) SpawnProjectile(angle, pitch float64, spawnedBy Entity) *Projectile {
	pSpawn := w.projectile.Clone().(*Projectile)

	// add weapon position offset based on where it is mounted
	x, y, z := WeaponPosition3D(spawnedBy, w.offset.X, w.offset.Y)

	pSpawn.SetPos(&geom.Vector2{X: x, Y: y})
	pSpawn.SetPosZ(z)
	pSpawn.SetAngle(angle)
	pSpawn.SetPitch(pitch)

	// keep track of what spawned it
	pSpawn.SetParent(spawnedBy)

	return pSpawn
}

func (w *MissileWeapon) Name() string {
	return w.name
}

func (w *MissileWeapon) ShortName() string {
	return w.short
}

func (w *MissileWeapon) Type() WeaponType {
	return MISSILE
}

func (w *MissileWeapon) Tech() TechBase {
	return w.tech
}

func (w *MissileWeapon) Tonnage() float64 {
	return w.tonnage
}

func (w *MissileWeapon) Damage() float64 {
	return w.damage
}

func (w *MissileWeapon) Heat() float64 {
	return w.heat
}

func (w *MissileWeapon) Distance() float64 {
	return w.distance
}

func (w *MissileWeapon) Velocity() float64 {
	return w.velocity
}

func (w *MissileWeapon) Cooldown() float64 {
	return w.cooldown
}

func (w *MissileWeapon) MaxCooldown() float64 {
	return w.Resource.Cooldown
}

func (w *MissileWeapon) DecreaseCooldown(decreaseBy float64) {
	if w.cooldown > 0 && decreaseBy > 0 {
		w.cooldown -= decreaseBy
	}
	if w.cooldown < 0 {
		w.cooldown = 0
	}
}

func (w *MissileWeapon) TriggerCooldown() {
	w.cooldown = w.MaxCooldown()
}

func (w *MissileWeapon) Offset() *geom.Vector2 {
	return w.offset
}

func (w *MissileWeapon) Parent() Entity {
	return w.parent
}
