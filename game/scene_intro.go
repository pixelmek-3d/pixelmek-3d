package game

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/joelschutz/stagehand"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	log "github.com/sirupsen/logrus"
)

type IntroScene struct {
	BaseScene
	animation     []*ebiten.Image
	geoM          ebiten.GeoM
	animationRate int
	numFrames     int
	animIndex     int
	animCounter   int
	loopCounter   int
}

func NewIntroScene(g *Game) *IntroScene {
	// PixelMek 3D intro animation

	// load all intro image frames
	introPath := "textures/intro"
	introFiles, err := resources.ReadDir(introPath)
	if err != nil {
		panic(err)
	}

	var geoM *ebiten.GeoM

	// import files into image array
	images := make([]*ebiten.Image, 0, 10)
	for _, f := range introFiles {
		if f.IsDir() {
			continue
		}
		fName := f.Name()
		if strings.HasPrefix(fName, "intro") && filepath.Ext(fName) == ".png" {
			// load image and scale to fit screen
			fPath := path.Join(introPath, fName)
			img, _, err := resources.NewImageFromFile(fPath)
			if err != nil {
				log.Error(err)
				continue
			}

			images = append(images, img)

			if geoM == nil {
				sW, sH := float64(g.screenWidth), float64(g.screenHeight)
				iW, iH := float64(img.Bounds().Dx()), float64(img.Bounds().Dy())

				iScale := sW / iW
				iX, iY := 0.0, (sH-iH*iScale)/2

				geoM = &ebiten.GeoM{}
				geoM.Scale(iScale, iScale)
				geoM.Translate(iX, iY)
			}
		}
	}

	return &IntroScene{
		BaseScene: BaseScene{
			game: g,
		},
		animation:     images,
		animationRate: 5, // TODO: define intro animation rate in a file that can be modded
		numFrames:     len(images),
		geoM:          *geoM,
	}
}

func (s *IntroScene) Load(st SceneState, sm stagehand.SceneController[SceneState]) {
	s.BaseScene.Load(st, sm)
	s.animCounter = 0
	s.loopCounter = 0
}

func (s *IntroScene) Update() error {
	// determine when to move to next animation frame
	if s.animationRate > 0 {
		if s.animCounter >= s.animationRate {
			s.animCounter = 0
			s.animIndex++
			if s.animIndex >= s.numFrames {
				s.animIndex = 0
				s.loopCounter++
			}
		} else {
			s.animCounter++
		}
	}

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
	if skip {
		s.sm.ProcessTrigger(SplashTrigger)
	}

	return nil
}

func (s *IntroScene) Draw(screen *ebiten.Image) {
	// draw current animation frame to screen
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest, GeoM: s.geoM}
	screen.DrawImage(s.animation[s.animIndex], op)

	// TODO: draw PixelMek 3D intro text on top
}
