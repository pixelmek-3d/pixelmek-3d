package transitions

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joelschutz/stagehand"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	log "github.com/sirupsen/logrus"
)

const SHADER_PIXELIZE = "shaders/pixelize.kage"

type PixelizeTransition[T any] struct {
	stagehand.BaseTransition[T]
	fromScene  stagehand.Scene[T]
	toScene    stagehand.Scene[T]
	shader     *ebiten.Shader
	noiseImage *ebiten.Image
	duration   float32
	tickDelta  float32
	time       float32
	_noise     *ebiten.Image
}

func NewPixelizeTransition[T any](duration time.Duration) *PixelizeTransition[T] {
	shader, err := resources.NewShaderFromFile(SHADER_PIXELIZE)
	if err != nil {
		log.Errorf("error loading shader: %s", SHADER_PIXELIZE)
		log.Fatal(err)
	}

	noise, _, _ := resources.NewImageFromFile("shaders/gray_noise_small.png")

	t := &PixelizeTransition[T]{
		shader:     shader,
		noiseImage: noise,
		duration:   float32(duration.Seconds()),
		tickDelta:  1 / float32(ebiten.TPS()),
	}

	return t
}

// Start starts the transition from the given "from" scene to the given "to" scene
func (t *PixelizeTransition[T]) Start(fromScene stagehand.Scene[T], toScene stagehand.Scene[T], sm stagehand.SceneController[T]) {
	t.BaseTransition.Start(fromScene, toScene, sm)
	t.time = 0

	// these have to be redefined and set since they are private fields in stagehand.BaseTransition
	t.fromScene = fromScene
	t.toScene = toScene
}

func (t *PixelizeTransition[T]) _updateNoiseImage(sceneImage *ebiten.Image) {
	update := false

	dW, dH := sceneImage.Bounds().Dx(), sceneImage.Bounds().Dy()
	if t._noise == nil {
		update = true
	} else {
		nW, nH := t._noise.Bounds().Dx(), t._noise.Bounds().Dy()
		if dW != nW || dH != nH {
			update = true
		}
	}

	if update {
		nW, nH := t.noiseImage.Bounds().Dx(), t.noiseImage.Bounds().Dy()
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(float64(dW)/float64(nW), float64(dH)/float64(nH))
		scaledImage := ebiten.NewImage(dW, dH)
		scaledImage.DrawImage(t.noiseImage, op)
		t._noise = scaledImage
	}
}

func (t *PixelizeTransition[T]) Update() error {
	if t.time+t.tickDelta < t.duration {
		t.time += t.tickDelta
	} else {
		t.End()
	}

	// Update the scenes
	return t.BaseTransition.Update()
}

func (t *PixelizeTransition[T]) Draw(screen *ebiten.Image) {
	toImg, fromImg := stagehand.PreDraw(screen.Bounds(), t.fromScene, t.toScene)
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()

	t._updateNoiseImage(toImg)

	// TODO: implement direction up or down transition
	//direction := 1.0

	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		//"Direction": direction,
		"Duration": t.duration,
		"Time":     t.time,
	}
	op.Images[0] = fromImg
	op.Images[1] = toImg
	op.Images[2] = t._noise

	// draw shader from buffer image
	screen.DrawRectShader(w, h, t.shader, op)
}
