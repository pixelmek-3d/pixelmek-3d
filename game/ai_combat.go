package game

import (
	"math"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
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

			// reset AI settings for previous targets
			a.gunnery.Reset()
			a.piloting.Reset()

			// TODO: create separate node for selecting a new target based on some criteria
			units := a.g.getSpriteUnits()

			// TODO: enemy units need to be able to target player unit
			// units = append(units, a.g.player)

			for _, t := range units {
				if t == a.u || t.IsDestroyed() || t.Team() == a.u.Team() {
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
			counter = AI_FIRE_WEAPONS_COUNTER_MIN

			target := model.EntityUnit(a.u.Target())
			if target == nil {
				return bt.Failure, nil
			}

			// check for angle/pitch proximity to target
			if a.gunnery.targetLeadPos != nil {
				targetLeadLine := &geom3d.Line3d{
					X1: a.u.Pos().X, Y1: a.u.Pos().Y, Z1: a.u.PosZ() + a.u.CockpitOffset().Y,
					X2: a.gunnery.targetLeadPos.X, Y2: a.gunnery.targetLeadPos.Y, Z2: target.PosZ() + target.CollisionHeight()/2,
				}

				deltaHeading := math.Abs(model.AngleDistance(a.u.TurretAngle(), targetLeadLine.Heading()))
				deltaPitch := math.Abs(model.AngleDistance(a.u.Pitch(), targetLeadLine.Pitch()))
				if deltaHeading > 0.05 || deltaPitch > 0.1 {
					//log.Debugf("[%s] not in proximity to [%s] @ (dH=%0.3f, dP=%0.3f)", a.u.ID(), target.ID(), geom.Degrees(deltaHeading), geom.Degrees(deltaPitch))
					return bt.Failure, nil
				}
			}

			// check walls for line of sight to target
			if !a.g.lineOfSight(a.u, target) {
				return bt.Failure, nil
			}

			// use angle/pitch for weapons line of fire checks, not current target position
			targetDist := model.EntityDistance2D(a.u, target)
			targetLine := geom.LineFromAngle(a.u.Pos().X, a.u.Pos().Y, a.u.TurretAngle(), targetDist)

			// check for friendly units in line of fire to target position
			units := a.g.getSpriteUnits()
			for _, s := range units {
				// TODO: make sure player unit is checked when same team as AI unit
				if s == a.u || s.IsDestroyed() || s.Team() != a.u.Team() {
					continue
				}

				zDiff := target.PosZ() - a.u.PosZ()
				if (zDiff > 0 && s.PosZ() < a.u.PosZ()) || (zDiff < 0 && s.PosZ() > a.u.PosZ()) {
					// TODO: use a 3-Dimensional line of fire check
					continue
				}

				sCollisionCircle := geom.Circle{X: s.Pos().X, Y: s.Pos().Y, Radius: s.CollisionRadius()}
				if len(geom.LineCircleIntersection(targetLine, sCollisionCircle, true)) > 0 {
					// wait to fire until line of fire is not blocked by friendly
					return bt.Failure, nil
				}
			}

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
				if targetDist > 1.25*w.Distance()/model.METERS_PER_UNIT {
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
			return bt.Failure, nil
		},
	)
}
