package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
)

type InstantActionScene struct {
	Game           *Game
	missionSelect  *MissionMenu
	unitSelect     *UnitMenu
	launchBriefing *LaunchMenu
}

func NewInstantActionScene(g *Game) *InstantActionScene {
	missionSelect := createMissionMenu(g)
	unitSelect := createUnitMenu(g)
	launchBriefing := createLaunchMenu(g)

	scene := &InstantActionScene{
		Game:           g,
		missionSelect:  missionSelect,
		unitSelect:     unitSelect,
		launchBriefing: launchBriefing,
	}
	scene.SetMenu(missionSelect)
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
		// back to unit select
		s.SetMenu(s.unitSelect)

	case s.unitSelect:
		// back to mission select
		s.SetMenu(s.missionSelect)

	case s.missionSelect:
		fallthrough
	default:
		// back to main menu
		g.scene = NewMenuScene(g)
	}
}

func (s *InstantActionScene) next() {
	g := s.Game

	switch g.menu {
	case s.launchBriefing:
		// launch mission scene
		if g.player == nil {
			// pick player unit at random
			g.SetPlayerUnit(g.randomUnit(model.MechResourceType))
		}

		g.scene = NewMissionScene(g)

	case s.unitSelect:
		// to pre-launch briefing after setting player unit and mission
		g.mission = s.missionSelect.selectedMission

		if s.unitSelect.selectedUnit == nil {
			// set player unit nil to indicate randomized pick for launch briefing
			g.player = nil
		} else {
			g.SetPlayerUnit(s.unitSelect.selectedUnit)
		}

		s.launchBriefing.loadBriefing()
		s.SetMenu(s.launchBriefing)

	case s.missionSelect:
		// to unit select
		s.SetMenu(s.unitSelect)

	default:
		// back to main menu
		g.scene = NewMenuScene(g)
	}
}
