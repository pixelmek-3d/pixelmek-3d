package game

import (
	"github.com/pixelmek-3d/pixelmek-3d/game/model"

	bt "github.com/joeycumines/go-behaviortree"
	log "github.com/sirupsen/logrus"
)

const (
	AI_FIRE_WEAPONS_COUNTER_MIN = 20
	AI_FIRE_WEAPONS_COUNTER_MAX = 60
)

func (a *AIBehavior) HasTarget() bt.Node {
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			if a.u.Target() != nil {
				// TODO: criteria for when to change to another target
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
	// randomly skip a number of ticks to not attempt a firing solution every tick
	counter := AI_FIRE_WEAPONS_COUNTER_MIN + model.Randish.Intn(AI_FIRE_WEAPONS_COUNTER_MAX-AI_FIRE_WEAPONS_COUNTER_MIN)
	return bt.New(
		func(children []bt.Node) (bt.Status, error) {
			if counter > 0 {
				counter--
				return bt.Success, nil
			}
			target := model.EntityUnit(a.u.Target())
			if target == nil {
				return bt.Failure, nil
			}

			targetDist := model.EntityDistance(a.u, target) * model.METERS_PER_UNIT

			readyWeapons := make([]model.Weapon, 0, len(a.u.Armament()))
			for _, w := range a.u.Armament() {
				if w.Cooldown() > 0 {
					// only weapons not on cooldown
					continue
				}
				if model.WeaponAmmoCount(w) <= 0 {
					// only weapons with ammo remaining
					continue
				}
				if targetDist > 1.25*w.Distance() {
					// only weapons within range
					continue
				}

				readyWeapons = append(readyWeapons, w)
			}

			unitHeat := a.u.Heat()

			weaponFired := false
			for _, w := range readyWeapons {
				if unitHeat+w.Heat() >= a.u.MaxHeat() {
					// only fire the weapon if it will not lead to overheating
					continue
				}

				// TODO: instead of alpha striking or firing each weapon as soon as it is ready, have a small random delay (except for machine guns)?

				// TODO: weapon convergence toward target
				if a.g.fireUnitWeapon(a.u, w) {
					weaponFired = true
					unitHeat = a.u.Heat()
				}
			}

			if weaponFired {
				//log.Debugf("[%s] fireWeapons @ %s", a.u.ID(), target.ID())
				counter = AI_FIRE_WEAPONS_COUNTER_MIN + model.Randish.Intn(AI_FIRE_WEAPONS_COUNTER_MAX-AI_FIRE_WEAPONS_COUNTER_MIN)
				return bt.Success, nil
			}
			counter = AI_FIRE_WEAPONS_COUNTER_MIN
			return bt.Failure, nil
		},
	)
}
