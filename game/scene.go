package game

import "github.com/hajimehoshi/ebiten/v2"

type Scene interface {
	Update() error
	Draw(*ebiten.Image)
}

type SceneEffect interface {
	Update() error
	Draw(*ebiten.Image)
}

type SceneTransition interface {
	Update() error
	Draw(*ebiten.Image)
}
