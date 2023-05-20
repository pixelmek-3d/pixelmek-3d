package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type IntroScene struct {
	Game *Game
}

func NewIntroScene(g *Game) *IntroScene {
	return &IntroScene{
		Game: g,
	}
}

func (s *IntroScene) Update() error {

	return nil
}

func (s *IntroScene) Draw(screen *ebiten.Image) {

}
