package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/joelschutz/stagehand"
)

type InterstitialScene struct {
	BaseScene
	clr     color.NRGBA
	trigger stagehand.SceneTransitionTrigger
	timeout float64
}

func NewInterstitialScene(clr color.NRGBA, trigger stagehand.SceneTransitionTrigger, timeout float64) *InterstitialScene {
	return &InterstitialScene{
		clr:     clr,
		trigger: trigger,
		timeout: timeout,
	}
}

func (s *InterstitialScene) PreTransition(toScene stagehand.Scene[SceneState]) SceneState {
	return s.BaseScene.PreTransition(toScene)
}

func (s *InterstitialScene) PostTransition(state SceneState, fromScene stagehand.Scene[SceneState]) {
	s.state.Timer = s.timeout
	s.BaseScene.PostTransition(state, fromScene)
}

func (s *InterstitialScene) Update() error {
	if s.state.OnTransition {
		// no further updates during transition
		return nil
	}

	keys := inpututil.AppendJustPressedKeys(nil)
	keyPressed := len(keys) > 0

	var buttonPressed bool
	gamepadIDs := ebiten.AppendGamepadIDs(nil)
	if len(gamepadIDs) > 0 {
		for _, g := range gamepadIDs {
			buttons := inpututil.AppendJustPressedGamepadButtons(g, nil)
			if len(buttons) > 0 {
				buttonPressed = true
				break
			}
		}
	}

	skip := keyPressed || buttonPressed || inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	s.state.Timer -= 1 / float64(ebiten.TPS())
	if skip || s.state.Timer <= 0 {
		s.sm.ProcessTrigger(s.trigger)
	}
	return nil
}

func (s *InterstitialScene) Draw(screen *ebiten.Image) {
	screen.Fill(s.clr)
}
