package game

import "github.com/hajimehoshi/ebiten/v2"

type Scene interface {
	Update() error
	Draw(screen *ebiten.Image)
}

type SceneEffect interface {
	Update() error
	Draw(screen *ebiten.Image)
}

type SceneShader interface {
	Update() error
	Draw(screen, img *ebiten.Image)
}

type SceneTransition interface {
	Completed() bool
	SetImage(img *ebiten.Image)
	Update() error
	Draw(screen *ebiten.Image)
}
