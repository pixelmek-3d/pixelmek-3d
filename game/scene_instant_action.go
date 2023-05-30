package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
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
		// launch mission
		g.scene = NewMissionScene(g)
	case s.unitSelect:
		// to pre-launch briefing
		s.SetMenu(s.launchBriefing)
	case s.missionSelect:
		// to unit select
		s.SetMenu(s.unitSelect)
	default:
		// back to main menu
		g.scene = NewMenuScene(g)
	}
}
