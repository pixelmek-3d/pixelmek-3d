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

func (c *CRT) Update() error {
	return nil
}

func (c *CRT) Draw(screen *ebiten.Image, img *ebiten.Image) {
	c.DrawWithOptions(screen, img, true)
}

func (c *CRT) DrawWithOptions(screen *ebiten.Image, img *ebiten.Image, curve bool) {
	showCurve := 0
	if curve {
		showCurve = 1
	}

	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"ShowCurve": showCurve,
	}
	op.Images[0] = img
	screen.DrawRectShader(w, h, c.shader, op)
}
