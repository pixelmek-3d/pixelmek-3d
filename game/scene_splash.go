package game

import (
	"image"

	"github.com/joelschutz/stagehand"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/effects"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	SPLASH_TIMEOUT = 2.5
)

type SplashScene struct {
	BaseScene
	splash       *SplashScreen
	splashRect   image.Rectangle
	bufferScreen *ebiten.Image
}

type SplashScreen struct {
	img    *ebiten.Image
	screen *ebiten.Image
	geoM   ebiten.GeoM
	effect SceneEffect
	shader SceneShader
}

func NewEbitengineSplashScene(g *Game) *SplashScene {
	// Ebitengine splash
	var splash *SplashScreen
	splashRect := g.uiRect()

	im, _, err := resources.NewImageFromFile("textures/ebitengine_splash.png")
	if err == nil {
		geoM := splashGeoM(im, splashRect)
		// tOpts := &transitions.TransitionOptions{
		// 	InDuration:   SPLASH_TIMEOUT * 2 / 5,
		// 	HoldDuration: SPLASH_TIMEOUT * 1.5 / 5,
		// 	OutDuration:  SPLASH_TIMEOUT * 1.5 / 5,
		// }

		splash = newSplashScreen(g)
		splash.img = im
		//splash.transition = transitions.NewDissolve(splash.screen, tOpts, ebiten.GeoM{})
		splash.geoM = geoM
	}

	return &SplashScene{
		BaseScene: BaseScene{
			game: g,
		},
		splash:     splash,
		splashRect: splashRect,
		//splashTimer:  SPLASH_TIMEOUT,
		bufferScreen: ebiten.NewImage(g.screenWidth, g.screenHeight),
	}
}

func NewGopherSplashScene(g *Game) *SplashScene {
	// Golang Gopher splash
	var splash *SplashScreen
	splashRect := g.uiRect()

	im, _, err := resources.NewImageFromFile("textures/gopher_space.png")
	if err == nil {
		geoM := splashGeoM(im, splashRect)
		// tOpts := &transitions.TransitionOptions{
		// 	InDuration:   SPLASH_TIMEOUT * 2 / 5,
		// 	HoldDuration: SPLASH_TIMEOUT * 1.5 / 5,
		// 	OutDuration:  SPLASH_TIMEOUT * 1.5 / 5,
		// }

		splash = newSplashScreen(g)
		splash.img = im
		splash.effect = effects.NewStars(g.screenWidth, g.screenHeight)
		//splash.transition = transitions.NewFade(splash.screen, tOpts, ebiten.GeoM{})
		splash.geoM = geoM
	}

	return &SplashScene{
		BaseScene: BaseScene{
			game: g,
		},
		splash:     splash,
		splashRect: splashRect,
		//splashTimer:  SPLASH_TIMEOUT,
		bufferScreen: ebiten.NewImage(g.screenWidth, g.screenHeight),
	}
}

func newSplashScreen(g *Game) *SplashScreen {
	return &SplashScreen{
		screen: ebiten.NewImage(g.screenWidth, g.screenHeight),
	}
}

func splashGeoM(splash *ebiten.Image, splashRect image.Rectangle) ebiten.GeoM {
	sW, sH := float64(splash.Bounds().Dx()), float64(splash.Bounds().Dy())
	bX, bY, bW, bH := splashRect.Min.X, splashRect.Min.Y, splashRect.Dx(), splashRect.Dy()

	// scale image to only take up a portion of the space
	sScale := 0.75 * float64(bH) / sH
	sW, sH = sW*sScale, sH*sScale
	sX, sY := float64(bX)+float64(bW)/2-sW/2, float64(bY)+float64(bH)/2-sH/2

	geoM := ebiten.GeoM{}
	geoM.Scale(sScale, sScale)
	geoM.Translate(sX, sY)
	return geoM
}

func (s *SplashScene) PreTransition(toScene stagehand.Scene[SceneState]) SceneState {
	return s.BaseScene.PreTransition(toScene)
}

func (s *SplashScene) PostTransition(state SceneState, fromScene stagehand.Scene[SceneState]) {
	s.state.Timer = SPLASH_TIMEOUT
	s.BaseScene.PostTransition(state, fromScene)
}

func (s *SplashScene) Update() error {
	splash := s.splash
	if splash == nil {
		return nil
	}
	if splash.effect != nil {
		splash.effect.Update()
	}
	if splash.shader != nil {
		splash.shader.Update()
	}
	// if splash.transition != nil {
	// 	splash.transition.Update()
	// }

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
		s.sm.ProcessTrigger(SplashTrigger)
	}

	return nil
}

func (s *SplashScene) Draw(screen *ebiten.Image) {
	// draw effect as splash image background
	splash := s.splash
	if splash == nil {
		return
	}
	splash.screen.Clear()
	s.bufferScreen.Clear()

	if splash.effect != nil {
		// draw effect
		splash.effect.Draw(splash.screen)
	}

	if splash.img != nil {
		// draw splash image
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM = splash.geoM
		splash.screen.DrawImage(splash.img, op)
	}

	if splash.shader != nil {
		// draw shader effect with splash screen to buffer
		splash.shader.Draw(s.bufferScreen, splash.screen)
	} else {
		// draw splash screen to buffer
		s.bufferScreen.DrawImage(splash.screen, nil)
	}

	// if splash.transition != nil {
	// 	// draw transition from buffer
	// 	splash.transition.SetImage(s.bufferScreen)
	// 	splash.transition.Draw(screen)
	// } else {

	// draw buffer directly to screen
	screen.DrawImage(s.bufferScreen, nil)
}
