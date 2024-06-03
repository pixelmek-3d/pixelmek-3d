package game

import (
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joelschutz/stagehand"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/transitions"
)

const (
	SplashTrigger stagehand.SceneTransitionTrigger = iota
	LaunchGameTrigger
	MissionDebriefTrigger
	MainMenuTrigger
	InstantActionTrigger
)

type SceneState struct {
	Timer        float64
	OnTransition bool
}

type BaseScene struct {
	game   *Game
	bounds image.Rectangle
	state  SceneState
	sm     *stagehand.SceneDirector[SceneState]
}

func (s *BaseScene) Layout(w, h int) (int, int) {
	s.bounds = image.Rect(0, 0, w, h)
	return w, h
}

func (s *BaseScene) Load(st SceneState, sm stagehand.SceneController[SceneState]) {
	s.state = st
	s.sm = sm.(*stagehand.SceneDirector[SceneState])
}

func (s *BaseScene) Unload() SceneState {
	return s.state
}

func (s *BaseScene) PreTransition(toScene stagehand.Scene[SceneState]) SceneState {
	s.state.OnTransition = true
	s.game.scene = toScene
	return s.state
}

func (s *BaseScene) PostTransition(state SceneState, fromScene stagehand.Scene[SceneState]) {
	s.state.OnTransition = false
}

type SceneEffect interface {
	Update() error
	Draw(screen *ebiten.Image)
}

type SceneShader interface {
	Update() error
	Draw(screen, img *ebiten.Image)
}

func (g *Game) initScenes() {
	// create scene director, scenes, triggers, and transitions
	state := SceneState{Timer: SPLASH_TIMEOUT}
	ebitenSplashScene := NewEbitengineSplashScene(g)
	gopherSplashScene := NewGopherSplashScene(g)
	introScene := NewIntroScene(g)

	mainMenuScene := NewMenuScene(g)
	instantActionScene := NewInstantActionScene(g)
	gameScene := NewGameScene(g)
	debriefScene := NewMissionDebriefScene(g)

	transDissolve := transitions.NewDissolveTransition[SceneState](time.Second * time.Duration(3))
	transFade := transitions.NewFadeTransition[SceneState](time.Second * time.Duration(4))
	transPixelize := transitions.NewPixelizeTransition[SceneState](time.Second * time.Duration(2))
	transSlideUp := stagehand.NewDurationTimedSlideTransition[SceneState](stagehand.BottomToTop, time.Millisecond*time.Duration(500))
	transSlideDown := stagehand.NewDurationTimedSlideTransition[SceneState](stagehand.TopToBottom, time.Millisecond*time.Duration(500))
	rs := map[stagehand.Scene[SceneState]][]stagehand.Directive[SceneState]{
		ebitenSplashScene: {
			stagehand.Directive[SceneState]{
				Dest:       gopherSplashScene,
				Trigger:    SplashTrigger,
				Transition: transDissolve,
			},
		},
		gopherSplashScene: {
			stagehand.Directive[SceneState]{
				Dest:       introScene,
				Trigger:    SplashTrigger,
				Transition: transPixelize,
			},
		},
		introScene: {
			stagehand.Directive[SceneState]{
				Dest:       mainMenuScene,
				Trigger:    SplashTrigger,
				Transition: transPixelize,
			},
		},
		mainMenuScene: {
			stagehand.Directive[SceneState]{
				Dest:       instantActionScene,
				Trigger:    InstantActionTrigger,
				Transition: transSlideDown,
			},
		},
		instantActionScene: {
			stagehand.Directive[SceneState]{
				Dest:       gameScene,
				Trigger:    LaunchGameTrigger,
				Transition: transFade,
			},
			stagehand.Directive[SceneState]{
				Dest:       mainMenuScene,
				Trigger:    MainMenuTrigger,
				Transition: transSlideUp,
			},
		},
		gameScene: {
			stagehand.Directive[SceneState]{
				Dest:       debriefScene,
				Trigger:    MissionDebriefTrigger,
				Transition: transSlideDown,
			},
		},
		debriefScene: {
			stagehand.Directive[SceneState]{
				Dest:       mainMenuScene,
				Trigger:    MainMenuTrigger,
				Transition: transSlideUp,
			},
		},
	}

	g.sm = stagehand.NewSceneDirector[SceneState](ebitenSplashScene, state, rs)
}
