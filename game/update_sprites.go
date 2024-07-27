package game

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"
)

func (g *Game) UpdateSprites() {
	// Update for animated sprite movement
	for spriteType := range g.sprites.sprites {
		g.sprites.sprites[spriteType].Range(func(k, _ interface{}) bool {
			g.updateSprite(spriteType, k.(raycaster.Sprite))
			return true
		})
	}
}

func (g *Game) updateSprite(spriteType SpriteType, sInterface raycaster.Sprite) {
	switch spriteType {
	case MapSpriteType:
		s := sInterface.(*render.Sprite)
		if s.IsDestroyed() {
			destroyCounter := s.DestroyCounter()
			if destroyCounter == 0 {
				// start the destruction process but do not remove yet
				// TODO: when tree is destroyed by projectile, add fire effect (energy and missile only)
				fxDuration := g.spawnGenericDestroyEffects(s, false)
				s.SetDestroyCounter(geom.ClampInt(fxDuration, 1, fxDuration))
			} else if destroyCounter == 1 {
				// delete when the counter is basically done (to differentiate with default int value 0)
				g.sprites.deleteMapSprite(s)
			} else {
				s.Update(g.player.Pos())
				s.SetDestroyCounter(destroyCounter - 1)
			}
			break
		}

		g.updateSpritePosition(s)
		s.Update(g.player.Pos())

	case MechSpriteType:
		s := sInterface.(*render.MechSprite)
		sUnit := model.EntityUnit(s.Entity)
		if s.IsDestroyed() {
			if s.MechAnimation() != render.MECH_ANIMATE_DESTRUCT {
				// play unit destruction animation
				s.SetMechAnimation(render.MECH_ANIMATE_DESTRUCT, false)

				// spawn ejection pod
				g.spawnEjectionPod(s.Sprite)

			} else if s.LoopCounter() >= 1 {
				// delete when animation is over
				g.sprites.deleteMechSprite(s)
			} else {
				s.Update(g.player.Pos())
			}

			if sUnit.JumpJets() > 0 {
				g.removeJumpJetEffect(s.Sprite)
			}

			g.spawnMechDestroyEffects(s)
			break
		}

		mech := s.Mech()
		g.updateMechPosition(s)
		s.Update(g.player.Pos())
		g.updateWeaponCooldowns(sUnit)

		if sUnit.Powered() != model.POWER_ON {
			poweringOn := s.AnimationReversed()
			if mech.PowerOffTimer > 0 &&
				(s.MechAnimation() != render.MECH_ANIMATE_SHUTDOWN || poweringOn) {

				// start shutdown animation since unit is powering off
				s.SetMechAnimation(render.MECH_ANIMATE_SHUTDOWN, false)
			}
			if mech.PowerOffTimer <= 0 && mech.PowerOnTimer > 0 &&
				(s.MechAnimation() != render.MECH_ANIMATE_SHUTDOWN || !poweringOn) {

				// reverse shutdown animation since unit is powering on
				s.SetMechAnimation(render.MECH_ANIMATE_SHUTDOWN, true)

			}
			if s.MechAnimation() != render.MECH_ANIMATE_SHUTDOWN {
				s.SetMechAnimation(render.MECH_ANIMATE_SHUTDOWN, true)
			}
		} else {
			if mech.JumpJetsActive() {
				falling := s.AnimationReversed()
				if s.MechAnimation() != render.MECH_ANIMATE_JUMP_JET || falling {
					s.SetMechAnimation(render.MECH_ANIMATE_JUMP_JET, false)

					// spawn jump jet effect when first starting jump jet
					g.spawnJumpJetEffect(s.Sprite)
				}
			} else if s.VelocityZ() < 0 {
				falling := s.AnimationReversed()
				if s.MechAnimation() != render.MECH_ANIMATE_JUMP_JET || !falling {
					// reverse jump jet animation for falling
					s.SetMechAnimation(render.MECH_ANIMATE_JUMP_JET, true)

					// remove jump jet effect since jump jet no longer active
					g.removeJumpJetEffect(s.Sprite)
				}
			} else if s.Velocity() == 0 && s.VelocityZ() == 0 {
				if s.MechAnimation() != render.MECH_ANIMATE_IDLE {
					s.SetMechAnimation(render.MECH_ANIMATE_IDLE, false)
				}
			} else {
				if s.MechAnimation() != render.MECH_ANIMATE_STRUT {
					s.SetMechAnimation(render.MECH_ANIMATE_STRUT, false)
				}
			}
		}

		if s.StrideStomp() {
			s.ResetStrideStomp()
			pos, posZ := s.Pos(), s.PosZ()
			mechStompFile, err := StompSFXForMech(mech)
			if err == nil {
				g.audio.PlayExternalAudio(g, mechStompFile, pos.X, pos.Y, posZ, 2.5, 0.35)
			}
		}

		if mech.JumpJets() > 0 {
			mechJumpFile, err := JumpJetSFXForMech(mech)
			if err == nil {
				switch {
				case mech.JumpJetsActive() && !s.JetsPlaying:
					s.JetsPlaying = true
					g.audio.PlayEntityAudioLoop(g, mechJumpFile, mech, 5.0, 0.35)
				case !mech.JumpJetsActive() && s.JetsPlaying:
					g.audio.StopEntityAudioLoop(g, mechJumpFile, mech)
					s.JetsPlaying = false
				}
			}
		}

	case VehicleSpriteType:
		s := sInterface.(*render.VehicleSprite)
		if s.IsDestroyed() {
			destroyCounter := s.DestroyCounter()
			if destroyCounter == 0 {
				// start the destruction process but do not remove yet
				fxDuration := g.spawnVehicleDestroyEffects(s)
				s.SetDestroyCounter(fxDuration)
			} else if destroyCounter == 1 {
				// delete when the counter is basically done (to differentiate with default int value 0)
				g.sprites.deleteVehicleSprite(s)
			} else {
				s.Update(g.player.Pos())
				s.SetDestroyCounter(destroyCounter - 1)
			}
			break
		}

		g.updateVehiclePosition(s)
		s.Update(g.player.Pos())
		g.updateWeaponCooldowns(model.EntityUnit(s.Entity))

	case VTOLSpriteType:
		s := sInterface.(*render.VTOLSprite)
		if s.IsDestroyed() {
			// unique VTOL destroy effect where it crashes towards the ground spinning
			destroyCounter := s.DestroyCounter()
			if destroyCounter == 0 {
				// start the destruction process but do not remove yet
				g.spawnVTOLDestroyEffects(s, true)
				s.SetVelocity(0)
				s.SetVelocityZ(0)

				// use the destroy counter to determine which effects to spawn
				s.SetDestroyCounter(1)
			} else if s.PosZ() <= 0 {
				// instantly delete if it gets below the ground
				g.sprites.deleteVTOLSprite(s)
				break
			} else {
				// spawn only smoke effects
				g.spawnVTOLDestroyEffects(s, false)
			}

			// fall towards the ground
			velocityZ := s.VelocityZ()
			s.SetVelocityZ(velocityZ - model.GRAVITY_UNITS_PTT)

			// put in a tailspin
			heading := s.Heading()
			s.SetHeading(model.ClampAngle2Pi(heading + (geom.Pi2 / model.TICKS_PER_SECOND)))

			hasCollision := g.updateSpritePosition(s.Sprite)
			if hasCollision {
				// instantly remove on collision with some more explosions
				g.spawnVTOLDestroyEffects(s, true)
				g.sprites.deleteVTOLSprite(s)
				break
			}

			s.Update(g.player.Pos())
			break
		}

		g.updateVTOLPosition(s)
		s.Update(g.player.Pos())
		g.updateWeaponCooldowns(model.EntityUnit(s.Entity))

	case InfantrySpriteType:
		s := sInterface.(*render.InfantrySprite)
		if s.IsDestroyed() {
			// infantry are destroyed immediately
			// TODO: if an infantry unit has death animation prior to deletion
			g.spawnInfantryDestroyEffects(s)
			g.sprites.deleteInfantrySprite(s)
			break
		}

		g.updateInfantryPosition(s)
		s.Update(g.player.Pos())
		g.updateWeaponCooldowns(model.EntityUnit(s.Entity))

	case EmplacementSpriteType:
		s := sInterface.(*render.EmplacementSprite)
		if s.IsDestroyed() {
			destroyCounter := s.DestroyCounter()
			if destroyCounter == 0 {
				// start the destruction process but do not remove yet
				fxDuration := g.spawnEmplacementDestroyEffects(s)
				s.SetDestroyCounter(fxDuration)
			} else if destroyCounter == 1 {
				// delete when the counter is basically done (to differentiate with default int value 0)
				g.sprites.deleteEmplacementSprite(s)
			} else {
				s.Update(g.player.Pos())
				s.SetDestroyCounter(destroyCounter - 1)
			}
			break
		}

		g.updateEmplacementPosition(s)
		s.Update(g.player.Pos())
		g.updateWeaponCooldowns(model.EntityUnit(s.Entity))
	}
}

func (g *Game) updateMechPosition(s *render.MechSprite) {
	if s.Mech().Powered() != model.POWER_ON {
		// TODO: refactor to use same update logic from player shutdown
		s.SetVelocity(0)
		s.SetVelocityZ(0)

		if s.Mech().Heat() < 0.7*s.Mech().MaxHeat() {
			s.Mech().SetPowered(model.POWER_ON)
		}
		s.Mech().Update()
		return
	}

	if s.Mech().Update() {
		// TODO: refactor to use same update function as g.updatePlayer()
		sPosition := s.Pos()
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		posZ, velocityZ := s.PosZ(), s.VelocityZ()
		if velocityZ != 0 {
			posZ += velocityZ
		}

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, newPosZ, isCollision, collisions := g.getValidMove(s.Entity, xCheck, yCheck, posZ, true)
		if !(newPos.Equals(s.Pos()) && newPosZ == s.PosZ()) {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}

		if isCollision && len(collisions) > 0 {
			// TODO: apply damage to the first sprite entity that was hit
		}
	}
}

func (g *Game) updateVehiclePosition(s *render.VehicleSprite) {
	if s.Vehicle().Powered() != model.POWER_ON {
		// TODO: refactor to use same update logic from player shutdown
		s.SetVelocity(0)
		s.SetVelocityZ(0)

		if s.Vehicle().Heat() < 0.7*s.Vehicle().MaxHeat() {
			s.Vehicle().SetPowered(model.POWER_ON)
		}
		s.Vehicle().Update()
		return
	}

	if s.Vehicle().Update() {
		// TODO: refactor to use same update function as g.updatePlayer()
		sPosition := s.Pos()
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		posZ, velocityZ := s.PosZ(), s.VelocityZ()
		if velocityZ != 0 {
			posZ += velocityZ
		}

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, newPosZ, isCollision, collisions := g.getValidMove(s.Entity, xCheck, yCheck, posZ, true)
		if !(newPos.Equals(s.Pos()) && newPosZ == s.PosZ()) {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}

		if isCollision && len(collisions) > 0 {
			// TODO: apply damage to the first sprite entity that was hit
		}
	}
}

func (g *Game) updateVTOLPosition(s *render.VTOLSprite) {
	if s.VTOL().Powered() != model.POWER_ON {
		// TODO: refactor to use same update logic from player shutdown
		s.SetVelocity(0)
		s.SetVelocityZ(0)

		if s.VTOL().Heat() < 0.7*s.VTOL().MaxHeat() {
			s.VTOL().SetPowered(model.POWER_ON)
		}
		s.VTOL().Update()
		return
	}

	if s.VTOL().Update() {
		// TODO: refactor to use same update function as g.updatePlayer()
		sPosition := s.Pos()
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		posZ, velocityZ := s.PosZ(), s.VelocityZ()
		if velocityZ != 0 {
			posZ += velocityZ
		}

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, newPosZ, isCollision, collisions := g.getValidMove(s.Entity, xCheck, yCheck, posZ, true)
		if !(newPos.Equals(s.Pos()) && newPosZ == s.PosZ()) {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}

		if isCollision && len(collisions) > 0 {
			// TODO: apply damage to the first sprite entity that was hit
		}
	}
}

func (g *Game) updateInfantryPosition(s *render.InfantrySprite) {
	if s.Infantry().Powered() != model.POWER_ON {
		// TODO: refactor to use same update logic from player shutdown
		s.SetVelocity(0)
		s.SetVelocityZ(0)

		if s.Infantry().Heat() < 0.7*s.Infantry().MaxHeat() {
			s.Infantry().SetPowered(model.POWER_ON)
		}
		s.Infantry().Update()
		return
	}

	if s.Infantry().Update() {
		// TODO: refactor to use same update function as g.updatePlayer()
		sPosition := s.Pos()
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		posZ, velocityZ := s.PosZ(), s.VelocityZ()
		if velocityZ != 0 {
			posZ += velocityZ
		}

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, newPosZ, isCollision, collisions := g.getValidMove(s.Entity, xCheck, yCheck, posZ, true)
		if !(newPos.Equals(s.Pos()) && newPosZ == s.PosZ()) {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}

		if isCollision && len(collisions) > 0 {
			// TODO: apply damage to the first sprite entity that was hit
		}
	}
}

func (g *Game) updateEmplacementPosition(s *render.EmplacementSprite) {
	if s.Emplacement().Powered() != model.POWER_ON {
		// TODO: refactor to use same update logic from player shutdown
		if s.Emplacement().Heat() < 0.7*s.Emplacement().MaxHeat() {
			s.Emplacement().SetPowered(model.POWER_ON)
		}
		return
	}
	s.Emplacement().Update()
}

func (g *Game) updateSpritePosition(s *render.Sprite) bool {
	if s.Velocity() != 0 || s.VelocityZ() != 0 {
		sPosition := s.Pos()
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2
		zCheck := s.PosZ() + s.VelocityZ()

		newPos, newPosZ, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, zCheck, false)
		if isCollision {
			return true
		} else {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}
	}
	return false
}
