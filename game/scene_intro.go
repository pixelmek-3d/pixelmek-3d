package game

import (
	"image"

	"github.com/harbdog/pixelmek-3d/game/render/effects"
	"github.com/harbdog/pixelmek-3d/game/render/transitions"
	"github.com/harbdog/pixelmek-3d/game/resources"

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
	*ebiten.Image
	geoM       ebiten.GeoM
	effect     SceneEffect
	transition SceneTransition
}

func NewIntroScene(g *Game) *IntroScene {
	// load intro splash images
	var splashes = make([]*SplashScreen, 0)
	splashRect := g.uiRect()

	im, _, err := resources.NewImageFromFile("textures/ebitengine_splash.png")
	if err == nil {
		geoM := splashGeoM(im, splashRect)
		tOpts := &transitions.TransitionOptions{
			InDuration:   2.0,
			HoldDuration: 1.5,
			OutDuration:  1.0,
		}
		splash := &SplashScreen{
			Image:      nil,
			transition: transitions.NewDissolve(im, tOpts, geoM),
			geoM:       geoM,
		}
		splashes = append(splashes, splash)
	}

	im, _, err = resources.NewImageFromFile("textures/gopher_space.png")
	if err == nil {
		geoM := splashGeoM(im, splashRect)
		splash := &SplashScreen{
			Image:  im,
			effect: effects.NewStars(g.screenWidth, g.screenHeight),
			geoM:   geoM,
		}
		splashes = append(splashes, splash)
	}

	return &IntroScene{
		Game:        g,
		splashes:    splashes,
		splashRect:  splashRect,
		splashTimer: SPLASH_TIMEOUT,
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
	skip := len(keys) > 0 || inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
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
		// TODO: make last splash require key/button press to move on to main menu?
		s.Game.scene = NewMenuScene(s.Game)
	}

	return nil
}

func (s *IntroScene) Draw(screen *ebiten.Image) {
	// draw effect as splash image background
	splash := s.currentSplash()
	if splash.effect != nil {
		splash.effect.Draw(screen)
	}

	if splash.transition != nil {
		splash.transition.Draw(screen)
	}

	if splash.Image != nil {
		// draw splash image
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM = splash.geoM
		screen.DrawImage(splash.Image, op)
	}
}
