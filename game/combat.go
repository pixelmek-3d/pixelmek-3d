package game

import (
	"fmt"
	"sync"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/sprites"
)

type ProjectileSpawn struct {
	delay      float64
	spread     float64
	weapon     model.Weapon
	parent     model.Entity
	sfxEnabled bool
}

func (g *Game) initCombatVariables() {
	g.delayedProjectiles = make(map[*ProjectileSpawn]struct{}, 256)
}

func NewProjectileSpawn(weapon model.Weapon, parent model.Entity) *ProjectileSpawn {
	p := &ProjectileSpawn{
		weapon:     weapon,
		parent:     parent,
		sfxEnabled: true,
	}
	return p
}

func NewDelayedProjectileSpawn(delay, spread float64, weapon model.Weapon, parent model.Entity, sfxEnabled bool) *ProjectileSpawn {
	p := &ProjectileSpawn{
		delay:      delay,
		spread:     spread,
		weapon:     weapon,
		parent:     parent,
		sfxEnabled: sfxEnabled,
	}
	return p
}

func destroyEntity(e model.Entity) {
	if e != nil {
		e.SetStructurePoints(0)
	}
}

// firePlayerWeapon fires currently selected player weapon/weapon group or input weapon group
func (g *Game) firePlayerWeapon(weaponGroupFire int) bool {
	// weapons test from model
	if g.player.IsDestroyed() {
		return false
	}
	if g.player.Powered() != model.POWER_ON {
		// TODO: when shutdown, show weapons as disabled and disallow cycling weapons
		return false
	}
	armament := g.player.Armament()
	if len(armament) == 0 {
		return false
	}

	weaponsFired := false

	if weaponGroupFire < 0 && g.player.fireMode == model.GROUP_FIRE {
		// indicate firing current selected weapon group
		weaponGroupFire = int(g.player.selectedGroup)
	}

	// if a weapon with lock required tries but cannot fire make a sound
	isWeaponWithLockRequiredNotFired := false
	// if a weapon with no ammo tries to fire make an empty click sound
	isWeaponWithNoAmmoNotFired := false

	for i, weapon := range armament {
		if weapon.Cooldown() > 0 {
			continue
		}

		isWeaponSelected := (weaponGroupFire < 0 && g.player.fireMode == model.CHAIN_FIRE && i == int(g.player.selectedWeapon)) ||
			(weaponGroupFire >= 0 && model.IsWeaponInGroup(weapon, uint(weaponGroupFire), g.player.weaponGroups))
		if !isWeaponSelected {
			continue
		}

		ammoBin := weapon.AmmoBin()
		if ammoBin != nil {
			// perform ammo check to see if need to play empty click sound
			ammoCount := ammoBin.AmmoCount()
			if ammoCount == 0 {
				isWeaponWithNoAmmoNotFired = true
				continue
			}
		}

		if g.fireUnitWeapon(g.player.Unit, weapon) {
			weaponsFired = true
		} else {
			missileWeapon, isMissile := weapon.(*model.MissileWeapon)
			if isMissile && missileWeapon.IsLockOnLockRequired() {
				isWeaponWithLockRequiredNotFired = true
			}
		}
	}

	if !weaponsFired && isWeaponWithLockRequiredNotFired && !g.audio.IsButtonAudioPlaying() {
		g.audio.PlayButtonAudio(AUDIO_BUTTON_NEG)
	}

	if !weaponsFired && isWeaponWithNoAmmoNotFired && !g.audio.IsButtonAudioPlaying() {
		g.audio.PlayButtonAudio(AUDIO_CLICK_NEG)
	}

	return weaponsFired
}

func (g *Game) fireUnitWeapon(unit model.Unit, weapon model.Weapon) bool {
	if unit == nil || weapon == nil {
		return false
	}

	if weapon.Cooldown() > 0 {
		return false
	}

	ammoBin := weapon.AmmoBin()
	if ammoBin != nil {
		// perform ammo check
		ammoCount := ammoBin.AmmoCount()
		if ammoCount == 0 {
			return false
		}
	}

	weaponFired := false
	if unit.TriggerWeapon(weapon) {
		weaponFired = true

		pSpawn := NewProjectileSpawn(weapon, unit)
		projectile := g.spawnProjectile(pSpawn)
		if projectile != nil {
			// queue creation of multiple projectiles after time delay
			if weapon.ProjectileCount() > 1 {
				for i := 1; i < weapon.ProjectileCount(); i++ {
					g.queueDelayedProjectile(i, weapon, unit)
				}
			}
		}

		// consume ammo
		if ammoBin != nil {
			ammoBin.ConsumeAmmo(weapon, 1)
			//log.Debugf("[%s %s] %s: %d", unit.Name(), unit.Variant(), weapon.ShortName(), ammoBin.AmmoCount())
		}
	}

	return weaponFired
}

// updateProjectiles updates the state of all projectiles in play
func (g *Game) updateProjectiles() {
	// update and spawn projectiles on delay timers
	g.updateDelayedProjectiles()

	// perform concurrent projectile updates
	var wg sync.WaitGroup

	g.sprites.RangeByType(sprites.ProjectileSpriteType, func(k, _ interface{}) bool {
		p := k.(*sprites.ProjectileSprite)
		p.DecreaseLifespan(1)
		if p.Lifespan() <= 0 {
			g.sprites.DeleteProjectile(p)
			return true
		}

		wg.Add(1)
		go g.asyncProjectileUpdate(p, &wg)

		return true
	})

	// Update animated effects
	g.sprites.RangeByType(sprites.EffectSpriteType, func(k, _ interface{}) bool {
		e := k.(*sprites.EffectSprite)
		e.Update(g.player.CameraPosXY())
		if e.LoopCounter() >= e.LoopCount {
			g.sprites.DeleteEffect(e)
		}

		return true
	})

	wg.Wait()
}

// asyncProjectileUpdate updates the positions of a projectile in a parallel fashion
func (g *Game) asyncProjectileUpdate(p *sprites.ProjectileSprite, wg *sync.WaitGroup) {
	defer wg.Done()

	_, isEnergy := p.Projectile.Weapon().(*model.EnergyWeapon)
	missileWeapon, isMissile := p.Projectile.Weapon().(*model.MissileWeapon)

	if isMissile && p.Velocity() < p.Projectile.MaxVelocity() {
		newVelocity := geom.Clamp(p.Velocity()+p.Projectile.Acceleration(), 0, p.Projectile.MaxVelocity())
		p.SetVelocity(newVelocity)
	}

	if p.Velocity() != 0 {
		pPos := p.Pos()

		// adjust pitch and heading if is a locked missile projectile
		if isMissile && missileWeapon.IsLockOn() && !p.Projectile.InExtremeRange() {
			pUnit := p.Projectile.Parent().(model.Unit)
			target := pUnit.Target()
			if target != nil {
				tPos := target.Pos()

				// add a small amount of random offset to X/Y/Z of target line
				pOffset := p.Projectile.LockOnOffset()

				// use target collision box to determine center of target offset
				collisionOffset := 0.0
				switch target.Anchor() {
				case raycaster.AnchorBottom:
					collisionOffset = 2 * target.CollisionHeight() / 3
				case raycaster.AnchorTop:
					collisionOffset = -1 * target.CollisionHeight() / 3
				}

				tLine := &geom3d.Line3d{
					X1: pPos.X, Y1: pPos.Y, Z1: p.PosZ(),
					X2: tPos.X + pOffset.X, Y2: tPos.Y + pOffset.Y, Z2: target.PosZ() + pOffset.Z + collisionOffset,
				}
				tHeading, tPitch := tLine.Heading(), tLine.Pitch()
				if tHeading < 0 {
					tHeading += geom.Pi2
				}

				pHeading, pPitch := p.Heading(), p.Pitch()
				if pHeading < 0 {
					pHeading += geom.Pi2
				}

				// only adjust heading/pitch angle by small amount towards target
				pDelta := missileWeapon.LockOnTurnRate() * pUnit.TargetLock()

				if tHeading != pHeading {
					isCCW := model.IsBetweenRadians(pHeading, pHeading-geom.Pi, tHeading)
					if isCCW {
						tHeading = geom.Clamp(tHeading, pHeading, pHeading+pDelta)
					} else {
						tHeading = geom.Clamp(tHeading, pHeading-pDelta, pHeading)
					}
				}

				if tPitch != pPitch {
					if tPitch > pPitch {
						tPitch = geom.Clamp(tPitch, pPitch, pPitch+pDelta)
					} else {
						tPitch = geom.Clamp(tPitch, pPitch-pDelta, pPitch)
					}
				}

				p.SetHeading(tHeading)
				p.SetPitch(tPitch)
			}
		}

		trajectory := geom3d.Line3dFromAngle(pPos.X, pPos.Y, p.PosZ(), p.Heading(), p.Pitch(), p.Velocity())

		if p.Projectile.InExtremeRange() && !isEnergy {
			// make projectile trajectory start to fall (except for energy weapons)
			extremeTrajectory := &trajectory
			extremeTrajectory.Z2 -= model.GRAVITY_UNITS_PTT
			p.SetPitch(extremeTrajectory.Pitch())

			if p.Velocity() > 0 {
				// for now just using gravity as basis for air resistance to reduce velocity at extreme range
				extremeVelocity := geom.Clamp(p.Velocity()-model.GRAVITY_UNITS_PTT, 0, p.Velocity())
				p.SetVelocity(extremeVelocity)
			}
		}

		xCheck := trajectory.X2
		yCheck := trajectory.Y2
		zCheck := trajectory.Z2

		newPos, newPosZ, isCollision, collisions := g.getValidMove(p.Entity, xCheck, yCheck, zCheck, false)
		if isCollision || p.PosZ() <= 0 {
			var collisionEntity *EntityCollision
			if len(collisions) > 0 {
				// apply damage to the first sprite entity that was hit
				collisionEntity = collisions[0]
				entity := collisionEntity.entity

				damage := p.Damage()
				entity.ApplyDamage(damage)

				// if g.debug {
				// 	unit := model.EntityUnit(entity)
				// 	hp, maxHP := entity.ArmorPoints()+entity.StructurePoints(), entity.MaxArmorPoints()+entity.MaxStructurePoints()
				// 	percentHP := 100 * (hp / maxHP)

				// 	if unit == g.player.Unit {
				// 		// TODO: visual response to player being hit
				// 		log.Debugf("[player] hit for %0.1f | HP: %0.1f/%0.0f (%0.2f%%)", damage, hp, maxHP, percentHP)
				// 	} else if unit != nil {
				// 		// TODO: ui indicator for showing damage was done
				// 		log.Debugf("[%s] hit for %0.1f | HP: %0.1f/%0.0f (%0.2f%%)", unit.ID(), damage, hp, maxHP, percentHP)
				// 	}
				// }
			}

			// destroy projectile after applying damage so it can calculate dropoff if needed
			p.SetPos(newPos)
			p.SetPosZ(newPosZ)
			p.Destroy()

			// make a sprite/wall getting hit by projectile cause some visual effect
			if p.ImpactEffect.Sprite != nil {
				if collisionEntity != nil {
					// use the first collision point to place effect at
					newPos = collisionEntity.collision
				}

				effect := p.SpawnEffect(newPos.X, newPos.Y, newPosZ, p.Heading(), p.Pitch())
				g.sprites.AddEffect(effect)
			}

			// play impact effect audio
			g.audio.PlayProjectileImpactAudio(g, p)

		} else {
			p.SetPos(newPos)
			p.SetPosZ(newPosZ)

			// smoke trail for non-player ejection pod
			if p.Projectile.Weapon() != nil && p.Projectile.Weapon().ShortName() == ejectName && p.Parent() != g.player.Unit {
				g.spawnEjectionPodSmokeEffects(p)
			}
		}
	}
	p.Update(g.player.CameraPosXY())
}

// queueDelayedProjectile queues a projectile on a timed delay (seconds) between shots
func (g *Game) queueDelayedProjectile(pIndex int, w model.Weapon, e model.Entity) {
	delay := float64(pIndex) * w.ProjectileDelay()
	spread := w.ProjectileSpread()

	playSFX := false
	switch w.Type() {
	case model.BALLISTIC:
		// most ballistics play the sound effect every shot
		switch w.Classification() {
		case model.BALLISTIC_MACHINEGUN:
			playSFX = false
		case model.BALLISTIC_LBX_AC:
			playSFX = false
		default:
			playSFX = true
		}
	case model.ENERGY:
		// energy weapons plays the sound effect every shot
		playSFX = true
	case model.MISSILE:
		// missile weapons play the sound effect every N shots based on weapon
		switch w.Classification() {
		case model.MISSILE_LRM:
			playSFX = pIndex%5 == 0
		case model.MISSILE_SRM:
			playSFX = pIndex%2 == 0
		default:
			panic(fmt.Sprintf("unhandled missile weapon classification (%v) for %s", w.Classification(), w.Name()))
		}
	default:
		panic(fmt.Sprintf("unhandled weapon type (%v) for %s", w.Type(), w.Name()))
	}

	p := NewDelayedProjectileSpawn(delay, spread, w, e, playSFX)
	g.delayedProjectiles[p] = struct{}{}
}

// updateDelayedProjectiles updates timers on delayed projectiles and spawns them as they finish counting down
func (g *Game) updateDelayedProjectiles() {
	for p := range g.delayedProjectiles {
		p.delay -= model.SECONDS_PER_TICK
		if p.delay <= 0 {
			delete(g.delayedProjectiles, p)
			g.spawnProjectile(p)
		}
	}
}

// spawnProjectile puts a projectile in play
func (g *Game) spawnProjectile(p *ProjectileSpawn) *model.Projectile {
	w, u := p.weapon, model.EntityUnit(p.parent)
	if u == nil {
		return nil
	}

	isPlayerProjectile := u == g.player.Unit
	var projectile *model.Projectile

	var spreadAngle, spreadPitch float64
	if p.spread > 0 {
		// randomly generate spread for this projectile
		spreadAngle = randFloat(-p.spread, p.spread)
		spreadPitch = randFloat(-p.spread, p.spread)
	}

	var convergencePoint *geom3d.Vector3
	if isPlayerProjectile {
		cSprite := g.spriteInCrosshairs()
		var cEntity model.Entity
		if cSprite != nil {
			cEntity = cSprite.Entity
		}
		convergencePoint = model.ConvergencePoint(u, cEntity)
	} else {
		convergencePoint = model.ConvergencePoint(u, u.Target())
	}

	pHeading, pPitch := u.Heading(), u.Pitch()
	if u.HasTurret() {
		pHeading = u.TurretAngle()
	}

	if convergencePoint == nil {
		projectile = w.SpawnProjectile(pHeading+spreadAngle, pPitch+spreadPitch, u)
	} else {
		projectile = w.SpawnProjectileToward(convergencePoint, u)
		if p.spread > 0 {
			projectile.SetHeading(projectile.Heading() + spreadAngle)
			projectile.SetPitch(projectile.Pitch() + spreadPitch)
		}
	}

	if projectile != nil {
		pTemplate := projectileSpriteForWeapon(w)
		pSprite := pTemplate.Clone()
		pSprite.Projectile = projectile
		pSprite.Entity = projectile
		g.sprites.AddProjectile(pSprite)

		if p.sfxEnabled {
			if u == g.player.Unit {
				g.audio.PlayLocalWeaponFireAudio(w)
			} else {
				g.audio.PlayExternalWeaponFireAudio(g, w, u)
			}
		}

		s := g.getSpriteFromEntity(p.parent)
		if s != nil {
			// illuminate source sprite unit firing the projectile
			s.SetIlluminationPeriod(5000, 0.35)
		}
	}
	return projectile
}
