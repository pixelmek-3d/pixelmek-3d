package transitions

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joelschutz/stagehand"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	log "github.com/sirupsen/logrus"
)

const SHADER_FADE = "shaders/fade.kage"

type FadeTransition[T any] struct {
	stagehand.BaseTransition[T]
	fromScene stagehand.Scene[T]
	toScene   stagehand.Scene[T]
	shader    *ebiten.Shader
	duration  float32
	tickDelta float32
	time      float32
	direction float32
}

func NewFadeTransition[T any](duration time.Duration) *FadeTransition[T] {
	shader, err := resources.NewShaderFromFile(SHADER_FADE)
	if err != nil {
		log.Errorf("error loading shader: %s", SHADER_FADE)
		log.Fatal(err)
	}

	return &FadeTransition[T]{
		shader:    shader,
		duration:  float32(duration.Seconds() / 2),
		tickDelta: 1 / float32(ebiten.TPS()),
	}
}

// Start starts the transition from the given "from" scene to the given "to" scene
func (t *FadeTransition[T]) Start(fromScene stagehand.Scene[T], toScene stagehand.Scene[T], sm stagehand.SceneController[T]) {
	t.BaseTransition.Start(fromScene, toScene, sm)
	t.time = 0
	t.direction = -1

	// these have to be redefined and set since they are private fields in stagehand.BaseTransition
	t.fromScene = fromScene
	t.toScene = toScene
}

func (t *FadeTransition[T]) Update() error {
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

func (t *FadeTransition[T]) Draw(screen *ebiten.Image) {
	toImg, fromImg := stagehand.PreDraw(screen.Bounds(), t.fromScene, t.toScene)
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()

	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"Direction": t.direction,
		"Duration":  t.duration,
		"Time":      t.time,
	}

	if t.direction < 0 {
		op.Images[0] = fromImg
	} else {
		op.Images[0] = toImg
	}

	screen.DrawRectShader(w, h, t.shader, op)
}
