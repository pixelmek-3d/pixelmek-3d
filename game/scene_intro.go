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
		splash := &SplashScreen{
			Image:      nil,
			transition: transitions.NewDissolve(im, SPLASH_TIMEOUT/2, geoM),
			geoM:       geoM,
		}
		splashes = append(splashes, splash)
	}

	im, _, err = resources.NewImageFromFile("textures/gopher_space.png")
	if err == nil {
		geoM := splashGeoM(im, splashRect)
		splash := &SplashScreen{
			Image:  im,
			effect: effects.NewStarSpace(g.width, g.height),
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
		// TODO: only update transition at beginning/end
		splash.transition.Update()
	}

	// TODO: add transitional animation between images?

	keys := inpututil.AppendJustPressedKeys(nil)
	if len(keys) > 0 || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		s.splashIndex += 1
	}

	// use timer to move on if no input
	s.splashTimer -= 1 / float64(ebiten.TPS())
	if s.splashTimer <= 0 {
		s.splashIndex += 1
		s.splashTimer = SPLASH_TIMEOUT
	}

	if s.splashIndex >= len(s.splashes) {
		// TODO: make last splash require key/button press to move on to main menu?
		s.Game.scene = NewMissionScene(s.Game)
	}

	return nil
}

func (s *IntroScene) Draw(screen *ebiten.Image) {
	// draw effect as splash image background
	splash := s.currentSplash()
	if splash.effect != nil {
		splash.effect.Draw(screen)
	}

	if splash.Image != nil {
		// draw splash image
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM = splash.geoM
		screen.DrawImage(splash.Image, op)
	}

	if splash.transition != nil {
		// TODO: only draw transition at beginning/end
		splash.transition.Draw(screen)
	}
}
