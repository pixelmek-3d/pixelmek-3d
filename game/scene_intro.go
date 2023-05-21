package game

import (
	"image"

	"github.com/harbdog/pixelmek-3d/game/render"
	"github.com/harbdog/pixelmek-3d/game/resources"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	SPLASH_TIMEOUT = 4.0
)

type IntroScene struct {
	Game        *Game
	starSpace   *render.StarSpace
	splashes    []*ebiten.Image
	splashRect  image.Rectangle
	splashIndex int
	splashTimer float64
}

func NewIntroScene(g *Game) *IntroScene {
	// load intro splash images
	var images = make([]*ebiten.Image, 0)

	im, _, err := resources.NewImageFromFile("textures/gopher_space.png")
	if err == nil {
		images = append(images, im)
	}

	im, _, err = resources.NewImageFromFile("textures/ebitengine_splash.png")
	if err == nil {
		images = append(images, im)
	}

	return &IntroScene{
		Game:        g,
		starSpace:   render.NewStarSpace(g.width, g.height),
		splashes:    images,
		splashRect:  g.uiRect(),
		splashTimer: SPLASH_TIMEOUT,
	}
}

func (s *IntroScene) Update() error {
	s.starSpace.Update()

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
	// draw star space as splash image background
	s.starSpace.Draw(screen)

	// draw splash image
	splash := s.splashes[s.splashIndex]
	sW, sH := float64(splash.Bounds().Dx()), float64(splash.Bounds().Dy())
	bX, bY, bW, bH := s.splashRect.Min.X, s.splashRect.Min.Y, s.splashRect.Dx(), s.splashRect.Dy()

	// scale image to only take up a portion of the space
	sScale := 0.75 * float64(bH) / sH
	sW, sH = sW*sScale, sH*sScale
	sX, sY := float64(bX)+float64(bW)/2-sW/2, float64(bY)+float64(bH)/2-sH/2

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(sScale, sScale)
	op.GeoM.Translate(sX, sY)
	screen.DrawImage(splash, op)
}
