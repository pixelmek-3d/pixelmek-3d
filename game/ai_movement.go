package game

import (
	"math"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"

	bt "github.com/joeycumines/go-behaviortree"
	log "github.com/sirupsen/logrus"
)

func (a *AIBehavior) updatePathingToPosition(toPos *geom.Vector2, recalcDistFactor float64) error {
	uPos := a.u.Pos()
	findNewPath := false
	pathing := a.piloting.pathing

	switch {
	case pathing.Len() == 0:
		findNewPath = true
	// TODO: case if stuck, findNewPath and use starting sequence size of 3?
	case int(toPos.X) != int(pathing.destPos.X) || int(toPos.Y) != int(pathing.destPos.Y):
		// if still some distance from target, do not recalc path to target until further
		toLine := geom.Line{
			X1: uPos.X, Y1: uPos.Y,
			X2: toPos.X, Y2: toPos.Y,
		}
		toDist := toLine.Distance()
		if toDist <= recalcDistFactor {
			findNewPath = true
		} else {
			deltaRecalcFactor := geom.Clamp(recalcDistFactor/2, 1, math.MaxFloat64)
			deltaX, deltaY := math.Abs(toPos.X-pathing.destPos.X), math.Abs(toPos.Y-pathing.destPos.Y)
			if deltaX > toDist/deltaRecalcFactor || deltaY > toDist/deltaRecalcFactor {
				findNewPath = true
			}
		}
	}

	if findNewPath {
		// find new path to reach target position
		toPath, err := a.g.mission.Pathing.FindPath(a.u.Pos(), toPos)
		if err != nil {
			log.Debug(err)
			return err
		} else {
			pathing.SetDestination(toPos, toPath)
			//log.Debugf("[%s] new pathing (%v -> %v): %+v", a.u.ID(), a.u.Pos(), pathing.destPos, pathing.destPath)
		}
	}

	if pathing.Len() > 0 {
		// determine if need to move to next position in path
		nextPos := pathing.Next()
		for geom.Distance2(uPos.X, uPos.Y, nextPos.X, nextPos.Y) < 1.0 {
			pathing.Pop()
			nextPos = pathing.Next()

			//log.Debugf("[%s] pathing pop (%v -> %v): %+v", a.u.ID(), a.u.Pos(), pathing.destPos, pathing.destPath)
			if nextPos == nil {
				break
			}
		}
	}
	return nil
}

func (a *AIBehavior) pathingHeading(toPos *geom.Vector2, toPosZ float64) float64 {
	// calculate heading from unit to position
	var toHeading float64 = a.u.Heading()
	if a.piloting.pathing.Len() > 0 {
		pos := a.u.Pos()
		nextPos := a.piloting.pathing.Next()
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

func (a *AIBehavior) pathingVelocity(targetHeading, targetVelocity float64) float64 {
	// return ideal target velocity to turn towards target heading
	headingDiff := model.AngleDistance(a.u.Heading(), targetHeading)
	if headingDiff > geom.Pi/4 {
		// reduce velocity more for sharper turns
		vTurnRatio := 0.25 + 0.75*(geom.Pi-math.Abs(headingDiff))/geom.Pi
		maxVelocity := geom.Clamp(a.u.MaxVelocity()*vTurnRatio, 0, targetVelocity)
		targetVelocity = geom.Clamp(targetVelocity, 0, maxVelocity)
	}
	return targetVelocity
}

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

		// log.Debugf("[%s] %0.1f -> turnToTarget @ %s", a.u.ID(), geom.Degrees(targetHeading), target.ID())
		a.u.SetTargetHeading(targetHeading)
		return bt.Success, nil
	}
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

		if a.u.HasTurret() {
			// log.Debugf("[%s] %0.1f|%0.1f turretToTarget @ %s", a.u.ID(), geom.Degrees(pHeading), geom.Degrees(pPitch), target.ID())
			a.u.SetTargetTurretAngle(pHeading)
		} else {
			// if weapon is not on cooldown and has LOS to target, override target heading towards target
			overrideHeading := iWeapon != nil && iWeapon.Cooldown() == 0 && a.g.lineOfSight(a.u, target)
			if overrideHeading {
				// log.Debugf("[%s] %0.1f|%0.1f (no turret)ToTarget @ %s", a.u.ID(), geom.Degrees(pHeading), geom.Degrees(pPitch), target.ID())
				a.u.SetTargetHeading(pHeading)
			}
		}

		a.u.SetTargetPitch(pPitch)
		return bt.Success, nil
	}
}

func (a *AIBehavior) VelocityToMax() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		targetVelocity := a.pathingVelocity(a.u.TargetHeading(), a.u.MaxVelocity())
		a.u.SetTargetVelocity(targetVelocity)
		// log.Debugf("[%s] %0.1f -> velocityMax", a.u.ID(), targetVelocity*model.VELOCITY_TO_KPH)
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

	var temporaryWithdrawArea = model.NewRect(0, 0, 2, 2)

	return func(_ []bt.Node) (bt.Status, error) {
		if a.u.WithdrawArea() == nil {
			// TODO: determine withdraw area based on starting position
			a.u.SetWithdrawArea(&temporaryWithdrawArea)
		}

		// TODO: randomly determine target withdraw position within area up front
		withdrawArea := a.u.WithdrawArea()
		withdrawPosition := &geom.Vector2{
			X: withdrawArea.X1 + withdrawArea.Dx()/2,
			Y: withdrawArea.Y1 + withdrawArea.Dy()/2,
		}

		a.updatePathingToPosition(withdrawPosition, 4)
		targetHeading := a.pathingHeading(withdrawPosition, a.u.PosZ())

		a.u.SetTargetHeading(targetHeading)

		// TODO: fighting withdraw if still has weapons?
		targetVelocity := a.pathingVelocity(targetHeading, a.u.MaxVelocity())
		a.u.SetTargetVelocity(targetVelocity)

		// log.Debugf("[%s] %0.1f -> turnToWithdraw", a.u.ID(), geom.Degrees(targetHeading))
		return bt.Success, nil
	}
}

func (a *AIBehavior) InWithdrawArea() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		if a.u.WithdrawArea() == nil {
			return bt.Failure, nil
		}

		withdrawArea := a.u.WithdrawArea()
		if withdrawArea.ContainsPoint(a.u.Pos().X, a.u.Pos().Y) {
			// log.Debugf("[%s] %v -> inWithdrawArea", a.u.ID(), a.u.Pos())
			return bt.Success, nil
		}

		return bt.Failure, nil
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

func (a *AIBehavior) GuardArea() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		guardArea := a.u.GuardArea()
		guardPath := a.u.PathStack()
		if guardArea == nil || guardPath == nil {
			return bt.Failure, nil
		}

		pos := a.u.Pos()

		// standing guard if guard area radius is 0 and close to guard area
		if guardArea.Radius == 0 && geom.Distance2(pos.X, pos.Y, guardArea.X, guardArea.Y) < 2 {
			guardPath.Reset()
			a.u.SetTargetVelocity(0)
			return bt.Success, nil
		}

		// use the path stack to keep track of next position to guard within the area
		nextPos := guardPath.Peek()
		if nextPos == nil {
			// select a random position within the guard area
			// TODO: verify random position isn't inside walls that would block reaching the position

			rng := model.NewRNG()
			rngAngle := model.RandFloat64In(0, 2*geom.Pi, rng)
			rngLine := geom.LineFromAngle(guardArea.X, guardArea.Y, rngAngle, guardArea.Radius)

			rngX := geom.Clamp(rngLine.X2, 0, float64(a.g.mapWidth))
			rngY := geom.Clamp(rngLine.Y2, 0, float64(a.g.mapHeight))
			nextPos = &geom.Vector2{X: rngX, Y: rngY}
			guardPath.Push(*nextPos)
		}

		if geom.Distance2(pos.X, pos.Y, nextPos.X, nextPos.Y) < 2 {
			guardPath.Pop()
		}

		a.updatePathingToPosition(nextPos, 4)
		targetHeading := a.pathingHeading(nextPos, a.u.PosZ())

		a.u.SetTargetHeading(targetHeading)
		a.u.SetTargetTurretAngle(targetHeading)

		// TODO: units on guard do not need to run, max velocity only if very far from guard area
		targetVelocity := a.pathingVelocity(targetHeading, a.u.MaxVelocity())
		a.u.SetTargetVelocity(targetVelocity)

		//log.Debugf("[%s] -> guard area -> nextPos: %v", a.u.ID(), nextPos)
		return bt.Success, nil
	}
}

func (a *AIBehavior) GuardUnit() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		guardPath := a.u.PathStack()
		if a.piloting.formation == nil || guardPath == nil {
			return bt.Failure, nil
		}

		leader := a.piloting.formation.leader
		if leader == nil || leader.IsDestroyed() || a.u == leader {
			return bt.Failure, nil
		}

		// TODO: determine guard position using some formation instead of some random position near leader

		a.updatePathingToPosition(leader.Pos(), 4)
		targetHeading := a.pathingHeading(leader.Pos(), leader.PosZ())

		a.u.SetTargetHeading(targetHeading)
		a.u.SetTargetTurretAngle(targetHeading)

		// TODO: match leader's velocity if close by
		targetVelocity := a.pathingVelocity(targetHeading, a.u.MaxVelocity())
		a.u.SetTargetVelocity(targetVelocity)

		//log.Debugf("[%s] -> guard unit [%s] -> heading: %0.2f", a.u.ID(), leader.ID(), geom.Degrees(targetHeading))
		return bt.Success, nil
	}
}

func (a *AIBehavior) PatrolPath() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		patrolPath := a.u.PathStack()
		if patrolPath == nil || patrolPath.Len() == 0 {
			return bt.Failure, nil
		}

		// determine if need to move to next position in path
		pos := a.u.Pos()
		nextPos := patrolPath.Peek()
		if geom.Distance2(pos.X, pos.Y, nextPos.X, nextPos.Y) < 2 {
			// unit is close enough, move to next path position for next cycle
			patrolPath.Push(*patrolPath.Pop())
			nextPos = patrolPath.Peek()
		}

		a.updatePathingToPosition(nextPos, 4)
		targetHeading := a.pathingHeading(nextPos, a.u.PosZ())

		a.u.SetTargetHeading(targetHeading)
		a.u.SetTargetTurretAngle(targetHeading)

		// TODO: units on patrol do not need to run, max velocity only if very far from next patrol position
		targetVelocity := a.pathingVelocity(targetHeading, a.u.MaxVelocity())
		a.u.SetTargetVelocity(targetVelocity)

		//log.Debugf("[%s] -> patrol path -> nextPos: %v", a.u.ID(), nextPos)
		return bt.Success, nil
	}
}

func (a *AIBehavior) HuntLastTargetArea() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		// TODO: hunt last target area
		//log.Debugf("[%s] -> hunt last target area", a.u.ID())
		return bt.Failure, nil
	}
}

func (a *AIBehavior) Wander() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		wanderPath := a.u.PathStack()
		if wanderPath == nil {
			return bt.Failure, nil
		}

		pos := a.u.Pos()

		// use the path stack to keep track of next position to guard within the area
		nextPos := wanderPath.Peek()
		if nextPos == nil {
			// select a random position within the map area
			// TODO: verify random position isn't inside walls that would block reaching the position

			rng := model.NewRNG()
			rngX := model.RandFloat64In(0, float64(a.g.mapWidth), rng)
			rngY := model.RandFloat64In(0, float64(a.g.mapHeight), rng)
			nextPos = &geom.Vector2{X: rngX, Y: rngY}
			wanderPath.Push(*nextPos)
		}

		if geom.Distance2(pos.X, pos.Y, nextPos.X, nextPos.Y) < 2 {
			// random position is too close, try another next cycle
			wanderPath.Pop()
			return bt.Success, nil
		}

		err := a.updatePathingToPosition(nextPos, 4)
		if err != nil {
			// error trying to reach random position, try another next cycle
			wanderPath.Pop()
			return bt.Failure, nil
		}
		targetHeading := a.pathingHeading(nextPos, a.u.PosZ())

		a.u.SetTargetHeading(targetHeading)
		a.u.SetTargetTurretAngle(targetHeading)

		// TODO: use cruise velocity for wandering
		a.u.SetTargetVelocity(a.u.MaxVelocity())

		//log.Debugf("[%s] -> wander path -> nextPos: %v", a.u.ID(), nextPos)
		return bt.Success, nil
	}
}
