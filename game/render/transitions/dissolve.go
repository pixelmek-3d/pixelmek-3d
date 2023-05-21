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
	time          int
	tps           float32
}

func NewDissolve(img *ebiten.Image, geoM ebiten.GeoM) *Dissolve {
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

	d := &Dissolve{
		dissolveImage: img,
		noiseImage:    noise,
		shader:        shader,
		geoM:          geoM,
		tps:           float32(ebiten.TPS()),
	}

	return d
}

func (d *Dissolve) Update() error {
	d.time++
	return nil
}

func (d *Dissolve) Draw(screen *ebiten.Image) {
	w, h := d.dissolveImage.Bounds().Dx(), d.dissolveImage.Bounds().Dy()
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"Time":       float32(d.time) / d.tps,
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
