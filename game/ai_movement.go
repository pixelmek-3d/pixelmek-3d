package game

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"

	bt "github.com/joeycumines/go-behaviortree"
	log "github.com/sirupsen/logrus"
)

func (a *AIBehavior) TurnToTarget() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			if a.u.UnitType() == model.EmplacementUnitType {
				// emplacements only have turrets
				return bt.Success, nil
			}

			target := model.EntityUnit(a.u.Target())
			if target == nil {
				return bt.Failure, nil
			}

			// calculate heading from unit to target
			var targetHeading float64

			// TODO: pathfinding does not need to be recalculated every tick
			a.pathing = a.g.mission.Pathing.FindPath(a.u.Pos(), target.Pos())
			log.Debugf("[%s] found path (%v -> %v): %+v", a.u.ID(), a.u.Pos(), target.Pos(), a.pathing)
			if len(a.pathing) > 0 {
				pos := a.u.Pos()
				nextPos := a.pathing[0]
				moveLine := &geom.Line{X1: pos.X, Y1: pos.Y, X2: nextPos.X, Y2: nextPos.Y}
				targetHeading = moveLine.Angle()
			} else {
				targetLine := geom3d.Line3d{
					X1: a.u.Pos().X, Y1: a.u.Pos().Y, Z1: a.u.PosZ(),
					X2: target.Pos().X, Y2: target.Pos().Y, Z2: target.PosZ(),
				}
				targetHeading = targetLine.Heading()
			}

			//log.Debugf("[%s] %0.1f -> turnToTarget @ %s", a.u.ID(), geom.Degrees(a.u.Heading()), target.ID())
			a.u.SetTargetHeading(targetHeading)
			return bt.Success, nil
		},
	)
}

func (a *AIBehavior) TurretToTarget() bt.Node {
	// TODO: handle units without turrets
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			target := model.EntityUnit(a.u.Target())
			if target == nil {
				return bt.Failure, nil
			}

			var zTargetOffset float64
			switch target.Anchor() {
			case raycaster.AnchorBottom:
				zTargetOffset = randFloat(target.CollisionHeight()/10, 4*target.CollisionHeight()/5)
			case raycaster.AnchorTop:
				zTargetOffset = -randFloat(target.CollisionHeight()/10, 4*target.CollisionHeight()/5)
			case raycaster.AnchorCenter:
				zTargetOffset = randFloat(-target.CollisionHeight()/2, target.CollisionHeight()/2)
			}

			// calculate distance from unit to target
			tLine := geom3d.Line3d{
				X1: a.u.Pos().X, Y1: a.u.Pos().Y, Z1: a.u.PosZ() + a.u.CockpitOffset().Y,
				X2: target.Pos().X, Y2: target.Pos().Y, Z2: target.PosZ() + zTargetOffset,
			}
			tDist := tLine.Distance()

			// determine approximate lead time needed for weapon projectile
			tWeapon := a.idealWeaponForDistance(tDist)
			if tWeapon != nil {
				// approximate position of target based on its current heading and speed for projectile flight time
				tProjectile := tWeapon.Projectile()
				tDelta := tDist / tProjectile.MaxVelocity()
				tLine = geom3d.Line3dFromAngle(target.Pos().X, target.Pos().Y, target.PosZ()+zTargetOffset, target.Heading(), 0, tDelta*target.Velocity())
			}

			// calculate angle/pitch from unit to target
			pLine := geom3d.Line3d{
				X1: a.u.Pos().X, Y1: a.u.Pos().Y, Z1: a.u.PosZ() + a.u.CockpitOffset().Y,
				X2: tLine.X2, Y2: tLine.Y2, Z2: tLine.Z2,
			}
			pHeading, pPitch := pLine.Heading(), pLine.Pitch()
			if a.u.TurretAngle() == pHeading && a.u.Pitch() == pPitch {
				return bt.Success, nil
			}

			//log.Debugf("[%s] %0.1f|%0.1f turretToTarget @ %s", a.u.ID(), geom.Degrees(a.u.TurretAngle()), geom.Degrees(pPitch), target.ID())
			a.u.SetTargetTurretAngle(pHeading)
			a.u.SetTargetPitch(pPitch)
			// TODO: return failure if not even close to target angle
			return bt.Success, nil
		},
	)
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

func (a *AIBehavior) VelocityToMax() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			if a.u.Velocity() == a.u.MaxVelocity() {
				return bt.Success, nil
			}

			//log.Debugf("[%s] %0.1f -> velocityMax", a.u.ID(), a.u.Velocity()*model.VELOCITY_TO_KPH)
			a.u.SetTargetVelocity(a.u.MaxVelocity())
			return bt.Success, nil
		},
	)
}

func (a *AIBehavior) DetermineForcedWithdrawal() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			if a.u.UnitType() == model.EmplacementUnitType {
				// emplacements cannot withdraw
				return bt.Failure, nil
			}

			// TODO: check if no weapons usable
			if a.u.StructurePoints() > 0.2*a.u.MaxStructurePoints() {
				return bt.Failure, nil
			}
			log.Debugf("[%s] -> determineForcedWithdrawal", a.u.ID())
			return bt.Success, nil
		},
	)
}

func (a *AIBehavior) TurnToWithdraw() bt.Node {
	var withdrawPosition = &geom.Vector2{X: 60, Y: 99}
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			// TODO: pathfinding does not need to be recalculated every tick
			path := a.g.mission.Pathing.FindPath(a.u.Pos(), withdrawPosition)
			if len(path) == 0 {
				return bt.Success, nil
			}

			pos := a.u.Pos()
			nextPos := path[0]
			moveLine := &geom.Line{X1: pos.X, Y1: pos.Y, X2: nextPos.X, Y2: nextPos.Y}

			log.Debugf("[%s] %0.1f -> turnToWithdraw", a.u.ID(), geom.Degrees(a.u.Heading()))
			a.u.SetTargetHeading(moveLine.Angle())
			return bt.Success, nil
		},
	)
}

func (a *AIBehavior) InWithdrawArea() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			posX, posY := a.u.Pos().X, a.u.Pos().Y
			delta := 1.5
			if posX > delta && posX < float64(a.g.mapWidth)-delta &&
				posY > delta && posY < float64(a.g.mapHeight)-delta {
				return bt.Failure, nil
			}

			//log.Debugf("[%s] %v -> inWithdrawArea", a.u.ID(), a.u.Pos())
			return bt.Success, nil
		},
	)
}

func (a *AIBehavior) Withdraw() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			//log.Debugf("[%s] -> withdraw", a.u.ID())

			// TODO: unit safely escapes
			a.u.SetStructurePoints(0)
			return bt.Success, nil
		},
	)
}
