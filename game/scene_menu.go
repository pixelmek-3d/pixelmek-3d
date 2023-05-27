package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type MenuScene struct {
	Game *Game
}

func NewMenuScene(g *Game) *MenuScene {
	g.menu = createMainMenu(g)
	return &MenuScene{
		Game: g,
	}
}

func (s *MenuScene) Update() error {
	g := s.Game

	// update the menu
	g.menu.Update()

	return nil
}

func (s *MenuScene) Draw(screen *ebiten.Image) {
	g := s.Game

	// draw menu
	g.menu.Draw(screen)
}
