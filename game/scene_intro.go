package game

import (
	"image/color"
	"path"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	renderFx "github.com/pixelmek-3d/pixelmek-3d/game/render/effects"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/transitions"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	log "github.com/sirupsen/logrus"
)

type IntroScene struct {
	Game          *Game
	textFace      *text.GoTextFaceSource
	splash        *SplashScreen
	animation     []*ebiten.Image
	animationRate int
	numFrames     int
	animIndex     int
	animCounter   int
	loopCounter   int
	bufferScreen  *ebiten.Image
}

func NewIntroScene(g *Game) Scene {
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

				iScale := sH / iH
				iX, iY := (sW-iW*iScale)/2, (sH-iH*iScale)/2

				geoM = &ebiten.GeoM{}
				geoM.Scale(iScale, iScale)
				geoM.Translate(iX, iY)
			}
		}
	}

	// load font
	fontFile, err := resources.FileAt("fonts/broken-machine.ttf")
	if err != nil {
		panic(err)
	}
	textFace, err := text.NewGoTextFaceSource(fontFile)
	if err != nil {
		panic(err)
	}

	splash := NewSplashScreen(g)
	splash.shader = renderFx.NewCRT()
	splash.transitionOpts = &transitions.TransitionOptions{
		InDuration:   SPLASH_TIMEOUT * 2 / 5,
		HoldDuration: -1,
		OutDuration:  SPLASH_TIMEOUT * 1.5 / 5,
	}
	splash.transition = transitions.NewPixelize(splash.screen, splash.transitionOpts, ebiten.GeoM{})

	splash.geoM = *geoM

	return &IntroScene{
		Game:          g,
		textFace:      textFace,
		bufferScreen:  ebiten.NewImage(g.screenWidth, g.screenHeight),
		animation:     images,
		animationRate: 5, // TODO: define intro animation rate in a file that can be modded
		numFrames:     len(images),
		splash:        splash,
	}
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

	splash := s.splash
	if splash.shader != nil {
		splash.shader.Update()
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

	skip := splash.transition.Completed() || keyPressed || buttonPressed || inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	if skip && splash.transitionOpts.HoldDuration < 0 {
		// first key press starts transition out
		splash.transitionOpts.HoldDuration = 0
	} else if skip && splash.transitionOpts.HoldDuration >= 0 {
		// second key press skips transition to go straight to menu
		s.Game.scene = NewMenuScene(s.Game)
	}

	return nil
}

func (s *IntroScene) Draw(screen *ebiten.Image) {
	// draw current animation frame to screen
	w, h := float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy())
	splash := s.splash
	splash.screen.Clear()
	s.bufferScreen.Clear()

	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest, GeoM: splash.geoM}
	splash.screen.DrawImage(s.animation[s.animIndex], op)

	// draw PixelMek 3D title at top center
	title := "PIXELMEK 3D"
	titleFace := &text.GoTextFace{
		Source: s.textFace,
		Size:   68,
	}
	tW, _ := text.Measure(title, titleFace, titleFace.Size*1.2)
	tScale := h / 500

	textOp := &text.DrawOptions{}
	textOp.Filter = ebiten.FilterNearest
	textOp.GeoM.Scale(tScale, tScale)
	textOp.GeoM.Translate((w-(tW*tScale))/2, 0)
	textOp.ColorScale.ScaleWithColor(color.Black)
	text.Draw(splash.screen, title, titleFace, textOp)

	// draw press any key at bottom center
	pressText := "<press any key>"
	pressFace := &text.GoTextFace{
		Source: s.textFace,
		Size:   32,
	}
	pW, pH := text.Measure(pressText, pressFace, pressFace.Size*1.2)
	pScale := h / 500

	textOp = &text.DrawOptions{}
	textOp.Filter = ebiten.FilterNearest
	textOp.GeoM.Scale(pScale, pScale)
	textOp.GeoM.Translate((w-(pW*pScale))/2, h-(pH*pScale))
	textOp.ColorScale.ScaleWithColor(color.Black)
	text.Draw(splash.screen, pressText, pressFace, textOp)

	if splash.shader != nil {
		// draw shader effect with splash screen to buffer
		splash.shader.Draw(s.bufferScreen, splash.screen)
	} else {
		// draw splash screen to buffer
		s.bufferScreen.DrawImage(splash.screen, nil)
	}

	if splash.transition != nil {
		// draw transition from buffer
		splash.transition.SetImage(s.bufferScreen)
		splash.transition.Draw(screen)
	} else {
		// draw buffer directly to screen
		screen.DrawImage(s.bufferScreen, nil)
	}
}
