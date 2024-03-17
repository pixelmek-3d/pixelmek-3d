package transitions

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	log "github.com/sirupsen/logrus"
)

const SHADER_FADE = "shaders/fade.kage"

type Fade struct {
	fadeImage *ebiten.Image
	shader    *ebiten.Shader
	geoM      ebiten.GeoM
	tOptions  *TransitionOptions
	time      float32
	tickDelta float32
	completed bool
}

func NewFade(img *ebiten.Image, tOptions *TransitionOptions, geoM ebiten.GeoM) *Fade {
	shader, err := resources.NewShaderFromFile(SHADER_FADE)
	if err != nil {
		log.Errorf("error loading shader: %s", SHADER_FADE)
		log.Fatal(err)
	}

	t := &Fade{
		shader:    shader,
		geoM:      geoM,
		tOptions:  tOptions,
		tickDelta: 1 / float32(ebiten.TPS()),
	}
	t.SetImage(img)

	return t
}

func (t *Fade) Completed() bool {
	return t.completed
}

func (t *Fade) SetImage(img *ebiten.Image) {
	t.fadeImage = img
}

func (t *Fade) Update() error {
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

func (t *Fade) Draw(screen *ebiten.Image) {
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

	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"Direction": direction,
		"Duration":  duration,
		"Time":      time,
	}
	op.Images[0] = t.fadeImage
	op.GeoM = t.geoM
	screen.DrawRectShader(w, h, t.shader, op)
}

func (t *Fade) SetGeoM(geoM ebiten.GeoM) {
	t.geoM = geoM
}
