package effects

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	log "github.com/sirupsen/logrus"
)

const SHADER_CRT = "shaders/crt.kage"

type CRT struct {
	shader *ebiten.Shader
}

func NewCRT() *CRT {
	shader, err := resources.NewShaderFromFile(SHADER_CRT)
	if err != nil {
		log.Errorf("error loading shader: %s", SHADER_CRT)
		log.Fatal(err)
	}
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
