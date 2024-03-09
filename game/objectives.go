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

type PlayerAliveObjective struct {
	*BasicObjective
}

type DestroyObjective struct {
	*BasicObjective
	_objective *model.MissionDestroyObjectives
	_units     []model.Unit
}

type ProtectObjective struct {
	*BasicObjective
	_objective *model.MissionProtectObjectives
	_units     []model.Unit
}

type VisitObjective struct {
	*BasicObjective
	_objective *model.MissionNavVisit
	_nav       *model.NavPoint
}

type DustoffObjective struct {
	*BasicObjective
	_objective     *model.MissionNavDustoff
	_nav           *model.NavPoint
	_verifyDustoff bool
}

func (o *BasicObjective) Current() bool {
	return !o.completed && !o.failed
}

func (o *BasicObjective) Completed() bool {
	return o.completed && !o.failed
}

func (o *BasicObjective) Failed() bool {
	return o.failed
}

func NewObjectivesHandler(g *Game, objectives *model.MissionObjectives) *ObjectivesHandler {
	if objectives == nil {
		// default objectives: destroy all
		objectives = &model.MissionObjectives{
			Destroy: []*model.MissionDestroyObjectives{
				{
					All: true,
				},
			},
		}
	}

	o := &ObjectivesHandler{
		_objectives: objectives,
		current:     make(map[Objective]time.Time),
		completed:   make(map[Objective]time.Time),
		failed:      make(map[Objective]time.Time),
	}

	all_units := g.getSpriteUnits()
	var iTime time.Time

	protectUnits := make([]model.Unit, 0, 16)
	destroyUnits := make([]model.Unit, 0, 16)

	for _, modelObjective := range objectives.Protect {
		unitID := modelObjective.Unit

		if len(unitID) > 0 {
			for _, unit := range all_units {
				if unitID == unit.ID() {
					protectUnits = append(protectUnits, unit)
				}
			}

			protectObjective := &ProtectObjective{
				BasicObjective: &BasicObjective{},
				_objective:     modelObjective,
				_units:         protectUnits,
			}
			o.current[protectObjective] = iTime
		}
	}

	for _, modelObjective := range objectives.Destroy {
		all := modelObjective.All
		unitID := modelObjective.Unit

		if all || len(unitID) > 0 {
			for _, unit := range all_units {
				if all || (len(unitID) > 0 && unitID == unit.ID()) {
					if model.InArray(protectUnits, unit) {
						// prevent protected units from also being a destroy objective
						if !all {
							log.Errorf("same unit ID found in protect and destroy objectives: %s", unit.ID())
						}
						continue
					}
					destroyUnits = append(destroyUnits, unit)
				}
			}

			destroyObjective := &DestroyObjective{
				BasicObjective: &BasicObjective{},
				_objective:     modelObjective,
				_units:         destroyUnits,
			}
			o.current[destroyObjective] = iTime
		}
	}

	if objectives.Nav != nil {
		for _, modelObjective := range objectives.Nav.Visit {
			navName := modelObjective.Name
			if len(navName) == 0 {
				continue
			}
			var objectiveNav *model.NavPoint
			for _, nav := range g.mission.NavPoints {
				if navName == nav.Name {
					objectiveNav = nav
					break
				}
			}
			if objectiveNav == nil {
				log.Errorf("visit objective nav point not found: %s", navName)
				continue
			}

			objectiveNav.SetIsObjective(true)
			visitObjective := &VisitObjective{
				BasicObjective: &BasicObjective{},
				_objective:     modelObjective,
				_nav:           objectiveNav,
			}
			o.current[visitObjective] = iTime
		}

		for _, modelObjective := range objectives.Nav.Dustoff {
			navName := modelObjective.Name
			if len(navName) == 0 {
				continue
			}
			var objectiveNav *model.NavPoint
			for _, nav := range g.mission.NavPoints {
				if navName == nav.Name {
					objectiveNav = nav
					break
				}
			}
			if objectiveNav == nil {
				log.Errorf("dustoff objective nav point not found: %s", navName)
				continue
			}

			objectiveNav.SetIsDustoff(true)
			visitObjective := &DustoffObjective{
				BasicObjective: &BasicObjective{},
				_objective:     modelObjective,
				_nav:           objectiveNav,
			}
			o.current[visitObjective] = iTime
		}
	}

	return o
}

func (o *ObjectivesHandler) Update(g *Game) {
	updated := time.Now()
	objsDestroy := make([]*DestroyObjective, 0, 16)
	objsProtect := make([]*ProtectObjective, 0, 16)
	objsVisit := make([]*VisitObjective, 0, 4)
	objsDustoff := make([]*DustoffObjective, 0, 1)
	for objective := range o.current {
		switch objective := objective.(type) {
		case *DestroyObjective:
			objsDestroy = append(objsDestroy, objective)
		case *ProtectObjective:
			objsProtect = append(objsProtect, objective)
		case *VisitObjective:
			objsVisit = append(objsVisit, objective)
		case *DustoffObjective:
			objsDustoff = append(objsDustoff, objective)
		}

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

	// special handling for Nav.Dustoff which cannot be completed until after all destroy/visit, where applicable
	if len(objsDustoff) > 0 {
		dustoffReady := (len(objsDestroy) == 0 && len(objsVisit) == 0)
		dustoffComplete := false
		for _, objective := range objsDustoff {
			if !dustoffReady && objective._verifyDustoff {
				// reset dustoff nav site visited flag until other objectives are complete
				objective._nav.SetVisited(false)
				objective._verifyDustoff = false
				continue
			}

			if dustoffReady && objective._verifyDustoff {
				log.Debugf("nav dustoff objective completed: %s", objective._nav.Name)
				dustoffComplete = true
				break
			}
		}

		if dustoffComplete {
			// only one nav needed to complete dustoff objective
			for _, objective := range objsDustoff {
				objective.completed = true
			}
		}
	}

	// special handling for Protect.Unit which cannot be completed until after all destroy/visit/dustoff, where applicable
	if len(objsProtect) > 0 && len(objsDestroy) == 0 && len(objsVisit) == 0 && len(objsDustoff) == 0 {
		for _, objective := range objsProtect {
			log.Debugf("protect objective completed: %s", objective._objective.Unit)
			objective.completed = true
		}
	}

	// special handling for player objective of staying alive
	if g.player.IsDestroyed() {
		o.failed[&PlayerAliveObjective{
			BasicObjective: &BasicObjective{
				failed: true,
			},
		}] = updated
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

func (o *PlayerAliveObjective) Update(g *Game) {}

func (o *DestroyObjective) Update(g *Game) {
	allDestroyed := true
	for _, unit := range o._units {
		if !unit.IsDestroyed() {
			allDestroyed = false
			break
		}
	}

	if allDestroyed {
		destroyedStr := o._objective.Unit
		if o._objective.All {
			destroyedStr = "all"
		}
		log.Debugf("destroy objective completed: %s", destroyedStr)
		o.completed = true
	}
}

func (o *ProtectObjective) Update(g *Game) {
	allAlive := true
	for _, unit := range o._units {
		if unit.IsDestroyed() {
			allAlive = false
			break
		}
	}

	if !allAlive {
		log.Debugf("protect objective failed: %s", o._objective.Unit)
		o.failed = true
	}
}

func (o *VisitObjective) Update(g *Game) {
	if o._nav.Visited() {
		log.Debugf("nav visit objective completed: %s", o._nav.Name)
		o.completed = true
	}
}

func (o *DustoffObjective) Update(g *Game) {
	// special handling for Dustoff to verify all non-dustoff objectives must first be completed
	if o._nav.Visited() {
		o._verifyDustoff = true
	}
}
