package game

import (
	"github.com/hajimehoshi/ebiten/v2"
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

	// update the menu
	g.menu.Update()

	return nil
}

func (s *InstantActionScene) Draw(screen *ebiten.Image) {
	g := s.Game

	// draw menu
	g.menu.Draw(screen)
}
