package transitions

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joelschutz/stagehand"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	log "github.com/sirupsen/logrus"
)

const SHADER_DISSOLVE = "shaders/dissolve.kage"

type DissolveTransition[T any] struct {
	stagehand.BaseTransition[T]
	fromScene  stagehand.Scene[T]
	toScene    stagehand.Scene[T]
	shader     *ebiten.Shader
	noiseImage *ebiten.Image
	duration   float32
	tickDelta  float32
	time       float32
	direction  float32
	_dissolve  *ebiten.Image
	_noise     *ebiten.Image
}

func NewDissolveTransition[T any](duration time.Duration) *DissolveTransition[T] {
	shader, err := resources.NewShaderFromFile(SHADER_DISSOLVE)
	if err != nil {
		log.Errorf("error loading shader: %s", SHADER_DISSOLVE)
		log.Fatal(err)
	}
	noise, _, _ := resources.NewImageFromFile("shaders/noise.png")
	return &DissolveTransition[T]{
		shader:     shader,
		noiseImage: noise,
		duration:   float32(duration.Seconds() / 2),
		tickDelta:  1 / float32(ebiten.TPS()),
	}
}

// Start starts the transition from the given "from" scene to the given "to" scene
func (t *DissolveTransition[T]) Start(fromScene stagehand.Scene[T], toScene stagehand.Scene[T], sm stagehand.SceneController[T]) {
	t.BaseTransition.Start(fromScene, toScene, sm)
	t.time = 0
	t.direction = -1

	// these have to be redefined and set since they are private fields in stagehand.BaseTransition
	t.fromScene = fromScene
	t.toScene = toScene
}

func (t *DissolveTransition[T]) _setDissolveImage(img *ebiten.Image) {
	t._dissolve = img

	// scale noise image to match dissolve image size
	if t._noise == nil {
		t.updateNoiseImage()
	} else {
		dW, dH := img.Bounds().Dx(), img.Bounds().Dy()
		nW, nH := t._noise.Bounds().Dx(), t._noise.Bounds().Dy()
		if dW != nW || dH != nH {
			t.updateNoiseImage()
		}
	}
}

func (t *DissolveTransition[T]) updateNoiseImage() {
	dW, dH := t._dissolve.Bounds().Dx(), t._dissolve.Bounds().Dy()
	nW, nH := t.noiseImage.Bounds().Dx(), t.noiseImage.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(float64(dW)/float64(nW), float64(dH)/float64(nH))
	scaledImage := ebiten.NewImage(dW, dH)
	scaledImage.DrawImage(t.noiseImage, op)
	t._noise = scaledImage
}

// Update updates the transition state
func (t *DissolveTransition[T]) Update() error {
	switch {
	case t.time+t.tickDelta < t.duration:
		t.time += t.tickDelta
	case t.direction < 0:
		t.direction = 1
		t.time = 0
	default:
		t.End()
	}

	// Update the scenes
	return t.BaseTransition.Update()
}

// Draw draws the transition effect
func (t *DissolveTransition[T]) Draw(screen *ebiten.Image) {
	toImg, fromImg := stagehand.PreDraw(screen.Bounds(), t.fromScene, t.toScene)
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()

	if t.direction < 0 {
		t._setDissolveImage(fromImg)
	} else {
		t._setDissolveImage(toImg)
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"Direction": t.direction,
		"Duration":  t.duration,
		"Time":      t.time,
	}
	op.Images[0] = t._dissolve
	op.Images[1] = t._noise
	screen.DrawRectShader(w, h, t.shader, op)
}
