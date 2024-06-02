package transitions

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	log "github.com/sirupsen/logrus"
)

const SHADER_PIXELIZE = "shaders/pixelize.kage"

type Pixelize struct {
	pixelizeImage *ebiten.Image
	shader        *ebiten.Shader
	noiseImage    *ebiten.Image
	geoM          ebiten.GeoM
	tOptions      *TransitionOptions
	time          float32
	tickDelta     float32
	completed     bool
	_blank        *ebiten.Image
	_noise        *ebiten.Image
}

func NewPixelize(img *ebiten.Image, tOptions *TransitionOptions, geoM ebiten.GeoM) *Pixelize {
	shader, err := resources.NewShaderFromFile(SHADER_PIXELIZE)
	if err != nil {
		log.Errorf("error loading shader: %s", SHADER_PIXELIZE)
		log.Fatal(err)
	}

	noise, _, _ := resources.NewImageFromFile("shaders/gray_noise_small.png")

	t := &Pixelize{
		shader:     shader,
		noiseImage: noise,
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

	// scale noise image to match dissolve image size
	if t._noise == nil {
		t.updateInternalImages()
	} else {
		dW, dH := img.Bounds().Dx(), img.Bounds().Dy()
		nW, nH := t._noise.Bounds().Dx(), t._noise.Bounds().Dy()
		if dW != nW || dH != nH {
			t.updateInternalImages()
		}
	}
}

func (t *Pixelize) updateInternalImages() {
	dW, dH := t.pixelizeImage.Bounds().Dx(), t.pixelizeImage.Bounds().Dy()
	nW, nH := t.noiseImage.Bounds().Dx(), t.noiseImage.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(float64(dW)/float64(nW), float64(dH)/float64(nH))
	scaledImage := ebiten.NewImage(dW, dH)
	scaledImage.DrawImage(t.noiseImage, op)
	t._noise = scaledImage
	t._blank = ebiten.NewImage(dW, dH)
}

func (t *Pixelize) Update() error {
	if t.completed {
		return nil
	}

	duration := t.tOptions.Duration()
	if t.time+t.tickDelta < duration {
		t.time += t.tickDelta
	} else {
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
	direction := 1.0
	switch t.tOptions.CurrentDirection {
	case TransitionOut:
		direction = -1.0
	case TransitionHold:
		direction = 0.0
		time = 0
	}

	// draw image to buffer with translation first (since this shader does not like any post-GeoM)
	w, h := t.pixelizeImage.Bounds().Dx(), t.pixelizeImage.Bounds().Dy()
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"Direction": direction,
		"Duration":  duration,
		"Time":      time,
	}
	op.Images[0] = t.pixelizeImage
	op.Images[1] = t._blank
	op.Images[2] = t._noise
	op.GeoM = t.geoM

	// draw shader from buffer image
	screen.DrawRectShader(w, h, t.shader, op)
}

func (t *Pixelize) SetGeoM(geoM ebiten.GeoM) {
	t.geoM = geoM
}
