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
}

func NewMissionScene(g *Game) Scene {
	missionSelect := createMissionMenu(g)
	unitSelect := createUnitMenu(g)
	launchBriefing := createLaunchMenu(g)

	scene := &MissionScene{
		Game:             g,
		missionSelect:    missionSelect,
		playerUnitSelect: unitSelect,
		launchBriefing:   launchBriefing,
	}
	scene.SetMenu(missionSelect)
	return scene
}

func (s *MissionScene) SetMenu(m Menu) {
	s.Game.menu = m
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

	switch g.menu {
	case s.launchBriefing:
		// back to unit select
		s.SetMenu(s.playerUnitSelect)

	case s.playerUnitSelect:
		// back to mission select
		s.SetMenu(s.missionSelect)

	case s.missionSelect:
		fallthrough
	default:
		// back to main menu
		g.scene = NewMainMenuScene(g)
	}
}

func (s *MissionScene) next() {
	g := s.Game

	switch g.menu {
	case s.launchBriefing:
		// launch game scene into mission
		if g.player == nil {
			// pick player unit at random
			g.SetPlayerUnit(g.RandomUnit(model.MechResourceType))
		}

		g.mission = s.missionSelect.selectedMission
		g.scene = NewGameScene(g)

	case s.playerUnitSelect:
		// to pre-launch briefing after setting player unit and mission
		if s.playerUnitSelect.selectedUnit == nil {
			// set player unit nil to indicate randomized pick for launch briefing
			g.player = nil
		} else {
			g.SetPlayerUnit(s.playerUnitSelect.selectedUnit)
		}

		s.launchBriefing.loadBriefing(s.missionSelect.selectedMission)
		s.SetMenu(s.launchBriefing)

	case s.missionSelect:
		// to unit select
		s.SetMenu(s.playerUnitSelect)

	default:
		// back to main menu
		g.scene = NewMainMenuScene(g)
	}
}
