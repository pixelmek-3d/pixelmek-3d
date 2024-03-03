package game

import (
	"image"

	"github.com/pixelmek-3d/pixelmek-3d/game/render/effects"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/transitions"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	SPLASH_TIMEOUT = 5.0
)

type IntroScene struct {
	Game        *Game
	splashes    []*SplashScreen
	splashRect  image.Rectangle
	splashIndex int
	splashTimer float64
}

type SplashScreen struct {
	img        *ebiten.Image
	screen     *ebiten.Image
	geoM       ebiten.GeoM
	effect     SceneEffect
	shader     SceneShader
	transition SceneTransition
}

func NewIntroScene(g *Game) *IntroScene {
	// load intro splash images
	var splashes = make([]*SplashScreen, 0)
	splashRect := g.uiRect()

	// Ebitengine splash
	im, _, err := resources.NewImageFromFile("textures/ebitengine_splash.png")
	if err == nil {
		geoM := splashGeoM(im, splashRect)
		tOpts := &transitions.TransitionOptions{
			InDuration:   SPLASH_TIMEOUT * 2 / 5,
			HoldDuration: SPLASH_TIMEOUT * 1.5 / 5,
			OutDuration:  SPLASH_TIMEOUT * 1.5 / 5,
		}
		splash := NewSplashScreen(g)
		splash.img = nil
		splash.transition = transitions.NewDissolve(im, tOpts, geoM)
		splash.geoM = geoM
		splashes = append(splashes, splash)
	}

	// Golang Gopher splash
	im, _, err = resources.NewImageFromFile("textures/gopher_space.png")
	if err == nil {
		geoM := splashGeoM(im, splashRect)
		splash := NewSplashScreen(g)
		splash.img = im
		splash.effect = effects.NewStars(g.screenWidth, g.screenHeight)
		//splash.shader = effects.NewCRT()
		splash.geoM = geoM
		splashes = append(splashes, splash)
	}

	return &IntroScene{
		Game:        g,
		splashes:    splashes,
		splashRect:  splashRect,
		splashTimer: SPLASH_TIMEOUT,
	}
}

func NewSplashScreen(g *Game) *SplashScreen {
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

func (s *IntroScene) currentSplash() *SplashScreen {
	if s.splashIndex < 0 || s.splashIndex >= len(s.splashes) {
		return nil
	}
	return s.splashes[s.splashIndex]
}

func (s *IntroScene) Update() error {
	splash := s.currentSplash()
	if splash.effect != nil {
		splash.effect.Update()
	}
	if splash.transition != nil {
		splash.transition.Update()
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
	if skip {
		s.splashIndex += 1
		s.splashTimer = SPLASH_TIMEOUT
	}

	// use timer to move on if no input
	s.splashTimer -= 1 / float64(ebiten.TPS())
	if s.splashTimer <= 0 {
		s.splashIndex += 1
		s.splashTimer = SPLASH_TIMEOUT
	}

	if s.splashIndex >= len(s.splashes) {
		// TODO: make a last splash that requires key/button press to move on to main menu
		s.Game.scene = NewMenuScene(s.Game)
	}

	return nil
}

func (s *IntroScene) Draw(screen *ebiten.Image) {
	// draw effect as splash image background
	splash := s.currentSplash()
	splash.screen.Clear()

	if splash.effect != nil {
		// draw effect
		splash.effect.Draw(splash.screen)
	}

	if splash.transition != nil {
		// draw transition image
		splash.transition.Draw(splash.screen)
	}

	if splash.img != nil {
		// draw splash image
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM = splash.geoM
		splash.screen.DrawImage(splash.img, op)
	}

	if splash.shader != nil {
		// draw shader effect with splash scene to screen
		splash.shader.Draw(screen, splash.screen)
	} else {
		// draw splash scene to screen
		screen.DrawImage(splash.screen, nil)
	}
}
