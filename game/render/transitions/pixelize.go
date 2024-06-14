package transitions

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	log "github.com/sirupsen/logrus"
)

const SHADER_PIXELIZE = "shaders/pixelize.kage"

type Pixelize struct {
	pixelizeImage *ebiten.Image
	noiseImage    *ebiten.Image
	blankImage    *ebiten.Image
	shader        *ebiten.Shader
	geoM          ebiten.GeoM
	tOptions      *TransitionOptions
	time          float32
	tickDelta     float32
	completed     bool
}

func NewPixelize(img *ebiten.Image, tOptions *TransitionOptions, geoM ebiten.GeoM) *Pixelize {
	shader, err := resources.NewShaderFromFile(SHADER_PIXELIZE)
	if err != nil {
		log.Errorf("error loading shader: %s", SHADER_PIXELIZE)
		log.Fatal(err)
	}

	noise, _, _ := resources.NewImageFromFile("shaders/gray_noise_small.png")

	t := &Pixelize{
		noiseImage: noise,
		shader:     shader,
		geoM:       geoM,
		tOptions:   tOptions,
		tickDelta:  1 / float32(ebiten.TPS()),
	}
	t.SetImage(img)

	return t
}

func (t *Pixelize) Completed() bool {
	return t.completed
}

func (t *Pixelize) SetImage(img *ebiten.Image) {
	t.pixelizeImage = img

	// scale noise image to match pixelize image size
	dW, dH := img.Bounds().Dx(), img.Bounds().Dy()
	nW, nH := t.noiseImage.Bounds().Dx(), t.noiseImage.Bounds().Dy()
	if dW != nW || dH != nH {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(float64(dW)/float64(nW), float64(dH)/float64(nH))
		scaledImage := ebiten.NewImage(dW, dH)
		scaledImage.DrawImage(t.noiseImage, op)
		t.noiseImage = scaledImage

		// also scale blank image used to transition with
		t.blankImage = ebiten.NewImage(dW, dH)
	}
}

func (t *Pixelize) Update() error {
	if t.completed {
		return nil
	}

	duration := t.tOptions.Duration()
	switch {
	case t.tOptions.CurrentDirection == TransitionHold && duration < 0:
		// keep at transition held state when duration less than zero
		t.time = 0
	case t.time+t.tickDelta < duration:
		t.time += t.tickDelta
	default:
		// move to next transition direction and reset timer
		t.tOptions.CurrentDirection += 1
		t.time = 0
	}

	if t.tOptions.CurrentDirection == TransitionCompleted {
		t.completed = true
	} else if t.time == 0 && t.tOptions.Duration() == 0 {
		// if the next transition unused, update to move on to the next
		t.Update()
	}

	return nil
}

func (t *Pixelize) Draw(screen *ebiten.Image) {
	time := t.time
	duration := t.tOptions.Duration()
	if t.tOptions.CurrentDirection == TransitionHold {
		time = 0
	}

	w, h := t.pixelizeImage.Bounds().Dx(), t.pixelizeImage.Bounds().Dy()
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"Duration": duration,
		"Time":     time,
	}

	switch t.tOptions.CurrentDirection {
	case TransitionOut:
		op.Images[0] = t.pixelizeImage
		op.Images[1] = t.blankImage
	default:
		op.Images[0] = t.blankImage
		op.Images[1] = t.pixelizeImage
	}

	op.Images[2] = t.noiseImage
	op.GeoM = t.geoM
	screen.DrawRectShader(w, h, t.shader, op)
}
