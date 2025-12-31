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

	menuOrder []Menu
	menuIndex int
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
		menuOrder: []Menu{
			mapSelect,
			unitSelect,
			enemySelect,
			launchBriefing,
		},
	}
	scene.SetMenu(mapSelect)
	return scene
}

func (s *InstantActionScene) SetMenu(m Menu) {
	s.Game.menu = m
}

func (s *InstantActionScene) getMenu() Menu {
	if s.menuIndex >= 0 && s.menuIndex < len(s.menuOrder) {
		return s.menuOrder[s.menuIndex]
	}
	return nil
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

	s.menuIndex -= 1

	prevMenu := s.getMenu()
	if s.menuIndex < 0 {
		// back to main menu
		g.scene = NewMainMenuScene(g)
	} else {
		// back to previous menu
		s.SetMenu(prevMenu)
	}
}

func (s *InstantActionScene) next() {
	g := s.Game

	// check actions for current menu
	currentMenu := s.getMenu()
	if currentMenu == s.launchBriefing {
		// launch game scene into map mission
		if g.player == nil {
			// pick player unit at random
			g.SetPlayerUnit(g.RandomUnit(model.MechResourceType))
		}
		g.scene = NewGameScene(g)
	}

	s.menuIndex += 1

	// check actions for next menu
	nextMenu := s.getMenu()
	if nextMenu == s.launchBriefing {
		// prepare briefing menu for display
		if s.playerUnitSelect.selectedUnit == nil {
			// set player unit nil to indicate randomized pick for launch briefing
			g.player = nil
		} else {
			g.SetPlayerUnit(s.playerUnitSelect.selectedUnit)
		}

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
	}

	if s.menuIndex < 0 {
		// back to main menu
		g.scene = NewMainMenuScene(g)
	} else if s.menuIndex < len(s.menuOrder) {
		// proceed to next menu
		s.SetMenu(nextMenu)
	}
}
