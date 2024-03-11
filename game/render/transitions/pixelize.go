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

	d := &Pixelize{
		shader:    shader,
		geoM:      geoM,
		tOptions:  tOptions,
		tickDelta: 1 / float32(ebiten.TPS()),
	}
	d.SetImage(img)

	return d
}

func (d *Pixelize) Completed() bool {
	return d.completed
}

func (d *Pixelize) SetImage(img *ebiten.Image) {
	d.pixelizeImage = img
}

func (d *Pixelize) screenBuffer(w, h int) {
	createBuffer := false
	if d.bufferImage == nil {
		createBuffer = true
	} else {
		bW, bH := d.bufferImage.Bounds().Dx(), d.bufferImage.Bounds().Dy()
		if w != bW || h != bH {
			createBuffer = true
		}
	}

	if createBuffer {
		d.bufferImage = ebiten.NewImage(w, h)
	}
}

func (d *Pixelize) Update() error {
	if d.completed {
		return nil
	}

	duration := d.tOptions.Duration()
	if d.time+d.tickDelta < duration {
		d.time += d.tickDelta
	} else {
		// move to next transition direction and reset timer
		d.tOptions.CurrentDirection += 1
		d.time = 0
	}

	if d.tOptions.CurrentDirection == TransitionCompleted {
		d.completed = true
	}

	return nil
}

func (d *Pixelize) Draw(screen *ebiten.Image) {
	time := d.time
	duration := d.tOptions.Duration()
	direction := 1.0
	switch d.tOptions.CurrentDirection {
	case TransitionOut:
		direction = -1.0
	case TransitionHold:
		direction = 0.0
		time = 0
	}

	// draw image to buffer with translation first (since this shader does not like any post-GeoM)
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	d.screenBuffer(w, h)
	d.bufferImage.DrawImage(d.pixelizeImage, &ebiten.DrawImageOptions{GeoM: d.geoM})

	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"Direction": direction,
		"Duration":  duration,
		"Time":      time,
	}
	op.Images[0] = d.bufferImage

	// draw shader from buffer image
	screen.DrawRectShader(w, h, d.shader, op)
}

func (d *Pixelize) SetGeoM(geoM ebiten.GeoM) {
	d.geoM = geoM
}
