package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"

	log "github.com/sirupsen/logrus"
)

type InstantActionScene struct {
	Game             *Game
	mapSelect        *MapMenu
	playerUnitSelect *UnitMenu
	enemyUnitSelect  *UnitMenu
	launchBriefing   *LaunchMenu
}

func NewInstantActionScene(g *Game) Scene {
	mapSelect := createMapMenu(g)
	unitSelect := createUnitMenu(g, PlayerUnitMenu)
	enemySelect := createUnitMenu(g, EnemyUnitMenu)
	launchBriefing := createLaunchMenu(g)

	scene := &InstantActionScene{
		Game:             g,
		mapSelect:        mapSelect,
		playerUnitSelect: unitSelect,
		enemyUnitSelect:  enemySelect,
		launchBriefing:   launchBriefing,
	}
	scene.SetMenu(mapSelect)
	return scene
}

func (s *InstantActionScene) SetMenu(m Menu) {
	s.Game.menu = m
}

func (s *InstantActionScene) Update() error {
	g := s.Game

	if g.input.ActionIsJustPressed(ActionBack) {
		s.back()
	}

	// update the menu
	g.menu.Update()

	return nil
}

func (s *InstantActionScene) Draw(screen *ebiten.Image) {
	g := s.Game

	// draw menu
	g.menu.Draw(screen)
}

func (s *InstantActionScene) back() {
	g := s.Game

	switch g.menu {
	case s.launchBriefing:
		// back to enemy unit select
		s.SetMenu(s.enemyUnitSelect)

	case s.enemyUnitSelect:
		// back to player unit select
		s.SetMenu(s.playerUnitSelect)

	case s.playerUnitSelect:
		// back to map select
		s.SetMenu(s.mapSelect)

	case s.mapSelect:
		fallthrough
	default:
		// back to main menu
		g.scene = NewMainMenuScene(g)
	}
}

func (s *InstantActionScene) next() {
	g := s.Game

	switch g.menu {
	case s.launchBriefing:
		// launch game scene into map mission
		if g.player == nil {
			// pick player unit at random
			g.SetPlayerUnit(g.RandomUnit(model.MechResourceType))
		}

		g.scene = NewGameScene(g)

	case s.enemyUnitSelect:
		// to pre-launch briefing after enemy player unit and map
		opts := &InstantActionMissionOpts{enemies: make([]model.Unit, 0, 1)}
		if s.enemyUnitSelect.selectedUnit != nil {
			// set enemy unit for mission spawning
			opts.enemies = append(opts.enemies, s.enemyUnitSelect.selectedUnit)
		}

		mission, err := g.LoadInstantActionFromMap(s.mapSelect.selectedMap, opts)
		if err != nil {
			log.Error("Error loading mission from map: ", s.mapSelect.selectedMap.Name)
			log.Error(err)
			exit(1)
		}
		s.launchBriefing.loadBriefing(mission)
		s.SetMenu(s.launchBriefing)

	case s.playerUnitSelect:
		// to enemy unit select after setting player unit
		if s.playerUnitSelect.selectedUnit == nil {
			// set player unit nil to indicate randomized pick for launch briefing
			g.player = nil
		} else {
			g.SetPlayerUnit(s.playerUnitSelect.selectedUnit)
		}

		s.SetMenu(s.enemyUnitSelect)

	case s.mapSelect:
		// to unit select
		s.SetMenu(s.playerUnitSelect)

	default:
		// back to main menu
		g.scene = NewMainMenuScene(g)
	}
}
