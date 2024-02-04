package effects

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
)

type CRT struct {
	shader *ebiten.Shader
}

func NewCRT() *CRT {
	shader, _ := resources.NewShaderFromFile("shaders/crt.kage")
	c := &CRT{
		shader: shader,
	}
	return c
}

func (c *CRT) Draw(screen *ebiten.Image, img *ebiten.Image, geoM ebiten.GeoM) {
	// TODO: add options to tweak CRT style, so ejection pod looks different
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = img
	op.GeoM = geoM
	screen.DrawRectShader(w, h, c.shader, op)
}
