package transitions

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/pixelmek-3d/game/resources"
)

type Dissolve struct {
	dissolveImage *ebiten.Image
	noiseImage    *ebiten.Image
	shader        *ebiten.Shader
	geoM          ebiten.GeoM
	direction     int
	duration      float32
	time          float32
	tickDelta     float32
}

func NewDissolve(img *ebiten.Image, duration float64, geoM ebiten.GeoM) *Dissolve {
	shader, _ := resources.NewShaderFromFile("shaders/dissolve.kage")
	noise, _, _ := resources.NewImageFromFile("shaders/noise.png")

	// scale noise image to match dissolve image size
	dW, dH := img.Bounds().Dx(), img.Bounds().Dy()
	nW, nH := noise.Bounds().Dx(), noise.Bounds().Dy()
	if dW != nW || dH != nH {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(float64(dW)/float64(nW), float64(dH)/float64(nH))
		scaledImage := ebiten.NewImage(dW, dH)
		scaledImage.DrawImage(noise, op)
		noise = scaledImage
	}

	// TODO: direction 1|0 needs to be a parameter for transition direction in|out

	d := &Dissolve{
		dissolveImage: img,
		noiseImage:    noise,
		shader:        shader,
		geoM:          geoM,
		direction:     1,
		duration:      float32(duration),
		tickDelta:     1 / float32(ebiten.TPS()),
	}

	return d
}

func (d *Dissolve) Update() error {
	if d.time+d.tickDelta < d.duration {
		d.time += d.tickDelta
	}

	return nil
}

func (d *Dissolve) Draw(screen *ebiten.Image) {
	w, h := d.dissolveImage.Bounds().Dx(), d.dissolveImage.Bounds().Dy()
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"Direction":  float32(d.direction),
		"Duration":   d.duration,
		"Time":       d.time,
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
