package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type InstantActionScene struct {
	Game *Game
	// mapSelect  *MissionMenu // TODO: MapMenu
	unitSelect *UnitMenu
	// enemySelect *EnemyMenu // TODO: EnemyMenu
	// launchBriefing *LaunchMenu // TODO: launch briefing menu
}

func NewInstantActionScene(g *Game) Scene {
	// mapSelect
	unitSelect := createUnitMenu(g)
	// enemySelect
	// launchBriefing

	scene := &InstantActionScene{
		Game:       g,
		unitSelect: unitSelect,
	}
	scene.SetMenu(unitSelect)
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
	// TODO:
	default:
		// back to main menu
		g.scene = NewMenuScene(g)
	}
}

func (s *InstantActionScene) next() {
	// FIXME: refactor back/next to work properly when not MissionScene
	g := s.Game

	switch g.menu {
	// TODO:
	default:
		// back to main menu
		g.scene = NewMenuScene(g)
	}
}
