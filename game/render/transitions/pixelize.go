package transitions

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	log "github.com/sirupsen/logrus"
)

const SHADER_PIXELIZE = "shaders/pixelize.kage"

type Pixelize struct {
	pixelizeImage *ebiten.Image
	bufferImage   *ebiten.Image
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

	t := &Pixelize{
		shader:    shader,
		geoM:      geoM,
		tOptions:  tOptions,
		tickDelta: 1 / float32(ebiten.TPS()),
	}
	t.SetImage(img)

	return t
}

func (t *Pixelize) Completed() bool {
	return t.completed
}

func (t *Pixelize) SetImage(img *ebiten.Image) {
	t.pixelizeImage = img
}

func (t *Pixelize) screenBuffer(w, h int) {
	createBuffer := false
	if t.bufferImage == nil {
		createBuffer = true
	} else {
		bW, bH := t.bufferImage.Bounds().Dx(), t.bufferImage.Bounds().Dy()
		if w != bW || h != bH {
			createBuffer = true
		}
	}

	if createBuffer {
		t.bufferImage = ebiten.NewImage(w, h)
	}
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
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	t.screenBuffer(w, h)
	t.bufferImage.DrawImage(t.pixelizeImage, &ebiten.DrawImageOptions{GeoM: t.geoM})

	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"Direction": direction,
		"Duration":  duration,
		"Time":      time,
	}
	op.Images[0] = t.bufferImage

	// draw shader from buffer image
	screen.DrawRectShader(w, h, t.shader, op)
}

func (t *Pixelize) SetGeoM(geoM ebiten.GeoM) {
	t.geoM = geoM
}
