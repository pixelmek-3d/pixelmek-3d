package game

import (
	"github.com/pixelmek-3d/pixelmek-3d/game/model"

	bt "github.com/joeycumines/go-behaviortree"
	log "github.com/sirupsen/logrus"
)

func (a *AIBehavior) HasTarget() bt.Node {
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

				log.Debugf("[%s] hasTarget == %s", a.u.ID(), t.ID())
				a.u.SetTarget(t)
				return bt.Success, nil
			}
			return bt.Failure, nil
		},
	)
}

func (a *AIBehavior) TargetIsAlive() bt.Node {
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

func (a *AIBehavior) FireWeapons() bt.Node {
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
