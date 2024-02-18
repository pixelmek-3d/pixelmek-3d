package game

import (
	"fmt"
	"time"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	log "github.com/sirupsen/logrus"
)

type ObjectivesStatus int

const (
	OBJECTIVES_IN_PROGRESS ObjectivesStatus = iota
	OBJECTIVES_COMPLETED
	OBJECTIVES_FAILED
)

type ObjectivesHandler struct {
	_objectives *model.MissionObjectives
	current     map[Objective]time.Time
	completed   map[Objective]time.Time
	failed      map[Objective]time.Time
}

type Objective interface {
	Update(*Game)
	Current() bool
	Completed() bool
	Failed() bool
}

type BasicObjective struct {
	completed bool
	failed    bool
}

type DestroyObjective struct {
	*BasicObjective
	_objective *model.MissionDestroyObjectives
	_units     []model.Unit
}

func (o *BasicObjective) Current() bool {
	return !o.completed && !o.failed
}

func (o *BasicObjective) Completed() bool {
	return o.completed && !o.failed
}

func (o *BasicObjective) Failed() bool {
	return o.completed && o.failed
}

func NewObjectivesHandler(g *Game, objectives *model.MissionObjectives) *ObjectivesHandler {
	o := &ObjectivesHandler{
		_objectives: objectives,
		current:     make(map[Objective]time.Time),
		completed:   make(map[Objective]time.Time),
		failed:      make(map[Objective]time.Time),
	}

	all_units := g.getSpriteUnits()

	var iTime time.Time
	for _, modelObjective := range objectives.Destroy {
		unitID := modelObjective.Unit

		if len(unitID) > 0 {
			objective_units := make([]model.Unit, 0, 1)
			for _, unit := range all_units {
				if unitID == unit.ID() {
					objective_units = append(objective_units, unit)
				}
			}

			destroyObjective := &DestroyObjective{
				BasicObjective: &BasicObjective{},
				_objective:     modelObjective,
				_units:         objective_units,
			}
			o.current[destroyObjective] = iTime
		}
	}

	return o
}

func (o *ObjectivesHandler) Update(g *Game) {
	updated := time.Now()
	for objective := range o.current {
		objective.Update(g)
		if objective.Current() {
			o.current[objective] = updated
			continue
		}

		delete(o.current, objective)
		switch {
		case objective.Completed():
			o.completed[objective] = updated
		case objective.Failed():
			o.failed[objective] = updated
		default:
			panic(fmt.Sprintf("unexpected objective state for %v", objective))
		}
	}
}

func (o *ObjectivesHandler) Status() ObjectivesStatus {
	switch {
	case len(o.failed) > 0:
		return OBJECTIVES_FAILED
	case len(o.current) == 0 && len(o.completed) > 0:
		return OBJECTIVES_COMPLETED
	}
	return OBJECTIVES_IN_PROGRESS
}

func (o *DestroyObjective) Update(g *Game) {
	all_destroyed := true
	for _, unit := range o._units {
		if !unit.IsDestroyed() {
			all_destroyed = false
			break
		}
	}

	if all_destroyed {
		log.Debugf("destroy objective completed: %s", o._objective.Unit)
		o.completed = true
	}
}
