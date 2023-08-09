package game

import (
	"sync"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	log "github.com/sirupsen/logrus"
)

type DelayedProjectileSpawn struct {
	delay  float64
	weapon model.Weapon
	parent model.Entity
}

func (g *Game) initCombatVariables() {
	g.delayedProjectiles = make(map[*DelayedProjectileSpawn]struct{}, 256)
}

// fireWeapon fires the currently selected player weapon/weapon group
func (g *Game) fireWeapon() {
	// weapons test from model
	armament := g.player.Armament()
	if len(armament) == 0 {
		return
	}

	// in case convergence point not set, use player heading and pitch
	pAngle, pPitch := g.player.Heading()+g.player.TurretAngle(), g.player.Pitch()
	convergencePoint := g.player.convergencePoint
	// convergenceDistance := g.player.ConvergenceDistance

	for i, weapon := range armament {
		if weapon.Cooldown() > 0 {
			continue
		}

		isWeaponSelected := (g.player.fireMode == model.CHAIN_FIRE && i == int(g.player.selectedWeapon)) ||
			(g.player.fireMode == model.GROUP_FIRE && model.IsWeaponInGroup(weapon, g.player.selectedGroup, g.player.weaponGroups))
		if !isWeaponSelected {
			continue
		}

		if g.player.TriggerWeapon(weapon) {
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
						g.queueDelayedProjectile(float64(i)*weapon.ProjectileDelay(), weapon, g.player.Unit)
					}
				}
			}

			g.audio.PlayLocalWeaponFireAudio(weapon)
		}
	}
}

func (g *Game) fireTestWeaponAtPlayer() {
	// Just for testing! Firing test projectiles at player
	playerPosition := g.player.Pos()
	playerPositionZ := g.player.PosZ()
	for spriteType := range g.sprites.sprites {
		g.sprites.sprites[spriteType].Range(func(k, _ interface{}) bool {
			var pX, pY, pZ float64
			var unit model.Unit

			switch spriteType {
			case MechSpriteType:
				s := k.(*render.MechSprite)
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()+0.4
				unit = model.EntityUnit(s.Entity)

			case VehicleSpriteType:
				s := k.(*render.VehicleSprite)
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()+0.2
				unit = model.EntityUnit(s.Entity)

			case VTOLSpriteType:
				s := k.(*render.VTOLSprite)
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()
				unit = model.EntityUnit(s.Entity)

			case InfantrySpriteType:
				s := k.(*render.InfantrySprite)
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()+0.1
				unit = model.EntityUnit(s.Entity)

			case EmplacementSpriteType:
				s := k.(*render.EmplacementSprite)
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()+0.1
				unit = model.EntityUnit(s.Entity)
			}

			if unit == nil {
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

			for _, weapon := range unit.Armament() {
				if weapon.Cooldown() > 0 {
					continue
				}

				if unit.TriggerWeapon(weapon) {
					projectile := weapon.SpawnProjectile(pHeading, pPitch, unit)
					if projectile != nil {
						// TODO: add muzzle flash effect on being fired at

						pTemplate := projectileSpriteForWeapon(weapon)
						pSprite := pTemplate.Clone()
						pSprite.Projectile = projectile
						pSprite.Entity = projectile
						g.sprites.addProjectile(pSprite)

						// queue creation of multiple projectiles after time delay
						if weapon.ProjectileCount() > 1 {
							for i := 1; i < weapon.ProjectileCount(); i++ {
								g.queueDelayedProjectile(float64(i)*weapon.ProjectileDelay(), weapon, unit)
							}
						}
					}
				}
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

		// adjust pitch and heading if is a locked missile projectile
		missileWeapon, isMissile := p.Projectile.Weapon().(*model.MissileWeapon)
		if isMissile && missileWeapon.IsLockOn() {
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

				hp, maxHP := entity.ArmorPoints()+entity.StructurePoints(), entity.MaxArmorPoints()+entity.MaxStructurePoints()
				percentHP := 100 * (hp / maxHP)

				if g.debug {
					if entity == g.player.Unit {
						// TODO: visual response to player being hit
						log.Debugf("[%0.2f%s] player hit for %0.1f (HP: %0.1f/%0.0f)", percentHP, "%", damage, hp, maxHP)
					} else {
						// TODO: visual method for showing damage was done
						log.Debugf("[%0.2f%s] unit hit for %0.1f (HP: %0.1f/%0.0f)", percentHP, "%", damage, hp, maxHP)
					}
				}
			}

			// destroy projectile after applying damage so it can calculate dropoff if needed
			p.Destroy()

			// make a sprite/wall getting hit by projectile cause some visual effect
			if p.ImpactEffect.Sprite != nil {
				if collisionEntity != nil {
					// use the first collision point to place effect at
					newPos = collisionEntity.collision
				}

				// TODO: give impact effect optional ability to have some velocity based on the projectile movement upon impact if it didn't hit a wall
				effect := p.SpawnEffect(newPos.X, newPos.Y, newPosZ, p.Heading(), p.Pitch())

				g.sprites.addEffect(effect)
			}

			if len(p.ImpactAudio) > 0 {
				// play impact effect audio
				// determine distance and player camera relative direction of impact for volume and panning
				playerPos := g.player.Pos()
				playerHeading := g.player.Heading() + g.player.TurretAngle()

				impactLine := geom3d.Line3d{
					X1: playerPos.X, Y1: playerPos.Y, Z1: g.player.cameraZ,
					X2: newPos.X, Y2: newPos.Y, Z2: newPosZ,
				}
				impactDist := impactLine.Distance()
				impactHeading := impactLine.Heading()

				relHeading := -model.AngleDistance(playerHeading, impactHeading)
				relPercent := 1 - (geom.HalfPi-relHeading)/geom.HalfPi

				impactVolume := (20 - impactDist) / 20
				if impactVolume > 0.05 {
					g.audio.PlaySFX(p.ImpactAudio, impactVolume, relPercent)
				}
			}

		} else {
			p.SetPos(newPos)
			p.SetPosZ(newPosZ)
		}
	}
	p.Update(g.player.Pos())
}

// queueDelayedProjectile queues a projectile on a timed delay (seconds)
func (g *Game) queueDelayedProjectile(delay float64, w model.Weapon, e model.Entity) {
	p := &DelayedProjectileSpawn{
		delay:  delay,
		weapon: w,
		parent: e,
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

	convergencePoint := g.player.convergencePoint
	if e != g.player.Unit || convergencePoint == nil {
		projectile = w.SpawnProjectile(e.Heading()+e.TurretAngle(), e.Pitch(), e)
	} else {
		projectile = w.SpawnProjectileToward(convergencePoint, e)
	}

	if projectile != nil {
		pTemplate := projectileSpriteForWeapon(w)
		pSprite := pTemplate.Clone()
		pSprite.Projectile = projectile
		pSprite.Entity = projectile
		g.sprites.addProjectile(pSprite)

		if e == g.player.Unit {
			g.audio.PlayLocalWeaponFireAudio(w)
		} // else {
		// TODO: PlayExternalWeaponFireAudio
		// }
	}
}
