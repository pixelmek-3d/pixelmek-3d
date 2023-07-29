package model

import (
	"math"
	"path/filepath"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/jinzhu/copier"
)

type MissileWeapon struct {
	Resource        *ModelMissileWeaponResource
	name            string
	short           string
	tech            TechBase
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

	missileTube        int
	missileTubeOffset  []*geom.Vector2
	lockOnLockRequired bool
	lockOnTurnRate     float64
	lockOnGroupRadius  float64
}

func NewMissileWeapon(r *ModelMissileWeaponResource, collisionRadius, collisionHeight float64, offset *geom.Vector2, onePxOffset *geom.Vector2, parent Entity) (*MissileWeapon, Projectile) {
	w := &MissileWeapon{
		Resource:          r,
		name:              r.Name,
		short:             r.ShortName,
		tech:              r.Tech.TechBase,
		tonnage:           r.Tonnage,
		damage:            r.Damage,
		heat:              r.Heat,
		distance:          r.Distance,
		extremeDistance:   r.ExtremeDistance,
		velocity:          r.Velocity,
		cooldown:          0,
		offset:            offset,
		missileTube:       0,
		missileTubeOffset: make([]*geom.Vector2, r.ProjectileCount),
		parent:            parent,
	}

	if r.LockOn != nil {
		w.lockOnLockRequired = r.LockOn.LockRequired
		w.lockOnTurnRate = geom.Radians(r.LockOn.TurnRate) / TICKS_PER_SECOND
		w.lockOnGroupRadius = r.LockOn.GroupRadius
	}

	// convert velocity from meters/second to unit distance per tick
	pVelocity := (w.velocity / METERS_PER_UNIT) * SECONDS_PER_TICK

	// convert distance and velocity to number of ticks for lifespan
	pLifespan := 2 * w.distance * (1 / w.velocity) * TICKS_PER_SECOND

	pDamage := w.damage
	if w.ProjectileCount() > 1 {
		// damage per projectile is divided from the total weapon damage
		pDamage /= float64(w.ProjectileCount())
	}

	if w.extremeDistance == 0 {
		w.extremeDistance = 2 * w.distance
	}
	pExtreme := w.extremeDistance * (1 / w.velocity) * TICKS_PER_SECOND

	// based on the number of tubes, create offsets for each missile so they spawn from slightly different positions
	w.loadMissileTubes(onePxOffset)

	if len(r.Audio) > 0 {
		w.audio = filepath.Join("audio/sfx/weapons", r.Audio)
	}

	p := *NewProjectile(r.Projectile, pDamage, pVelocity, pLifespan, pExtreme, collisionRadius, collisionHeight)
	w.projectile = p
	return w, p
}

func (w *MissileWeapon) loadMissileTubes(onePxOffset *geom.Vector2) {
	if w.ProjectileCount() <= 1 {
		return
	}

	// distribute missile tubes somewhat evenly around the weapon offset center into rows
	tubeCount := w.ProjectileCount()
	tubeRows := 0
	switch {
	case tubeCount == 2:
		tubeRows = 1
	case tubeCount >= 4 && tubeCount < 10:
		tubeRows = 2
	case tubeCount >= 19 && tubeCount < 20:
		tubeRows = 3
	case tubeCount >= 20:
		tubeRows = 4
	default:
		tubeRows = int(math.Sqrt(float64(tubeCount)))
	}

	if tubeRows == 0 {
		return
	}

	// adjust x/y position used to approximate middle of missile tube box
	tubeCols := int(math.Ceil(float64(tubeCount) / float64(tubeRows)))
	adjustX, adjustY := float64(tubeCols)/2, float64(tubeRows)/2

	tubeIndex := 0
	for row := 0; row < tubeRows; row++ {
		for col := 0; col < tubeCols; col++ {

			tubeOffX, tubeOffY := (float64(col)-adjustX)*onePxOffset.X, (float64(row)-adjustY)*onePxOffset.Y
			w.missileTubeOffset[tubeIndex] = &geom.Vector2{X: tubeOffX, Y: tubeOffY}

			tubeIndex++
			if tubeIndex >= tubeCount {
				return
			}
		}
	}
}

func (w *MissileWeapon) getMissileTubeOffset() *geom.Vector2 {
	if w.missileTube >= len(w.missileTubeOffset) {
		// "reload" tubes
		w.missileTube = 0
	}
	return w.missileTubeOffset[w.missileTube]
}

func (w *MissileWeapon) IsLockOn() bool {
	return w.Resource.LockOn != nil
}

func (w *MissileWeapon) IsLockOnLockRequired() bool {
	return w.lockOnLockRequired
}

func (w *MissileWeapon) LockOnTurnRate() float64 {
	return w.lockOnTurnRate
}

func (w *MissileWeapon) LockOnGroupRadius() float64 {
	return w.lockOnGroupRadius
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

func (w *MissileWeapon) SpawnProjectileToward(target *geom3d.Vector3, spawnedBy Unit) *Projectile {
	wPos := WeaponPosition3D(spawnedBy, w.offset.X, w.offset.Y)
	angle, pitch := HeadingPitchTowardPoint3D(wPos, target)
	return w.SpawnProjectile(angle, pitch, spawnedBy)
}

func (w *MissileWeapon) SpawnProjectile(angle, pitch float64, spawnedBy Unit) *Projectile {
	pSpawn := w.projectile.Clone().(*Projectile)

	// add weapon position offset based on where it is mounted, along with missile tube offset of current missile
	tubeOffset := w.getMissileTubeOffset()
	wPos := WeaponPosition3D(spawnedBy, w.offset.X+tubeOffset.X, w.offset.Y+tubeOffset.Y)

	pSpawn.SetPos(&geom.Vector2{X: wPos.X, Y: wPos.Y})
	pSpawn.SetPosZ(wPos.Z)
	pSpawn.SetHeading(angle)
	pSpawn.SetPitch(pitch)

	// keep track of what spawned it
	pSpawn.SetParent(spawnedBy)
	pSpawn.SetWeapon(w)

	w.missileTube++

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

func (w *MissileWeapon) Audio() string {
	return w.audio
}

func (w *MissileWeapon) Parent() Entity {
	return w.parent
}
