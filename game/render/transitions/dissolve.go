package transitions

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
)

type Dissolve struct {
	dissolveImage *ebiten.Image
	noiseImage    *ebiten.Image
	shader        *ebiten.Shader
	geoM          ebiten.GeoM
	tOptions      *TransitionOptions
	time          float32
	tickDelta     float32
}

func NewDissolve(img *ebiten.Image, tOptions *TransitionOptions, geoM ebiten.GeoM) *Dissolve {
	shader, _ := resources.NewShaderFromFile("shaders/dissolve.kage")
	noise, _, _ := resources.NewImageFromFile("shaders/noise.png")

	d := &Dissolve{
		noiseImage: noise,
		shader:     shader,
		geoM:       geoM,
		tOptions:   tOptions,
		tickDelta:  1 / float32(ebiten.TPS()),
	}
	d.SetImage(img)

	return d
}

func (d *Dissolve) SetImage(img *ebiten.Image) {
	d.dissolveImage = img

	// scale noise image to match dissolve image size
	dW, dH := img.Bounds().Dx(), img.Bounds().Dy()
	nW, nH := d.noiseImage.Bounds().Dx(), d.noiseImage.Bounds().Dy()
	if dW != nW || dH != nH {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(float64(dW)/float64(nW), float64(dH)/float64(nH))
		scaledImage := ebiten.NewImage(dW, dH)
		scaledImage.DrawImage(d.noiseImage, op)
		d.noiseImage = scaledImage
	}
}

func (d *Dissolve) Update() error {
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

func (d *Dissolve) Draw(screen *ebiten.Image) {
	time := d.time
	duration := d.tOptions.Duration()
	direction := 1.0
	switch d.tOptions.CurrentDirection {
	case TransitionOut:
		direction = 0.0
	case TransitionHold:
		direction = 0.0
		time = 0
	}

	w, h := d.dissolveImage.Bounds().Dx(), d.dissolveImage.Bounds().Dy()
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"Direction":  direction,
		"Duration":   duration,
		"Time":       time,
		"ScreenSize": []float32{float32(w), float32(h)},
	}
	op.Images[0] = d.dissolveImage
	op.Images[1] = d.noiseImage
	op.GeoM = d.geoM
	screen.DrawRectShader(w, h, d.shader, op)
}

func (d *Dissolve) SetGeoM(geoM ebiten.GeoM) {
	d.geoM = geoM
}
