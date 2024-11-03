package game

import (
	"math/rand"
	"sort"

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
