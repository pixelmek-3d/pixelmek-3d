package game

import (
	"fmt"
	"sync"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"
	log "github.com/sirupsen/logrus"
)

type DelayedProjectileSpawn struct {
	delay      float64
	spread     float64
	weapon     model.Weapon
	parent     model.Entity
	sfxEnabled bool
}

func (g *Game) initCombatVariables() {
	g.delayedProjectiles = make(map[*DelayedProjectileSpawn]struct{}, 256)
}

// firePlayerWeapon fires currently selected player weapon/weapon group or input weapon group
func (g *Game) firePlayerWeapon(weaponGroupFire int) bool {
	// weapons test from model
	weaponsFired := false
	if g.player.Powered() != model.POWER_ON {
		// TODO: when shutdown, show weapons as disabled and disallow cycling weapons
		return weaponsFired
	}
	armament := g.player.Armament()
	if len(armament) == 0 {
		return weaponsFired
	}

	if weaponGroupFire < 0 && g.player.fireMode == model.GROUP_FIRE {
		// indicate firing current selected weapon group
		weaponGroupFire = int(g.player.selectedGroup)
	}

	// in case convergence point not set, use player heading and pitch
	pAngle, pPitch := g.player.Heading()+g.player.TurretAngle(), g.player.Pitch()
	convergencePoint := g.player.convergencePoint
	// convergenceDistance := g.player.ConvergenceDistance

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
			// perform ammo check
			ammoCount := ammoBin.AmmoCount()
			if ammoCount == 0 {
				isWeaponWithNoAmmoNotFired = true
				continue
			}
		}

		if g.player.TriggerWeapon(weapon) {
			weaponsFired = true

			var projectile *model.Projectile
			if convergencePoint == nil {
				projectile = weapon.SpawnProjectile(pAngle, pPitch, g.player.Unit)
			} else {
				projectile = weapon.SpawnProjectileToward(convergencePoint, g.player.Unit)
			}
			if projectile != nil {
				pTemplate := projectileSpriteForWeapon(weapon)
				pSprite := pTemplate.Clone()
				pSprite.Projectile = projectile
				pSprite.Entity = projectile
				g.sprites.addProjectile(pSprite)

				// queue creation of multiple projectiles after time delay
				if weapon.ProjectileCount() > 1 {
					for i := 1; i < weapon.ProjectileCount(); i++ {
						g.queueDelayedProjectile(i, weapon, g.player.Unit)
					}
				}
			}

			// consume ammo
			if ammoBin != nil {
				ammoBin.ConsumeAmmo(weapon, 1)
				log.Debugf("[player] %s: %d", weapon.ShortName(), ammoBin.AmmoCount())
			}

			// play sound effect
			g.audio.PlayLocalWeaponFireAudio(weapon)
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

func (g *Game) fireTestWeaponAtPlayer() {
	// Just for testing! Firing test projectiles at player
	playerPosition := g.player.Pos()
	playerPositionZ := g.player.PosZ()
	var playerTarget model.Unit
	if g.player.Target() != nil {
		// if player has a target, only it shoots at the player
		playerTarget = model.EntityUnit(g.player.Target())
	}
	for spriteType := range g.sprites.sprites {
		g.sprites.sprites[spriteType].Range(func(k, _ interface{}) bool {
			var pX, pY, pZ float64
			var unit model.Unit
			var sprite *render.Sprite

			switch spriteType {
			case MechSpriteType:
				s := k.(*render.MechSprite)
				sprite = s.Sprite
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()+0.4
				unit = model.EntityUnit(s.Entity)

			case VehicleSpriteType:
				s := k.(*render.VehicleSprite)
				sprite = s.Sprite
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()+0.2
				unit = model.EntityUnit(s.Entity)

			case VTOLSpriteType:
				s := k.(*render.VTOLSprite)
				sprite = s.Sprite
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()
				unit = model.EntityUnit(s.Entity)

			case InfantrySpriteType:
				s := k.(*render.InfantrySprite)
				sprite = s.Sprite
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()+0.1
				unit = model.EntityUnit(s.Entity)

			case EmplacementSpriteType:
				s := k.(*render.EmplacementSprite)
				sprite = s.Sprite
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()+0.1
				unit = model.EntityUnit(s.Entity)
			}

			if unit == nil || (playerTarget != nil && playerTarget != unit) {
				return true
			}

			pLine := geom3d.Line3d{X1: pX, Y1: pY, Z1: pZ, X2: playerPosition.X, Y2: playerPosition.Y, Z2: playerPositionZ + randFloat(0.1, 0.7)}
			pHeading, pPitch := pLine.Heading(), pLine.Pitch()

			if unit.HasTurret() {
				unit.SetTurretAngle(pHeading)
			} else {
				unit.SetHeading(pHeading)
			}
			unit.SetPitch(pPitch)

			weaponFired := false
			for _, weapon := range unit.Armament() {
				if weapon.Cooldown() > 0 {
					continue
				}

				ammoBin := weapon.AmmoBin()
				if ammoBin != nil {
					// perform ammo check
					ammoCount := ammoBin.AmmoCount()
					if ammoCount == 0 {
						continue
					}
				}

				if unit.TriggerWeapon(weapon) {
					weaponFired = true
					projectile := weapon.SpawnProjectile(pHeading, pPitch, unit)
					if projectile != nil {
						pTemplate := projectileSpriteForWeapon(weapon)
						pSprite := pTemplate.Clone()
						pSprite.Projectile = projectile
						pSprite.Entity = projectile
						g.sprites.addProjectile(pSprite)

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
						log.Debugf("[%s %s] %s: %d", unit.Name(), unit.Variant(), weapon.ShortName(), ammoBin.AmmoCount())
					}
				}
			}

			if weaponFired {
				// illuminate source sprite unit firing the weapon
				sprite.SetIlluminationPeriod(5000, 0.35)
			}

			return true
		})
	}
}

// updateProjectiles updates the state of all projectiles in play
func (g *Game) updateProjectiles() {
	// update and spawn projectiles on delay timers
	g.updateDelayedProjectiles()

	// perform concurrent projectile updates
	var wg sync.WaitGroup

	g.sprites.sprites[ProjectileSpriteType].Range(func(k, _ interface{}) bool {
		p := k.(*render.ProjectileSprite)
		p.DecreaseLifespan(1)
		if p.Lifespan() <= 0 {
			g.sprites.deleteProjectile(p)
			return true
		}

		wg.Add(1)
		go g.asyncProjectileUpdate(p, &wg)

		return true
	})

	// Update animated effects
	g.sprites.sprites[EffectSpriteType].Range(func(k, _ interface{}) bool {
		e := k.(*render.EffectSprite)
		e.Update(g.player.Pos())
		if e.LoopCounter() >= e.LoopCount {
			g.sprites.deleteEffect(e)
		}

		return true
	})

	wg.Wait()
}

// asyncProjectileUpdate updates the positions of a projectile in a parallel fashion
func (g *Game) asyncProjectileUpdate(p *render.ProjectileSprite, wg *sync.WaitGroup) {
	defer wg.Done()

	if p.Velocity() != 0 {
		pPos := p.Pos()

		_, isBallistic := p.Projectile.Weapon().(*model.BallisticWeapon)
		_, isEnergy := p.Projectile.Weapon().(*model.EnergyWeapon)
		missileWeapon, isMissile := p.Projectile.Weapon().(*model.MissileWeapon)

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

		if isBallistic || (p.Projectile.InExtremeRange() && !isEnergy) {
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

				if g.debug {
					hp, maxHP := entity.ArmorPoints()+entity.StructurePoints(), entity.MaxArmorPoints()+entity.MaxStructurePoints()
					percentHP := 100 * (hp / maxHP)

					if entity == g.player.Unit {
						// TODO: visual response to player being hit
						log.Debugf("[%0.2f%s] player hit for %0.1f (HP: %0.1f/%0.0f)", percentHP, "%", damage, hp, maxHP)
					} else {
						// TODO: ui indicator for showing damage was done
						log.Debugf("[%0.2f%s] unit hit for %0.1f (HP: %0.1f/%0.0f)", percentHP, "%", damage, hp, maxHP)
					}
				}
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
				g.sprites.addEffect(effect)
			}

			// play impact effect audio
			g.audio.PlayProjectileImpactAudio(g, p)

		} else {
			p.SetPos(newPos)
			p.SetPosZ(newPosZ)

			// smoke trail for ejection pod
			if p.Projectile.Weapon() != nil && p.Projectile.Weapon().ShortName() == ejectName {
				g.spawnEjectionPodSmokeEffects(p)
			}
		}
	}
	p.Update(g.player.Pos())
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

	p := &DelayedProjectileSpawn{
		delay:      delay,
		spread:     spread,
		weapon:     w,
		parent:     e,
		sfxEnabled: playSFX,
	}
	g.delayedProjectiles[p] = struct{}{}
}

// updateDelayedProjectiles updates timers on delayed projectiles and spawns them as they finish counting down
func (g *Game) updateDelayedProjectiles() {
	for p := range g.delayedProjectiles {
		p.delay -= model.SECONDS_PER_TICK
		if p.delay <= 0 {
			g.spawnDelayedProjectile(p)
		}
	}
}

// spawnDelayedProjectile puts a projectile in play that was delayed
func (g *Game) spawnDelayedProjectile(p *DelayedProjectileSpawn) {
	delete(g.delayedProjectiles, p)

	w, e := p.weapon, model.EntityUnit(p.parent)
	if e == nil {
		return
	}

	var projectile *model.Projectile

	var spreadAngle, spreadPitch float64
	if p.spread > 0 {
		// randomly generate spread for this projectile
		spreadAngle = randFloat(-p.spread, p.spread)
		spreadPitch = randFloat(-p.spread, p.spread)
	}

	convergencePoint := g.player.convergencePoint
	if e != g.player.Unit || convergencePoint == nil {
		projectile = w.SpawnProjectile(e.Heading()+e.TurretAngle()+spreadAngle, e.Pitch()+spreadPitch, e)
	} else {
		projectile = w.SpawnProjectileToward(convergencePoint, e)
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
		g.sprites.addProjectile(pSprite)

		if p.sfxEnabled {
			if e == g.player.Unit {
				g.audio.PlayLocalWeaponFireAudio(w)
			} else {
				g.audio.PlayExternalWeaponFireAudio(g, w, e)
			}
		}

		s := g.getSpriteFromEntity(p.parent)
		if s != nil {
			// illuminate source sprite unit firing the projectile
			s.SetIlluminationPeriod(5000, 0.35)
		}
	}
}
