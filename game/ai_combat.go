package game

import (
	"math"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"

	bt "github.com/joeycumines/go-behaviortree"
)

func (a *AIBehavior) HasTarget() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		if a.u.Target() != nil {
			stayOnTarget := true
			if a.newInitiative {
				// only evaluate choosing a new target at beginning of new initiative set
				// TODO: better criteria for when to change to another target
				stayOnTarget = false
			}
			if stayOnTarget {
				return bt.Success, nil
			}
		}

		// TODO: create separate node for selecting a new target based on some criteria?

		// TODO: different detection range for different units
		pUnits := a.g.getProximitySpriteUnits(a.u.Pos(), 1000/model.METERS_PER_UNIT)
		for _, p := range pUnits {
			t := p.unit
			if t == a.u || t.IsDestroyed() || a.g.IsFriendly(a.u, t) {
				continue
			}

			// log.Debugf("[%s] hasTarget == %s", a.u.ID(), t.ID())
			if a.u.Target() != t {
				// reset AI settings for previous targets
				a.gunnery.Reset()

				a.u.SetTarget(t)
			}

			return bt.Success, nil
		}
		a.u.SetTarget(nil)
		return bt.Failure, nil
	}
}

func (a *AIBehavior) TargetIsAlive() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		if a.u.Target() != nil {
			if !a.u.Target().IsDestroyed() {
				return bt.Success, nil
			}
			a.u.SetTarget(nil)
		}
		return bt.Failure, nil
	}
}

func (a *AIBehavior) FireWeapons() func([]bt.Node) (bt.Status, error) {
	return func(_ []bt.Node) (bt.Status, error) {
		if a.gunnery.ticksSinceFired < math.MaxUint {
			// use number of AI ticks since last weapon fired to influence random decision to not fire this tick
			a.gunnery.ticksSinceFired += 1
		}

		target := model.EntityUnit(a.u.Target())
		if target == nil {
			return bt.Failure, nil
		}

		// chance to fire this tick gradually increases as number of ticks without firing goes up
		chanceToFire := float64(a.gunnery.ticksSinceFired) / (model.TICKS_PER_SECOND / AI_INITIATIVE_SLOTS)
		if chanceToFire < 1 {
			r := model.RandFloat64In(0, 1.0, a.rng)
			if r > chanceToFire {
				return bt.Failure, nil
			}
		}

		targetDist := model.EntityDistance2D(a.u, target)

		unitHeat := a.u.Heat()
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

		if len(readyWeapons) == 0 {
			return bt.Failure, nil
		}

		// check walls for line of sight to target
		if !a.g.lineOfSight(a.u, target) {
			// log.Debugf("[%s] wall in LOS to %s", a.u.ID(), target.ID())
			return bt.Failure, nil
		}

		// TODO: sort ready weapons based on which is most ideal to fire given the current circumstances

		// check for angle/pitch proximity to target center mass
		if a.gunnery.targetLeadPos != nil {
			targetLeadLine := &geom3d.Line3d{
				X1: a.u.Pos().X, Y1: a.u.Pos().Y, Z1: a.u.PosZ() + a.u.CockpitOffset().Y,
				X2: a.gunnery.targetLeadPos.X, Y2: a.gunnery.targetLeadPos.Y, Z2: target.PosZ() + target.CollisionHeight()/2,
			}

			// use target collision size for proximity check
			targetLeadLine2D := &geom.Line{
				X1: a.u.Pos().X, Y1: a.u.Pos().Y,
				X2: a.gunnery.targetLeadPos.X, Y2: a.gunnery.targetLeadPos.Y,
			}
			targetLeadDist2D := targetLeadLine2D.Distance()

			// find the angle using opposite/adjacent, then use hypotenuse length to get the line to the edge of target collision radius
			targetProximityAngle2D := math.Atan(target.CollisionRadius() / targetLeadDist2D)
			targetProximityLength2D := model.Hypotenuse(targetLeadDist2D, target.CollisionRadius())
			targetProximityLine2D := geom.LineFromAngle(a.u.Pos().X, a.u.Pos().Y, targetLeadLine2D.Angle()+targetProximityAngle2D, targetProximityLength2D)

			targetProximityLine := &geom3d.Line3d{
				X1: a.u.Pos().X, Y1: a.u.Pos().Y, Z1: a.u.PosZ() + a.u.CockpitOffset().Y,
				X2: targetProximityLine2D.X2, Y2: targetProximityLine2D.Y2, Z2: target.PosZ() + target.CollisionHeight(),
			}

			proximityHeading := math.Abs(model.AngleDistance(targetProximityLine.Heading(), targetLeadLine.Heading())) * 1.1
			proximityPitch := math.Abs(model.AngleDistance(targetProximityLine.Pitch(), targetLeadLine.Pitch())) * 1.1

			deltaHeading := math.Abs(model.AngleDistance(a.u.TurretAngle(), targetLeadLine.Heading()))
			deltaPitch := math.Abs(model.AngleDistance(a.u.Pitch(), targetLeadLine.Pitch()))
			if deltaHeading > proximityHeading || deltaPitch > proximityPitch {
				// log.Debugf("[%s] not in proximity to [%s] (pH=%0.3f, pP=%0.3f) @ (dH=%0.3f, dP=%0.3f)", a.u.ID(), target.ID(), geom.Degrees(proximityHeading), geom.Degrees(proximityPitch), geom.Degrees(deltaHeading), geom.Degrees(deltaPitch))
				return bt.Failure, nil
			}
		}

		// check for friendly units in line of fire to target position
		units := a.g.getSpriteUnits()
		// use angle/pitch for weapons line of fire checks, not current target position
		targetLine := geom.LineFromAngle(a.u.Pos().X, a.u.Pos().Y, a.u.TurretAngle(), targetDist)
		for _, s := range units {
			if s == a.u || s.IsDestroyed() || !a.g.IsFriendly(a.u, s) {
				continue
			}

			zDiff := target.PosZ() - a.u.PosZ()
			if (zDiff > 0 && s.PosZ() < a.u.PosZ()) || (zDiff < 0 && s.PosZ() > a.u.PosZ()) {
				// TODO: use a 3-Dimensional line of fire check
				continue
			}

			// TODO: make sure player unit is checked when same team as AI unit

			sCollisionCircle := geom.Circle{X: s.Pos().X, Y: s.Pos().Y, Radius: s.CollisionRadius()}
			if len(geom.LineCircleIntersection(targetLine, sCollisionCircle, true)) > 0 {
				// wait to fire until line of fire is not blocked by friendly
				// log.Debugf("[%s] friendly in LOS to %s", a.u.ID(), target.ID())
				return bt.Failure, nil
			}
		}

		var weaponFired model.Weapon
		for _, w := range readyWeapons {
			if unitHeat+w.Heat() >= a.u.MaxHeat() {
				// only fire the weapon if it will not lead to overheating
				// log.Debugf("[%s] weapon (%s) too hot to fire @ %s", a.u.ID(), w.ShortName(), target.ID())
				continue
			}

			// only fire the weapon at the same time if similar to other weapons fired this cycle
			if weaponFired != nil && !similarWeapons(weaponFired, w) {
				continue
			}

			// TODO: weapon convergence toward target
			if a.g.fireUnitWeapon(a.u, w) {
				weaponFired = w
				unitHeat = a.u.Heat()
			}
		}

		if weaponFired != nil {
			// log.Debugf("[%s] fireWeapons @ %s", a.u.ID(), target.ID())
			a.gunnery.ticksSinceFired = 0
			return bt.Success, nil
		}
		return bt.Failure, nil
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

func similarWeapons(w1, w2 model.Weapon) bool {
	if w1 == nil || w2 == nil {
		return false
	}

	// TODO: consider different sized weapons within classifications as dissimilar
	return w1.Classification() == w2.Classification()
}
