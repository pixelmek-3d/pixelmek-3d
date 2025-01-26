package game

import (
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"

	log "github.com/sirupsen/logrus"
)

func (g *Game) updateUnitPosition(u model.Unit) {
	if u.Powered() != model.POWER_ON {
		u.SetVelocity(0)
		u.SetVelocityZ(0)
		u.Update()
		return
	}

	if u.Update() {
		position, posZ := u.Pos(), u.PosZ()
		velocity, velocityZ := u.Velocity(), u.VelocityZ()

		moveHeading := u.Heading()
		if u.JumpJetsActive() || (posZ > 0 && u.JumpJets() > 0) {
			// while jumping, or still in air after jumping, continue from last jump jet active heading and velocity
			moveHeading = u.JumpJetHeading()
			velocity = u.JumpJetVelocity()
		}
		moveLine := geom.LineFromAngle(position.X, position.Y, moveHeading, velocity)
		moveX, moveY, moveZ := moveLine.X2, moveLine.Y2, posZ+velocityZ

		newPos, newPosZ, isCollision, collisions := g.getValidMove(u, moveX, moveY, moveZ, true)
		if !(newPos.Equals(position) && newPosZ == posZ) {
			u.SetPos(newPos)
			u.SetPosZ(newPosZ)
			//log.Debugf("[%s] unit moved (%v -> %v) heading @ %0.3f", u.ID(), position, newPos, moveHeading)
		}

		if isCollision && len(collisions) > 0 {
			// apply damage to the first sprite entity that was hit
			collisionEntity := collisions[0]

			// TODO: collision damage based on unit type, size, speed, and collision entity type
			collisionDamage := 0.01

			// apply more damage if it is a tree or foliage (MapSprite)
			mapSprite := g.getMapSpriteFromEntity(collisionEntity.entity)
			if mapSprite != nil {
				collisionDamage = 0.1
			}

			collisionEntity.entity.ApplyDamage(collisionDamage)
			if g.debug {
				hp, maxHP := collisionEntity.entity.ArmorPoints()+collisionEntity.entity.StructurePoints(), collisionEntity.entity.MaxArmorPoints()+collisionEntity.entity.MaxStructurePoints()
				log.Debugf("collided for %0.1f (HP: %0.1f/%0.1f)", collisionDamage, hp, maxHP)
			}
		}
	}
}
