package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/common"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"

	log "github.com/sirupsen/logrus"
)

type InstantActionScene struct {
	Game             *Game
	mapSelect        *MapMenu
	missionOpts      *InstantActionMissionOpts
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
		missionOpts:      &InstantActionMissionOpts{},
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

	if g.input.ActionIsJustPressed(ActionMenuBack) {
		// go back only if no open windows
		if w := s.Game.menu.CloseWindow(); w == nil {
			s.back()
		}
	}

	// update the menu
	g.menu.Update()

	return nil
}

func (s *InstantActionScene) Draw(screen *ebiten.Image) {
	g := s.Game

	// draw menu
	g.menu.Draw(screen)

	// draw FPS display
	g.drawFPS(&render.DrawHudOptions{Screen: screen})
}

func (s *InstantActionScene) back() {
	g := s.Game

	// check actions for current menu
	currentMenu := s.getMenu()
	switch currentMenu {
	case s.mapSelect:
		g.mission = nil
	}

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
	switch currentMenu {
	case s.mapSelect:
		if g.mission != nil && len(g.mission.Teams) > 0 {
			// set previously selected camo
			s.missionOpts.teams = g.mission.Teams
		} else if len(s.missionOpts.teams) == 0 {
			// set random camo selections
			camoFiles, _ := model.ListCamoFilenames()
			if len(camoFiles) > 0 {
				var enemyCamo, friendlyCamo string
				enemyCamo, camoFiles = common.PopRandom(camoFiles)
				if len(camoFiles) > 0 {
					friendlyCamo, _ = common.PopRandom(camoFiles)
				}
				s.missionOpts.teams = map[int]*model.MissionTeam{
					0:  {Camo: enemyCamo},
					-1: {Camo: friendlyCamo},
				}
			}
		}

		_, err := g.LoadInstantActionFromMap(s.mapSelect.selectedMap, s.missionOpts)
		if err != nil {
			log.Error("Error loading mission from map: ", s.mapSelect.selectedMap.Name)
			log.Error(err)
			exit(1)
		}

	case s.enemyUnitSelect:
		if s.enemyUnitSelect.selectedUnit != nil {
			// set enemy unit for mission spawning
			s.missionOpts.enemies = []model.Unit{s.enemyUnitSelect.selectedUnit}
		}

	case s.launchBriefing:
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
	switch nextMenu {
	case s.launchBriefing:
		// prepare briefing menu for display
		if s.playerUnitSelect.selectedUnit == nil {
			// set player unit nil to indicate randomized pick for launch briefing
			g.player = nil
		} else {
			g.SetPlayerUnit(s.playerUnitSelect.selectedUnit)
		}

		// load mission from map again to refresh changes from previous menus
		_, err := g.LoadInstantActionFromMap(s.mapSelect.selectedMap, s.missionOpts)
		if err != nil {
			log.Error("Error loading mission from map: ", s.mapSelect.selectedMap.Name)
			log.Error(err)
			exit(1)
		}
		s.launchBriefing.loadBriefing(g.mission)
	}

	if s.menuIndex < 0 {
		// back to main menu
		g.scene = NewMainMenuScene(g)
	} else if s.menuIndex < len(s.menuOrder) {
		// proceed to next menu
		s.SetMenu(nextMenu)
	}
}
