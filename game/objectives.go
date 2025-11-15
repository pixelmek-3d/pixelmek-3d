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
	objectives *model.MissionObjectives
	current    map[Objective]time.Time
	completed  map[Objective]time.Time
	failed     map[Objective]time.Time

	objectivesText string
}

type Objective interface {
	Update(*Game)
	Current() bool
	Completed() bool
	Failed() bool
	Text() string
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
	objective *model.MissionDestroyObjectives
	units     []model.Unit
}

type ProtectObjective struct {
	*BasicObjective
	objective *model.MissionProtectObjectives
	units     []model.Unit
}

type VisitObjective struct {
	*BasicObjective
	objective *model.MissionNavVisit
	nav       *model.NavPoint
}

type DustoffObjective struct {
	*BasicObjective
	objective     *model.MissionNavDustoff
	nav           *model.NavPoint
	verifyDustoff bool
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
		objectives: objectives,
		current:    make(map[Objective]time.Time),
		completed:  make(map[Objective]time.Time),
		failed:     make(map[Objective]time.Time),
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
					unit.SetObjective(model.ProtectUnitObjective)
					protectUnits = append(protectUnits, unit)
				}
			}

			protectObjective := &ProtectObjective{
				BasicObjective: &BasicObjective{},
				objective:      modelObjective,
				units:          protectUnits,
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
					if g.IsFriendly(g.player, unit) {
						// prevent units on player team from being a destroy objective
						if !all {
							log.Errorf("unit ID found on player team as destroy objective: %s", unit.ID())
						}
						continue
					}
					unit.SetObjective(model.DestroyUnitObjective)
					destroyUnits = append(destroyUnits, unit)
				}
			}

			destroyObjective := &DestroyObjective{
				BasicObjective: &BasicObjective{},
				objective:      modelObjective,
				units:          destroyUnits,
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

			objectiveNav.SetObjective(model.NavVisitObjective)
			visitObjective := &VisitObjective{
				BasicObjective: &BasicObjective{},
				objective:      modelObjective,
				nav:            objectiveNav,
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

			objectiveNav.SetObjective(model.NavDustoffObjective)
			visitObjective := &DustoffObjective{
				BasicObjective: &BasicObjective{},
				objective:      modelObjective,
				nav:            objectiveNav,
			}
			o.current[visitObjective] = iTime
		}
	}

	o.updateObjectivesText()

	return o
}

func (o *ObjectivesHandler) Update(g *Game) {
	update := false
	currTime := time.Now()

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
			o.current[objective] = currTime
			continue
		}

		delete(o.current, objective)
		switch {
		case objective.Completed():
			o.completed[objective] = currTime
		case objective.Failed():
			o.failed[objective] = currTime
		default:
			panic(fmt.Sprintf("unexpected objective state for %v", objective))
		}

		update = true
	}

	// special handling for Nav.Dustoff which cannot be completed until after all destroy/visit, where applicable
	if len(objsDustoff) > 0 {
		dustoffReady := (len(objsDestroy) == 0 && len(objsVisit) == 0)
		dustoffComplete := false
		for _, objective := range objsDustoff {
			if !dustoffReady && objective.verifyDustoff {
				// reset dustoff nav site visited flag until other objectives are complete
				objective.nav.SetVisited(false)
				objective.verifyDustoff = false
				continue
			}

			if dustoffReady && objective.verifyDustoff {
				log.Debugf("nav dustoff objective completed: %s", objective.nav.Name)
				dustoffComplete = true
				break
			}
		}

		if dustoffComplete {
			update = true
			// only one nav needed to complete dustoff objective
			for _, objective := range objsDustoff {
				objective.completed = true
			}
		}
	}

	// special handling for Protect.Unit which cannot be completed until after all destroy/visit/dustoff, where applicable
	if len(objsProtect) > 0 && len(objsDestroy) == 0 && len(objsVisit) == 0 && len(objsDustoff) == 0 {
		update = true
		for _, objective := range objsProtect {
			log.Debugf("protect objective completed: %s", objective.objective.Unit)
			objective.completed = true
		}
	}

	// special handling for player objective of staying alive
	if g.player.IsDestroyed() {
		update = true
		o.failed[&PlayerAliveObjective{
			BasicObjective: &BasicObjective{
				failed: true,
			},
		}] = currTime
	}

	if update {
		o.updateObjectivesText()
	}
}

func (o *ObjectivesHandler) updateObjectivesText() {
	oText := ""

	if len(o.current) > 0 {
		for objective := range o.current {
			oText += objective.Text() + "\n"
		}
	}

	if len(o.failed) > 0 {
		oText += "\n*FAILED*\n"
		for objective := range o.failed {
			oText += objective.Text() + "\n"
		}
	}

	if len(o.completed) > 0 {
		oText += "\n-COMPLETED-\n"
		for objective := range o.completed {
			oText += objective.Text() + "\n"
		}
	}

	o.objectivesText = oText
}

func (o *ObjectivesHandler) Text() string {
	return o.objectivesText
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
func (o *PlayerAliveObjective) Text() string {
	return ""
}

func (o *DestroyObjective) Update(g *Game) {
	allDestroyed := true
	for _, unit := range o.units {
		if !unit.IsDestroyed() {
			allDestroyed = false
			break
		}
	}

	if !allDestroyed {
		return
	}

	if o.objective.Waves {
		log.Debug("spawning next wave")
		unit := g.spawnUnit("fire_moth_prime") // FIXME: use unit selected by user
		if unit != nil {
			o.units = append(o.units, unit)
			return
		}
	}

	destroyedStr := o.objective.Unit
	if o.objective.All {
		destroyedStr = "all"
	}
	log.Debugf("destroy objective completed: %s", destroyedStr)
	o.completed = true
}
func (o *DestroyObjective) Text() string {
	if o.objective.All {
		return `Destroy All Enemies`
	}
	return `Destroy ` + o.objective.Unit
}

func (o *ProtectObjective) Update(g *Game) {
	allAlive := true
	for _, unit := range o.units {
		if unit.IsDestroyed() {
			allAlive = false
			break
		}
	}

	if !allAlive {
		log.Debugf("protect objective failed: %s", o.objective.Unit)
		o.failed = true
	}
}
func (o *ProtectObjective) Text() string {
	return `Protect ` + o.objective.Unit
}

func (o *VisitObjective) Update(g *Game) {
	if o.nav.Visited() {
		log.Debugf("nav visit objective completed: %s", o.nav.Name)
		o.completed = true
	}
}
func (o *VisitObjective) Text() string {
	return `Visit Nav ` + o.objective.Name
}

func (o *DustoffObjective) Update(g *Game) {
	// special handling for Dustoff to verify all non-dustoff objectives must first be completed
	if o.nav.Visited() {
		o.verifyDustoff = true
	}
}
func (o *DustoffObjective) Text() string {
	return `Dustoff Nav ` + o.objective.Name
}
