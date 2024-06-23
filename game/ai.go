package game

import (
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"

	bt "github.com/joeycumines/go-behaviortree"
	log "github.com/sirupsen/logrus"
)

type AIHandler struct {
	g  *Game
	ai []*AIBehavior
}

type AIBehavior struct {
	bt.Node
	g *Game
	u model.Unit
}

func NewAIHandler(g *Game) *AIHandler {
	units := g.getSpriteUnits()
	unitAI := make([]*AIBehavior, 0, len(units))
	aiHandler := &AIHandler{
		g: g,
	}

	for _, u := range units {
		if u.Team() < 0 {
			unitAI = append(unitAI, aiHandler.NewFriendlyAI(u))
		} else {
			unitAI = append(unitAI, aiHandler.NewAI(u))
		}
	}
	aiHandler.ai = unitAI

	return aiHandler
}

func (h *AIHandler) NewAI(u model.Unit) *AIBehavior {
	a := &AIBehavior{g: h.g, u: u}
	a.Node = a.ForcedWithdrawal()
	return a
}

func (h *AIHandler) NewFriendlyAI(u model.Unit) *AIBehavior {
	a := &AIBehavior{g: h.g, u: u}
	a.Node = bt.New(
		bt.Sequence,
		a.ChaseTarget(),
		a.ShootTarget(),
	)
	return a
}

func (h *AIHandler) Update() {
	for _, a := range h.ai {
		if a.u.IsDestroyed() || a.u.Powered() != model.POWER_ON {
			continue
		}
		_, err := a.Tick()
		if err != nil {
			log.Error(err)
		}
	}
}

func (a *AIBehavior) ChaseTarget() bt.Node {
	return bt.New(
		bt.Sequence,
		a.determineTargetStatus(),
		a.moveToTarget(),
	)
}

func (a *AIBehavior) moveToTarget() bt.Node {
	return bt.New(
		bt.Sequence,
		a.turnToTarget(),
		a.velocityToMax(),
	)
}

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

			//log.Debugf("[%s] %0.1f -> turnToTarget @ %s", a.u.ID(), geom.Degrees(a.u.Heading()), target.ID())
			a.u.SetTargetHeading(pHeading)
			return bt.Success, nil
		},
	)
}

func (a *AIBehavior) ShootTarget() bt.Node {
	return bt.New(
		bt.Sequence,
		a.determineTargetStatus(),
		a.turretToTarget(),
		a.fireWeapons(),
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

			//log.Debugf("[%s] %0.1f|%0.1f turretToTarget @ %s", a.u.ID(), geom.Degrees(a.u.TurretAngle()), geom.Degrees(pPitch), target.ID())
			a.u.SetTargetTurretAngle(pHeading)
			a.u.SetTargetPitch(pPitch)
			// TODO: return failure if not even close to target angle
			return bt.Success, nil
		},
	)
}

func (a *AIBehavior) fireWeapons() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			target := model.EntityUnit(a.u.Target())
			if target == nil {
				return bt.Failure, nil
			}

			weaponFired := false
			for _, w := range a.u.Armament() {
				// TODO: only fire weapons within range
				// TODO: only fire some weapons, try not to overheat (much)
				// TODO: weapon convergence toward target
				if a.g.fireUnitWeapon(a.u, w) {
					weaponFired = true
				}
			}

			if weaponFired {
				// TODO: // illuminate source sprite unit firing the weapon
				// combat.go: sprite.SetIlluminationPeriod(5000, 0.35)
				//log.Debugf("[%s] fireWeapons @ %s", a.u.ID(), target.ID())
				return bt.Success, nil
			}
			return bt.Failure, nil
		},
	)
}

func (a *AIBehavior) determineTargetStatus() bt.Node {
	return bt.New(
		bt.Sequence,
		a.hasTarget(),
		a.targetIsAlive(),
		// a.targetInRange(),
	)
}

func (a *AIBehavior) hasTarget() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			if a.u.Target() != nil {
				return bt.Success, nil
			}

			// TODO: create separate node for selecting a new target based on some criteria
			units := a.g.getSpriteUnits()
			// TODO: enemy units need to be able to target player unit
			for _, t := range units {
				if t.IsDestroyed() || t.Team() == a.u.Team() {
					continue
				}

				//log.Debugf("[%s] hasTarget == %s", a.u.ID(), t.ID())
				a.u.SetTarget(t)
				return bt.Success, nil
			}
			return bt.Failure, nil
		},
	)
}

func (a *AIBehavior) targetIsAlive() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			if a.u.Target() != nil {
				if !a.u.Target().IsDestroyed() {
					return bt.Success, nil
				}
				a.u.SetTarget(nil)
			}
			return bt.Failure, nil
		},
	)
}

func (a *AIBehavior) ForcedWithdrawal() bt.Node {
	return bt.New(
		bt.Selector,
		a.attemptWithdraw(),
		a.moveToWithdrawArea(),
	)
}

func (a *AIBehavior) attemptWithdraw() bt.Node {
	return bt.New(
		bt.Sequence,
		a.inWithdrawArea(),
		a.withdraw(),
	)
}

func (a *AIBehavior) moveToWithdrawArea() bt.Node {
	return bt.New(
		bt.Sequence,
		a.turnToWithdraw(),
		a.velocityToMax(),
	)
}

func (a *AIBehavior) turnToWithdraw() bt.Node {
	var withdrawHeading float64 = 1.57
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			if a.u.Heading() == withdrawHeading {
				return bt.Success, nil
			}

			//log.Debugf("[%s] %0.1f -> turnToWithdraw", a.u.ID(), geom.Degrees(a.u.Heading()))
			a.u.SetTargetHeading(withdrawHeading)
			return bt.Success, nil
		},
	)
}

func (a *AIBehavior) velocityToMax() bt.Node {
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

func (a *AIBehavior) inWithdrawArea() bt.Node {
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

func (a *AIBehavior) withdraw() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			//log.Debugf("[%s] -> withdraw", a.u.ID())

			// TODO: unit safely escapes
			a.u.SetStructurePoints(0)
			return bt.Success, nil
		},
	)
}
