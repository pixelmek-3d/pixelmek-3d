package game

import (
	"fmt"
	"sync"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"
	"github.com/harbdog/raycaster-go/geom3d"
)

type DelayedProjectileSpawn struct {
	delay  float64
	weapon model.Weapon
	parent model.Entity
}

func (g *Game) initCombatVariables() {
	g.delayedProjectiles = make(map[*DelayedProjectileSpawn]struct{}, 256)
}

func (g *Game) fireWeapon() {
	// weapons test from model
	armament := g.player.Armament()
	if len(armament) == 0 {
		return
	}

	// in case convergence point not set, use player heading and pitch
	pAngle, pPitch := g.player.Angle(), g.player.Pitch()
	convergencePoint := g.player.ConvergencePoint
	// convergenceDistance := g.player.ConvergenceDistance

	for _, weapon := range armament {
		if weapon.Cooldown() > 0 {
			continue
		}

		var projectile *model.Projectile
		if convergencePoint == nil {
			projectile = weapon.SpawnProjectile(pAngle, pPitch, g.player.Entity)
		} else {
			projectile = weapon.SpawnProjectileToward(convergencePoint, g.player.Entity)
		}

		if projectile != nil {
			weapon.TriggerCooldown()

			// TODO: make projectiles spawned by player use their head-on facing angle for the first several frames to avoid
			//       them using a facing that looks weird (like lasers are doing when fired from arm location)

			pTemplate := projectileSpriteForWeapon(weapon)
			pSprite := pTemplate.Clone()
			pSprite.Entity = projectile
			g.sprites.addProjectile(pSprite)

			// use go routine to handle creation of multiple projectiles after time delay
			if weapon.ProjectileCount() > 1 {
				for i := 1; i < weapon.ProjectileCount(); i++ {
					g.queueDelayedProjectile(float64(i)*weapon.ProjectileDelay(), weapon, g.player.Entity)
				}
			}
		}
	}
}

func (g *Game) fireTestWeaponAtPlayer() {
	// Just for testing! Firing test projectiles at player
	playerPosition := g.player.Pos()
	for spriteType := range g.sprites.sprites {
		g.sprites.sprites[spriteType].Range(func(k, _ interface{}) bool {
			var pX, pY, pZ float64
			var entity model.Entity

			switch spriteType {
			case MechSpriteType:
				s := k.(*render.MechSprite)
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()+0.4
				entity = s.Entity

			case VehicleSpriteType:
				s := k.(*render.VehicleSprite)
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()+0.2
				entity = s.Entity

			case VTOLSpriteType:
				s := k.(*render.VTOLSprite)
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()
				entity = s.Entity

			case InfantrySpriteType:
				s := k.(*render.InfantrySprite)
				sPosition := s.Pos()
				pX, pY, pZ = sPosition.X, sPosition.Y, s.PosZ()+0.1
				entity = s.Entity
			}

			if entity == nil {
				return true
			}

			pLine := geom3d.Line3d{X1: pX, Y1: pY, Z1: pZ, X2: playerPosition.X, Y2: playerPosition.Y, Z2: randFloat(0.1, 0.7)}
			pHeading, pPitch := pLine.Heading(), pLine.Pitch()

			// TESTING: needed until turret heading is separated from heading angle so projectiles come from correct postion
			entity.SetAngle(pHeading)
			entity.SetPitch(pPitch)

			for _, weapon := range entity.Armament() {
				if weapon.Cooldown() > 0 {
					continue
				}

				projectile := weapon.SpawnProjectile(pHeading, pPitch, entity)
				if projectile != nil {
					// TODO: add muzzle flash effect on being fired at
					weapon.TriggerCooldown()

					pTemplate := projectileSpriteForWeapon(weapon)
					pSprite := pTemplate.Clone()
					pSprite.Entity = projectile
					g.sprites.addProjectile(pSprite)

					// use go routine to handle creation of multiple projectiles after time delay
					if weapon.ProjectileCount() > 1 {
						for i := 1; i < weapon.ProjectileCount(); i++ {
							g.queueDelayedProjectile(float64(i)*weapon.ProjectileDelay(), weapon, entity)
						}
					}
				}
			}

			return true
		})
	}
}

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

func (g *Game) asyncProjectileUpdate(p *render.ProjectileSprite, wg *sync.WaitGroup) {
	defer wg.Done()

	if p.Velocity() != 0 {
		pPosition := p.Pos()
		trajectory := geom3d.Line3dFromAngle(pPosition.X, pPosition.Y, p.PosZ(), p.Angle(), p.Pitch(), p.Velocity())
		xCheck := trajectory.X2
		yCheck := trajectory.Y2
		zCheck := trajectory.Z2

		newPos, isCollision, collisions := g.getValidMove(p.Entity, xCheck, yCheck, zCheck, false)
		if isCollision || p.PosZ() <= 0 {
			var collisionEntity *EntityCollision
			if len(collisions) > 0 {
				// apply damage to the first sprite entity that was hit
				collisionEntity = collisions[0]
				entity := collisionEntity.entity

				if entity == g.player.Entity {
					// TODO: visual response to player being hit
					println("ouch!")
				} else {
					damage := p.Damage()
					entity.ApplyDamage(damage)

					// TODO: visual method for showing damage was done
					hp, maxHP := entity.ArmorPoints()+entity.StructurePoints(), entity.MaxArmorPoints()+entity.MaxStructurePoints()
					percentHP := 100 * (hp / maxHP)
					fmt.Printf("[%0.2f%s] hit for %0.1f (HP: %0.1f/%0.0f)\n", percentHP, "%", damage, hp, maxHP)
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
				effect := p.SpawnEffect(newPos.X, newPos.Y, p.PosZ(), p.Angle(), p.Pitch())

				g.sprites.addEffect(effect)
			}

		} else {
			p.SetPos(newPos)
			p.SetPosZ(zCheck)
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

func (g *Game) spawnDelayedProjectile(p *DelayedProjectileSpawn) {
	delete(g.delayedProjectiles, p)

	w, e := p.weapon, p.parent
	var projectile *model.Projectile

	convergencePoint := g.player.ConvergencePoint
	if e != g.player.Entity || convergencePoint == nil {
		projectile = w.SpawnProjectile(e.Angle(), e.Pitch(), e)
	} else {
		projectile = w.SpawnProjectileToward(convergencePoint, e)
	}

	if projectile != nil {
		pTemplate := projectileSpriteForWeapon(w)
		pSprite := pTemplate.Clone()
		pSprite.Entity = projectile
		g.sprites.addProjectile(pSprite)
	}
}
