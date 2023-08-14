package model

import (
	"path/filepath"
	"strings"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/jinzhu/copier"
)

type BallisticWeapon struct {
	Resource        *ModelBallisticWeaponResource
	name            string
	short           string
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

func NewBallisticWeapon(r *ModelBallisticWeaponResource, collisionRadius, collisionHeight float64, offset *geom.Vector2, parent Entity) (*BallisticWeapon, Projectile) {
	w := &BallisticWeapon{
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

	// convert velocity from meters/second to unit distance per tick
	pVelocity := (w.velocity / METERS_PER_UNIT) * SECONDS_PER_TICK

	// convert distance and velocity to number of ticks for lifespan
	pLifespan := 2 * w.distance * (1 / w.velocity) * TICKS_PER_SECOND

	if w.extremeDistance == 0 {
		w.extremeDistance = 2 * w.distance
	}
	pExtreme := w.extremeDistance * (1 / w.velocity) * TICKS_PER_SECOND

	pDamage := w.damage
	if w.ProjectileCount() > 1 {
		// damage per projectile is divided from the total weapon damage
		pDamage /= float64(w.ProjectileCount())
	}

	if len(r.Audio) > 0 {
		w.audio = filepath.Join("audio/sfx/weapons", r.Audio)
	}

	p := *NewProjectile(r.Projectile, pDamage, pVelocity, pLifespan, pExtreme, collisionRadius, collisionHeight)
	w.projectile = p
	return w, p
}

func (w *BallisticWeapon) Clone() Weapon {
	wClone := &BallisticWeapon{}
	pClone := w.projectile.Clone().(*Projectile)

	copier.Copy(wClone, w)
	w.projectile = *pClone

	return wClone
}

func (w *BallisticWeapon) ProjectileCount() int {
	return w.Resource.ProjectileCount
}

func (w *BallisticWeapon) ProjectileDelay() float64 {
	return w.Resource.ProjectileDelay
}

func (w *BallisticWeapon) SpawnProjectileToward(target *geom3d.Vector3, spawnedBy Unit) *Projectile {
	wPos := WeaponPosition3D(spawnedBy, w.offset.X, w.offset.Y)
	angle, pitch := HeadingPitchTowardPoint3D(wPos, target)
	return w.SpawnProjectile(angle, pitch, spawnedBy)
}

func (w *BallisticWeapon) SpawnProjectile(angle, pitch float64, spawnedBy Unit) *Projectile {
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

	return pSpawn
}

func (w *BallisticWeapon) Name() string {
	return w.name
}

func (w *BallisticWeapon) ShortName() string {
	return w.short
}

func (w *BallisticWeapon) Type() WeaponType {
	return BALLISTIC
}

func (w *BallisticWeapon) Classification() WeaponClassification {
	return w.classification
}

func (w *BallisticWeapon) loadClassification() {
	s := strings.ToLower(w.short)
	switch {
	case strings.Contains(s, "ac"):
		w.classification = BALLISTIC_AUTOCANNON
	case strings.Contains(s, "gauss"):
		w.classification = BALLISTIC_GAUSS
	case strings.Contains(s, "mg"):
		w.classification = BALLISTIC_MACHINEGUN
	default:
		w.classification = WEAPON_CLASS_UNDEFINED
	}
}

func (w *BallisticWeapon) Tech() TechBase {
	return w.tech
}

func (w *BallisticWeapon) Tonnage() float64 {
	return w.tonnage
}

func (w *BallisticWeapon) Damage() float64 {
	return w.damage
}

func (w *BallisticWeapon) Heat() float64 {
	return w.heat
}

func (w *BallisticWeapon) Distance() float64 {
	return w.distance
}

func (w *BallisticWeapon) Velocity() float64 {
	return w.velocity
}

func (w *BallisticWeapon) Cooldown() float64 {
	return w.cooldown
}

func (w *BallisticWeapon) MaxCooldown() float64 {
	return w.Resource.Cooldown
}

func (w *BallisticWeapon) DecreaseCooldown(decreaseBy float64) {
	if w.cooldown > 0 && decreaseBy > 0 {
		w.cooldown -= decreaseBy
	}
	if w.cooldown < 0 {
		w.cooldown = 0
	}
}

func (w *BallisticWeapon) TriggerCooldown() {
	w.cooldown = w.MaxCooldown()
}

func (w *BallisticWeapon) Offset() *geom.Vector2 {
	return w.offset
}

func (w *BallisticWeapon) Audio() string {
	return w.audio
}

func (w *BallisticWeapon) Parent() Entity {
	return w.parent
}
