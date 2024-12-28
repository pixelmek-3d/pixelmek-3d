package game

import (
	"math"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"

	bt "github.com/joeycumines/go-behaviortree"
	log "github.com/sirupsen/logrus"
)

func (a *AIBehavior) TurnToTarget() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		if a.u.UnitType() == model.EmplacementUnitType {
			// emplacements only have turrets
			return bt.Success, nil
		}

		target := model.EntityUnit(a.u.Target())
		if target == nil {
			return bt.Failure, nil
		}

		a.updatePathingToPosition(target.Pos(), 8)
		targetHeading := a.pathingHeading(target.Pos(), target.PosZ())

		// log.Debugf("[%s] %0.1f -> turnToTarget @ %s", a.u.ID(), geom.Degrees(a.u.Heading()), target.ID())
		a.u.SetTargetHeading(targetHeading)
		return bt.Success, nil
	}
}

func (a *AIBehavior) updatePathingToPosition(toPos *geom.Vector2, recalcDistFactor float64) {
	findNewPath := false
	switch {
	case a.piloting.Len() == 0:
		findNewPath = true
	case int(toPos.X) != int(a.piloting.destPos.X) || int(toPos.Y) != int(a.piloting.destPos.Y):
		// if still some distance from target, do not recalc path to target until further
		uPos := a.u.Pos()
		toLine := geom.Line{
			X1: uPos.X, Y1: uPos.Y,
			X2: toPos.X, Y2: toPos.Y,
		}
		toDist := toLine.Distance()
		if toDist <= recalcDistFactor {
			findNewPath = true
		} else {
			deltaRecalcFactor := geom.Clamp(recalcDistFactor/2, 1, math.MaxFloat64)
			deltaX, deltaY := math.Abs(toPos.X-a.piloting.destPos.X), math.Abs(toPos.Y-a.piloting.destPos.Y)
			if deltaX > toDist/deltaRecalcFactor || deltaY > toDist/deltaRecalcFactor {
				findNewPath = true
			}
		}
	}

	if findNewPath {
		// find new path to reach target position
		a.piloting = &AIPiloting{
			destPos:  toPos,
			destPath: a.g.mission.Pathing.FindPath(a.u.Pos(), toPos),
		}
		// log.Debugf("[%s] new path (%v -> %v): %+v", a.u.ID(), a.u.Pos(), a.pathing.pos, a.pathing.path)
	} else if a.piloting.Len() > 0 {
		// determine if need to move to next position in path
		pos := a.u.Pos()
		nextPos := a.piloting.Next()
		if geom.Distance2(pos.X, pos.Y, nextPos.X, nextPos.Y) < 1 {
			// unit is close to next path position
			a.piloting.Pop()
			// log.Debugf("[%s] path pop (%v -> %v): %+v", a.u.ID(), a.u.Pos(), a.pathing.pos, a.pathing.path)
		}
	}
}

func (a *AIBehavior) pathingHeading(toPos *geom.Vector2, toPosZ float64) float64 {
	// calculate heading from unit to position
	var toHeading float64 = a.u.Heading()
	if a.piloting.Len() > 0 {
		pos := a.u.Pos()
		nextPos := a.piloting.Next()
		toLine := geom.Line{X1: pos.X, Y1: pos.Y, X2: nextPos.X, Y2: nextPos.Y}
		toHeading = toLine.Angle()
	} else {
		toLine := geom3d.Line3d{
			X1: a.u.Pos().X, Y1: a.u.Pos().Y, Z1: a.u.PosZ(),
			X2: toPos.X, Y2: toPos.Y, Z2: toPosZ,
		}
		toHeading = toLine.Heading()
	}
	return toHeading
}

func (a *AIBehavior) TurretToTarget() func([]bt.Node) (bt.Status, error) {
	// TODO: handle units without turrets
	return func(_ []bt.Node) (bt.Status, error) {
		target := model.EntityUnit(a.u.Target())
		if target == nil {
			return bt.Failure, nil
		}

		// calculate distance from unit to target
		tLine := geom3d.Line3d{
			X1: a.u.Pos().X, Y1: a.u.Pos().Y, Z1: a.u.PosZ(),
			X2: target.Pos().X, Y2: target.Pos().Y, Z2: target.PosZ(),
		}
		tDist := tLine.Distance()

		// determine approximate lead distance needed for weapon projectile
		iWeapon := a.idealWeaponForDistance(tDist)
		iPos := model.TargetLeadPosition(a.u, target, iWeapon)

		// set intended target lead position for weapons fire decision
		a.gunnery.targetLeadPos = &geom.Vector2{X: iPos.X, Y: iPos.Y}

		// calculate angle/pitch from unit to target
		pLine := geom3d.Line3d{
			X1: a.u.Pos().X, Y1: a.u.Pos().Y, Z1: a.u.PosZ() + a.u.CockpitOffset().Y,
			X2: iPos.X, Y2: iPos.Y, Z2: iPos.Z,
		}
		pHeading, pPitch := pLine.Heading(), pLine.Pitch()
		currHeading, currPitch := a.u.TurretAngle(), a.u.Pitch()

		// TODO: if more distant, decrease angle/pitch check for target lock proximity
		acquireLock := model.AngleDistance(currHeading, pHeading) <= 0.5 && model.AngleDistance(currPitch, pPitch) <= 0.5

		// TODO: decrease lock percent delta if further from target
		lockDelta := 1.0 / model.TICKS_PER_SECOND
		if !acquireLock {
			lockDelta = -0.15 / model.TICKS_PER_SECOND
		}

		targetLock := a.u.TargetLock() + lockDelta
		if targetLock > 1.0 {
			targetLock = 1.0
		} else if targetLock < 0 {
			targetLock = 0
		}
		a.u.SetTargetLock(targetLock)

		// log.Debugf("[%s] %0.1f|%0.1f turretToTarget @ %s", a.u.ID(), geom.Degrees(a.u.TurretAngle()), geom.Degrees(pPitch), target.ID())
		a.u.SetTargetTurretAngle(pHeading)
		a.u.SetTargetPitch(pPitch)
		return bt.Success, nil
	}
}

func (a *AIBehavior) idealWeaponForDistance(dist float64) model.Weapon {
	realDist := dist * model.METERS_PER_UNIT

	var idealWeapon model.Weapon
	var idealDist float64
	for _, w := range a.u.Armament() {
		if model.WeaponAmmoCount(w) <= 0 {
			// only weapons with ammo remaining
			continue
		}
		weaponDist := w.Distance()
		if realDist > weaponDist {
			// only weapons within range
			continue
		}

		switch {
		case idealWeapon == nil:
			fallthrough
		case idealWeapon.Cooldown() > 0 && w.Cooldown() == 0:
			fallthrough
		case weaponDist < idealDist:
			idealWeapon = w
			idealDist = weaponDist
		}

	}
	return idealWeapon
}

func (a *AIBehavior) VelocityToMax() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		if a.u.Velocity() == a.u.MaxVelocity() {
			return bt.Success, nil
		}

		// log.Debugf("[%s] %0.1f -> velocityMax", a.u.ID(), a.u.Velocity()*model.VELOCITY_TO_KPH)
		a.u.SetTargetVelocity(a.u.MaxVelocity())
		return bt.Success, nil
	}
}

func (a *AIBehavior) DetermineForcedWithdrawal() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		if a.u.UnitType() == model.EmplacementUnitType {
			// emplacements cannot withdraw
			return bt.Failure, nil
		}

		// TODO: check if no weapons usable
		if a.u.StructurePoints() > 0.2*a.u.MaxStructurePoints() {
			return bt.Failure, nil
		}
		// log.Debugf("[%s] -> determineForcedWithdrawal", a.u.ID())
		return bt.Success, nil
	}
}

func (a *AIBehavior) TurnToWithdraw() func([]bt.Node) (bt.Status, error) {
	var withdrawPosition = &geom.Vector2{X: 60, Y: 99}
	return func(_ []bt.Node) (bt.Status, error) {
		// TODO: pathfinding does not need to be recalculated every tick
		// TODO: refactor to use similar method used for pathing TurnToTarget
		path := a.g.mission.Pathing.FindPath(a.u.Pos(), withdrawPosition)
		if len(path) == 0 {
			return bt.Success, nil
		}

		pos := a.u.Pos()
		nextPos := path[0]
		moveLine := &geom.Line{X1: pos.X, Y1: pos.Y, X2: nextPos.X, Y2: nextPos.Y}

		// log.Debugf("[%s] %0.1f -> turnToWithdraw", a.u.ID(), geom.Degrees(a.u.Heading()))
		a.u.SetTargetHeading(moveLine.Angle())
		return bt.Success, nil
	}
}

func (a *AIBehavior) InWithdrawArea() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		posX, posY := a.u.Pos().X, a.u.Pos().Y
		delta := 1.5
		if posX > delta && posX < float64(a.g.mapWidth)-delta &&
			posY > delta && posY < float64(a.g.mapHeight)-delta {
			return bt.Failure, nil
		}

		// log.Debugf("[%s] %v -> inWithdrawArea", a.u.ID(), a.u.Pos())
		return bt.Success, nil
	}
}

func (a *AIBehavior) Withdraw() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		log.Debugf("[%s] -> withdraw", a.u.ID())

		// TODO: unit safely escapes without exploding
		a.u.SetStructurePoints(0)
		return bt.Success, nil
	}
}

func (a *AIBehavior) GuardUnit() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		// TODO: guard unit
		//log.Debugf("[%s] -> guard unit", a.u.ID())
		return bt.Failure, nil
	}
}

func (a *AIBehavior) GuardArea() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		// TODO: guard area
		//log.Debugf("[%s] -> guard area", a.u.ID())
		return bt.Failure, nil
	}
}

func (a *AIBehavior) HuntLastTargetArea() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		// TODO: hunt last target area
		//log.Debugf("[%s] -> hunt last target area", a.u.ID())
		return bt.Failure, nil
	}
}

func (a *AIBehavior) PatrolPath() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		patrolPath := a.u.PatrolPath()
		if patrolPath != nil && patrolPath.Len() > 0 {
			// determine if need to move to next position in path
			pos := a.u.Pos()
			nextPos := patrolPath.Peek()
			if geom.Distance2(pos.X, pos.Y, nextPos.X, nextPos.Y) < 1 {
				// unit is close enough, move to next path position for next cycle
				patrolPath.Push(*patrolPath.Pop())
				nextPos = patrolPath.Peek()
			}

			a.updatePathingToPosition(nextPos, 4)
			targetHeading := a.pathingHeading(nextPos, a.u.PosZ())

			a.u.SetTargetHeading(targetHeading)
			a.u.SetTargetTurretAngle(targetHeading)

			a.u.SetTargetVelocity(a.u.MaxVelocity())

			//log.Debugf("[%s] -> patrol path -> nextPos: %v", a.u.ID(), nextPos)
			return bt.Success, nil
		}

		return bt.Failure, nil
	}
}

func (a *AIBehavior) Wander() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		// TODO: wander randomly
		//log.Debugf("[%s] -> wander randomly", a.u.ID())
		return bt.Failure, nil
	}
}
