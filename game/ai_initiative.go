package game

import (
	"math/rand"
	"sort"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	log "github.com/sirupsen/logrus"
)

const (
	// AI_INITIATIVE_SLOTS represents the limited number of initiative time slots an AI can be updated (every N ticks)
	AI_INITIATIVE_SLOTS = 4

	// AI_INITIATIVE_TIMER is how many ticks between rerolling initiative slots
	AI_INITIATIVE_TIMER = 300
)

type AIInitiative struct {
	ai    []*AIBehavior
	stack [][]*AIBehavior
	timer uint
}

func NewAIInitiative(aiList []*AIBehavior) *AIInitiative {
	a := &AIInitiative{ai: aiList}
	a.roll()
	return a
}

func (n *AIInitiative) clear() {
	n.stack = make([][]*AIBehavior, AI_INITIATIVE_SLOTS)
	for i := 0; i < AI_INITIATIVE_SLOTS; i++ {
		n.stack[i] = make([]*AIBehavior, 0, 1)
	}
}

func (n *AIInitiative) roll() {
	n.clear()

	type initiativeRoll struct {
		ai   *AIBehavior
		roll float64
	}

	// determine initiative for each AI (higher is better)
	rolls := make([]*initiativeRoll, 0, len(n.ai))
	for _, ai := range n.ai {
		if ai.u.IsDestroyed() {
			continue
		}
		rolls = append(rolls, &initiativeRoll{
			ai:   ai,
			roll: rand.Float64(),
		})

		// set flag to indicate a new initiative order has started
		ai.newInitiative = true
	}

	sort.Slice(rolls, func(i, j int) bool {
		return rolls[i].roll > rolls[j].roll
	})

	// distribute AI evenly amongst the initiative slots
	slot := 0
	numPerSlot := len(rolls) / AI_INITIATIVE_SLOTS
	numRemainder := len(rolls) % AI_INITIATIVE_SLOTS
	for len(rolls) > 0 {
		for i := 0; i < numPerSlot; i++ {
			n.stack[slot] = append(n.stack[slot], rolls[i].ai)
		}
		rolls = rolls[numPerSlot:]

		if numRemainder > 0 {
			n.stack[slot] = append(n.stack[slot], rolls[0].ai)
			rolls = rolls[1:]
			numRemainder--
		}

		slot++
	}

	if log.GetLevel() == log.DebugLevel {
		log.Debug("rolled initiative:")
		for i := 0; i < len(n.stack); i++ {
			for j := 0; j < len(n.stack[i]); j++ {
				log.Debugf("  %d.%d - %v", i, j, n.stack[i][j].u.ID())
			}
		}
	}
}

func (n *AIInitiative) Next() []*AIBehavior {
	slot := n.timer % AI_INITIATIVE_SLOTS

	if n.timer >= AI_INITIATIVE_TIMER-1 {
		n.timer = 0
		n.roll()
	} else {
		n.timer++
	}

	return n.stack[slot]
}

// UpdateForNewInitiativeSet performs certain updates that only occur
// at the beginning of a new initiative set
func (a *AIBehavior) UpdateForNewInitiativeSet() {
	a.initiativeTargetAcquisition()
	a.initiativeIdealWeaponsDistance()
	a.newInitiative = false
}

// initiativeTargetAcquisition evaluates if the unit should select a new target
// at the beginning of a new initiative set
func (a *AIBehavior) initiativeTargetAcquisition() {
	if a.u.Target() != nil {
		stayOnTarget := true
		if a.newInitiative {
			// TODO: better criteria for when to change to another target
			stayOnTarget = false
		}
		if !stayOnTarget {
			a.u.SetTarget(nil)
		}
	}
}

// initiativeIdealWeaponsDistance used weighted weapon distances to determine
// what distance to keep from target for ideal weapon ranges
func (a *AIBehavior) initiativeIdealWeaponsDistance() {
	type weaponWeighting struct {
		distance  float64
		weighting float64
	}
	weights := make([]weaponWeighting, 0, len(a.u.Armament()))
	for _, w := range a.u.Armament() {
		if model.WeaponAmmoCount(w) <= 0 {
			// only consider weapons with ammo remaining
			continue
		}
		dist := w.Distance() / model.METERS_PER_UNIT
		dps := w.Damage() / w.MaxCooldown()

		weighting := weaponWeighting{
			distance:  dist,
			weighting: dps,
		}
		weights = append(weights, weighting)
	}
	// TODO: short term - just get distance of highest weighted weapon and halve it or something
	// TODO: long term - use weapon weighting average/median distance?
}
