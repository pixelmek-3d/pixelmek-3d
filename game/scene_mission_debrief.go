package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type MissionDebriefScene struct {
	Game           *Game
	missionDebrief *DebriefMenu
}

func NewMissionDebriefScene(g *Game) Scene {
	missionDebrief := createDebriefMenu(g)

	scene := &MissionDebriefScene{
		Game:           g,
		missionDebrief: missionDebrief,
	}
	scene.SetMenu(missionDebrief)
	return scene
}

func (s *MissionDebriefScene) SetMenu(m Menu) {
	s.Game.menu = m
}

func (s *MissionDebriefScene) Update() error {
	g := s.Game

	if g.input.ActionIsJustPressed(ActionBack) {
		s.back()
	}

	// update the menu
	g.menu.Update()

	return nil
}

func (s *MissionDebriefScene) Draw(screen *ebiten.Image) {
	g := s.Game

	// draw menu
	g.menu.Draw(screen)
}

func (s *MissionDebriefScene) back() {
	g := s.Game

	// back to main menu
	g.scene = NewMenuScene(g)
}
