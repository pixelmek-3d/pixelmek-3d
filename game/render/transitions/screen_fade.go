package transitions

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	log "github.com/sirupsen/logrus"
)

const SHADER_SCREEN_FADE = "shaders/screen_fade.kage"

type ScreenFade struct {
	screenImage *ebiten.Image
	shader      *ebiten.Shader
	geoM        ebiten.GeoM
	tOptions    *TransitionOptions
	time        float32
	tickDelta   float32
}

func NewScreenFade(img *ebiten.Image, tOptions *TransitionOptions, geoM ebiten.GeoM) *ScreenFade {
	shader, err := resources.NewShaderFromFile(SHADER_SCREEN_FADE)
	if err != nil {
		log.Errorf("error loading shader: %s", SHADER_SCREEN_FADE)
		log.Fatal(err)
	}

	d := &ScreenFade{
		shader:    shader,
		geoM:      geoM,
		tOptions:  tOptions,
		tickDelta: 1 / float32(ebiten.TPS()),
	}
	d.SetImage(img)

	return d
}

func (d *ScreenFade) SetImage(img *ebiten.Image) {
	d.screenImage = img
}

func (d *ScreenFade) Update() error {
	duration := d.tOptions.Duration()
	if d.time+d.tickDelta < duration {
		d.time += d.tickDelta
	} else {
		// move to next transition direction and reset timer
		d.tOptions.CurrentDirection += 1
		d.time = 0
	}

	return nil
}

func (d *ScreenFade) Draw(screen *ebiten.Image) {
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

	w, h := d.screenImage.Bounds().Dx(), d.screenImage.Bounds().Dy()
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"Direction":  direction,
		"Duration":   duration,
		"Time":       time,
		"ScreenSize": []float32{float32(w), float32(h)},
	}
	op.Images[0] = d.screenImage
	op.GeoM = d.geoM
	screen.DrawRectShader(w, h, d.shader, op)
}

func (d *ScreenFade) SetGeoM(geoM ebiten.GeoM) {
	d.geoM = geoM
}
