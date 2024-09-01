package model

import (
	"path"
	"strings"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/jinzhu/copier"
)

type EnergyWeapon struct {
	Resource        *ModelEnergyWeaponResource
	name            string
	short           string
	summary         string
	tech            TechBase
	classification  WeaponClassification
	tonnage         float64
	damage          float64
	heat            float64
	distance        float64
	extremeDistance float64
	velocity        float64
	cooldown        float64
	offset          *geom.Vector2
	projectile      Projectile
	audio           string
	parent          Entity
}

func NewEnergyWeapon(r *ModelEnergyWeaponResource, collisionRadius, collisionHeight float64, offset *geom.Vector2, parent Entity) (*EnergyWeapon, Projectile) {
	w := &EnergyWeapon{
		Resource:        r,
		name:            r.Name,
		short:           r.ShortName,
		tech:            r.Tech.TechBase,
		tonnage:         r.Tonnage,
		damage:          r.Damage,
		heat:            r.Heat,
		distance:        r.Distance,
		extremeDistance: r.ExtremeDistance,
		velocity:        r.Velocity,
		cooldown:        0,
		offset:          offset,
		parent:          parent,
	}

	// load general classification of weapon programmatically
	w.loadClassification()
	w.summary = weaponSummary(w)

	// convert velocity from meters/second to unit distance per tick
	pVelocity := (w.velocity / METERS_PER_UNIT) * SECONDS_PER_TICK

	// convert distance and velocity to number of ticks for lifespan
	pLifespan := w.distance * (1 / w.velocity) * TICKS_PER_SECOND

	if w.extremeDistance == 0 {
		w.extremeDistance = 2 * w.distance
	}
	// subtract normal distance lifespan since it gets added on once it hits extreme ranges
	pExtreme := (w.extremeDistance * (1 / w.velocity) * TICKS_PER_SECOND) - pLifespan

	pDamage := w.damage
	if w.ProjectileCount() > 1 {
		// damage per projectile is divided from the total weapon damage
		pDamage /= float64(w.ProjectileCount())
	}

	if len(r.Audio) > 0 {
		w.audio = path.Join("audio/sfx/weapons", r.Audio)
	}

	p := *NewProjectile(r.Projectile, pDamage, pVelocity, pLifespan, pExtreme, collisionRadius, collisionHeight)
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

func (w *EnergyWeapon) Summary() string {
	return w.summary
}

func (w *EnergyWeapon) AmmoPerTon() int {
	// no ammo needed for energy weapons
	return 0
}

func (w *EnergyWeapon) AmmoBin() *AmmoBin {
	return nil
}

func (w *EnergyWeapon) SetAmmoBin(ammoBin *AmmoBin) {}

func (w *EnergyWeapon) Projectile() Projectile {
	return w.projectile
}

func (w *EnergyWeapon) ProjectileCount() int {
	return w.Resource.ProjectileCount
}

func (w *EnergyWeapon) ProjectileDelay() float64 {
	return w.Resource.ProjectileDelay
}

func (w *EnergyWeapon) ProjectileSpread() float64 {
	// no spread for energy weapons
	return 0
}

func (w *EnergyWeapon) SpawnProjectileToward(target *geom3d.Vector3, spawnedBy Unit) *Projectile {
	wPos := WeaponPosition3D(spawnedBy, w.offset.X, w.offset.Y)
	angle, pitch := HeadingPitchTowardPoint3D(wPos, target)
	return w.SpawnProjectile(angle, pitch, spawnedBy)
}

func (w *EnergyWeapon) SpawnProjectile(angle, pitch float64, spawnedBy Unit) *Projectile {
	pSpawn := w.projectile.Clone().(*Projectile)

	// add weapon position offset based on where it is mounted
	wPos := WeaponPosition3D(spawnedBy, w.offset.X, w.offset.Y)

	pSpawn.SetPos(&geom.Vector2{X: wPos.X, Y: wPos.Y})
	pSpawn.SetPosZ(wPos.Z)
	pSpawn.SetHeading(angle)
	pSpawn.SetPitch(pitch)

	// keep track of what spawned it
	pSpawn.SetParent(spawnedBy)
	pSpawn.SetWeapon(w)
	pSpawn.SetTeam(spawnedBy.Team())

	return pSpawn
}

func (w *EnergyWeapon) File() string {
	return w.Resource.File
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

func (w *EnergyWeapon) Classification() WeaponClassification {
	return w.classification
}

func (w *EnergyWeapon) loadClassification() {
	s := strings.ToLower(w.short)
	switch {
	case strings.Contains(s, "laser"):
		w.classification = ENERGY_LASER
	case strings.Contains(s, "ppc"):
		w.classification = ENERGY_PPC
	case strings.Contains(s, "flamer"):
		w.classification = ENERGY_FLAMER
	default:
		w.classification = WEAPON_CLASS_UNDEFINED
	}
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

func (w *EnergyWeapon) Audio() string {
	return w.audio
}

func (w *EnergyWeapon) Parent() Entity {
	return w.parent
}
