package game

import (
	"github.com/harbdog/raycaster-go/geom"
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
		unitAI = append(unitAI, aiHandler.NewAI(u))
	}
	aiHandler.ai = unitAI

	return aiHandler
}

func (h *AIHandler) NewAI(u model.Unit) *AIBehavior {
	a := &AIBehavior{g: h.g, u: u}
	a.Node = a.ForcedWithdrawal()
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

func (a *AIBehavior) ForcedWithdrawal() bt.Node {
	return bt.New(
		bt.Sequence,
		a.turnToWithdraw(),
		a.velocityToMax(),
		a.withdraw(),
	)
}

func (a *AIBehavior) turnToWithdraw() bt.Node {
	var withdrawHeading float64 = 1.57
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			if a.u.Heading() == withdrawHeading {
				return bt.Success, nil
			}

			log.Debugf("[%s]%s-%s: %0.1f -> turnToWithdraw", a.u.ID(), a.u.Name(), a.u.Variant(), geom.Degrees(a.u.Heading()))
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

			log.Debugf("[%s]%s-%s: %0.1f -> velocityMax", a.u.ID(), a.u.Name(), a.u.Variant(), a.u.Velocity()*model.VELOCITY_TO_KPH)
			a.u.SetTargetVelocity(a.u.MaxVelocity())
			return bt.Success, nil
		},
	)
}

func (a *AIBehavior) withdraw() bt.Node {
	return bt.New(
		bt.Sequence,
		a.inWithdrawPosition(),
		a.eject(), // TODO: unit safely escapes
	)
}

func (a *AIBehavior) inWithdrawPosition() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			posX, posY := a.u.Pos().X, a.u.Pos().Y
			delta := 1.5
			if posX > delta && posX < float64(a.g.mapWidth)-delta &&
				posY > delta && posY < float64(a.g.mapHeight)-delta {
				return bt.Failure, nil
			}

			log.Debugf("[%s]%s-%s: %v -> inWithdrawPosition", a.u.ID(), a.u.Name(), a.u.Variant(), a.u.Pos())
			return bt.Success, nil
		},
	)
}

func (a *AIBehavior) eject() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			log.Debugf("[%s]%s-%s: -> eject", a.u.ID(), a.u.Name(), a.u.Variant())
			a.u.SetStructurePoints(0)
			return bt.Success, nil
		},
	)
}
