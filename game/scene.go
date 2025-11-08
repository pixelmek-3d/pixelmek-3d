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

func (g *Game) SetInitialSceneFunc(sceneFunc func(g *Game) Scene) {
	g.initSceneFunc = sceneFunc
}

func (g *Game) SetScene(scene Scene) {
	g.scene = scene
}

func (g *Game) StopSceneTransition() {
	if g.scene == nil {
		return
	}
	switch g.scene.(type) {
	case *MenuScene:
		g.scene.(*MenuScene).transition = nil
	case *GameScene:
		g.scene.(*GameScene).transition = nil
	}
}
