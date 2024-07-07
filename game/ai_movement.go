package game

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"

	bt "github.com/joeycumines/go-behaviortree"
	log "github.com/sirupsen/logrus"
)

func (a *AIBehavior) turnToTarget() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			target := model.EntityUnit(a.u.Target())
			if target == nil {
				return bt.Failure, nil
			}

			// calculate heading from unit to target
			pLine := geom3d.Line3d{
				X1: a.u.Pos().X, Y1: a.u.Pos().Y, Z1: a.u.PosZ(),
				X2: target.Pos().X, Y2: target.Pos().Y, Z2: target.PosZ(),
			}
			pHeading := pLine.Heading()
			if a.u.Heading() == pHeading {
				return bt.Success, nil
			}

			log.Debugf("[%s] %0.1f -> turnToTarget @ %s", a.u.ID(), geom.Degrees(a.u.Heading()), target.ID())
			a.u.SetTargetHeading(pHeading)
			return bt.Success, nil
		},
	)
}

func (a *AIBehavior) turretToTarget() bt.Node {
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

			// calculate angle/pitch from unit to target
			pLine := geom3d.Line3d{
				X1: a.u.Pos().X, Y1: a.u.Pos().Y, Z1: a.u.PosZ() + a.u.CockpitOffset().Y,
				X2: target.Pos().X, Y2: target.Pos().Y, Z2: target.PosZ() + zTargetOffset,
			}
			pHeading, pPitch := pLine.Heading(), pLine.Pitch()
			if a.u.TurretAngle() == pHeading && a.u.Pitch() == pPitch {
				return bt.Success, nil
			}

			log.Debugf("[%s] %0.1f|%0.1f turretToTarget @ %s", a.u.ID(), geom.Degrees(a.u.TurretAngle()), geom.Degrees(pPitch), target.ID())
			a.u.SetTargetTurretAngle(pHeading)
			a.u.SetTargetPitch(pPitch)
			// TODO: return failure if not even close to target angle
			return bt.Success, nil
		},
	)
}

func (a *AIBehavior) VelocityToMax() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			if a.u.Velocity() == a.u.MaxVelocity() {
				return bt.Success, nil
			}

			log.Debugf("[%s] %0.1f -> velocityMax", a.u.ID(), a.u.Velocity()*model.VELOCITY_TO_KPH)
			a.u.SetTargetVelocity(a.u.MaxVelocity())
			return bt.Success, nil
		},
	)
}

func (a *AIBehavior) DetermineForcedWithdrawal() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			if a.u.StructurePoints() > 0.2*a.u.MaxStructurePoints() {
				return bt.Failure, nil
			}
			log.Debugf("[%s] -> determineForcedWithdrawal", a.u.ID())
			return bt.Success, nil
		},
	)
}

func (a *AIBehavior) TurnToWithdraw() bt.Node {
	var withdrawHeading float64 = 1.57
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			if a.u.Heading() == withdrawHeading {
				return bt.Success, nil
			}

			log.Debugf("[%s] %0.1f -> turnToWithdraw", a.u.ID(), geom.Degrees(a.u.Heading()))
			a.u.SetTargetHeading(withdrawHeading)
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
