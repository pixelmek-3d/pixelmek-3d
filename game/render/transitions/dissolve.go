package transitions

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	log "github.com/sirupsen/logrus"
)

const SHADER_DISSOLVE = "shaders/dissolve.kage"

type Dissolve struct {
	dissolveImage *ebiten.Image
	noiseImage    *ebiten.Image
	shader        *ebiten.Shader
	geoM          ebiten.GeoM
	tOptions      *TransitionOptions
	time          float32
	tickDelta     float32
	completed     bool
}

func NewDissolve(img *ebiten.Image, tOptions *TransitionOptions, geoM ebiten.GeoM) *Dissolve {
	shader, err := resources.NewShaderFromFile(SHADER_DISSOLVE)
	if err != nil {
		log.Errorf("error loading shader: %s", SHADER_DISSOLVE)
		log.Fatal(err)
	}

	noise, _, _ := resources.NewImageFromFile("shaders/noise.png")

	t := &Dissolve{
		noiseImage: noise,
		shader:     shader,
		geoM:       geoM,
		tOptions:   tOptions,
		tickDelta:  1 / float32(ebiten.TPS()),
	}
	t.SetImage(img)

	return t
}

func (t *Dissolve) Completed() bool {
	return t.completed
}

func (t *Dissolve) SetImage(img *ebiten.Image) {
	t.dissolveImage = img

	// scale noise image to match dissolve image size
	dW, dH := img.Bounds().Dx(), img.Bounds().Dy()
	nW, nH := t.noiseImage.Bounds().Dx(), t.noiseImage.Bounds().Dy()
	if dW != nW || dH != nH {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(float64(dW)/float64(nW), float64(dH)/float64(nH))
		scaledImage := ebiten.NewImage(dW, dH)
		scaledImage.DrawImage(t.noiseImage, op)
		t.noiseImage = scaledImage
	}
}

func (t *Dissolve) Update() error {
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

func (t *Dissolve) Draw(screen *ebiten.Image) {
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

	w, h := t.dissolveImage.Bounds().Dx(), t.dissolveImage.Bounds().Dy()
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"Direction": direction,
		"Duration":  duration,
		"Time":      time,
	}
	op.Images[0] = t.dissolveImage
	op.Images[1] = t.noiseImage
	op.GeoM = t.geoM
	screen.DrawRectShader(w, h, t.shader, op)
}

func (t *Dissolve) SetGeoM(geoM ebiten.GeoM) {
	t.geoM = geoM
}
