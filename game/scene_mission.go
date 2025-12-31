package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
)

type MissionScene struct {
	Game             *Game
	missionSelect    *MissionMenu
	playerUnitSelect *UnitMenu
	launchBriefing   *LaunchMenu

	menuOrder []Menu
	menuIndex int
}

func NewMissionScene(g *Game) Scene {
	missionSelect := createMissionMenu(g)
	unitSelect := createUnitMenu(g, PlayerUnitMenu)
	launchBriefing := createLaunchMenu(g)

	scene := &MissionScene{
		Game:             g,
		missionSelect:    missionSelect,
		playerUnitSelect: unitSelect,
		launchBriefing:   launchBriefing,
		menuOrder: []Menu{
			missionSelect,
			unitSelect,
			launchBriefing,
		},
	}
	scene.SetMenu(missionSelect)
	return scene
}

func (s *MissionScene) SetMenu(m Menu) {
	s.Game.menu = m
}

func (s *MissionScene) getMenu() Menu {
	if s.menuIndex >= 0 && s.menuIndex < len(s.menuOrder) {
		return s.menuOrder[s.menuIndex]
	}
	return nil
}

func (s *MissionScene) Update() error {
	g := s.Game

	if g.input.ActionIsJustPressed(ActionBack) {
		s.back()
	}

	// update the menu
	g.menu.Update()

	return nil
}

func (s *MissionScene) Draw(screen *ebiten.Image) {
	g := s.Game

	// draw menu
	g.menu.Draw(screen)
}

func (s *MissionScene) back() {
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

func (s *MissionScene) next() {
	g := s.Game

	// check actions for current menu
	currentMenu := s.getMenu()
	if currentMenu == s.launchBriefing {
		// launch game scene into mission
		if g.player == nil {
			// pick player unit at random
			g.SetPlayerUnit(g.RandomUnit(model.MechResourceType))
		}

		g.mission = s.missionSelect.selectedMission
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

		s.launchBriefing.loadBriefing(s.missionSelect.selectedMission)
		s.SetMenu(s.launchBriefing)
	}

	if s.menuIndex < 0 {
		// back to main menu
		g.scene = NewMainMenuScene(g)
	} else if s.menuIndex < len(s.menuOrder) {
		// proceed to next menu
		s.SetMenu(nextMenu)
	}
}
