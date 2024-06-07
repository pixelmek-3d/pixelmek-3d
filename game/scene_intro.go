package game

import (
	"image/color"
	"path"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/joelschutz/stagehand"
	renderFx "github.com/pixelmek-3d/pixelmek-3d/game/render/effects"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	log "github.com/sirupsen/logrus"
)

type IntroScene struct {
	BaseScene
	textFace      *text.GoTextFaceSource
	shader        SceneShader
	bufferScreen  *ebiten.Image
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

	return &IntroScene{
		BaseScene: BaseScene{
			game: g,
		},
		textFace:      textFace,
		shader:        renderFx.NewCRT(),
		bufferScreen:  ebiten.NewImage(g.screenWidth, g.screenHeight),
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

	if s.shader != nil {
		s.shader.Update()
	}

	if s.state.onTransition {
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
		s.sm.ProcessTrigger(PostIntroTrigger)
	}

	return nil
}

func (s *IntroScene) Draw(screen *ebiten.Image) {
	// draw current animation frame to screen
	w, h := float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy())
	s.bufferScreen.Clear()

	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest, GeoM: s.geoM}
	s.bufferScreen.DrawImage(s.animation[s.animIndex], op)

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
	text.Draw(s.bufferScreen, title, titleFace, textOp)

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
	text.Draw(s.bufferScreen, pressText, pressFace, textOp)

	if s.shader != nil {
		// draw shader effect with buffer to screen
		s.shader.Draw(screen, s.bufferScreen)
	} else {
		// draw buffer directly to screen
		screen.DrawImage(s.bufferScreen, nil)
	}
}
