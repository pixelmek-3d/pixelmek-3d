package effects

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
)

const SHADER_BLUR = "shaders/blur.kage"

// BlurImage applies a 2-pass blur shader with final pass dither and returns as new image
func BlurImage(img0 *ebiten.Image) *ebiten.Image {
	width, height := img0.Bounds().Dx(), img0.Bounds().Dy()

	shader, err := resources.NewShaderFromFile(SHADER_BLUR)
	if err != nil {
		log.Fatalf("error loading blur shader: %s", err)
	}

	img1 := ebiten.NewImage(width, height)
	img2 := ebiten.NewImage(width, height)

	uniforms := map[string]any{
		"Direction": [2]float32{1, 0},
		"Dither":    float32(0),
	}
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = uniforms
	op.Images[0] = img0
	img1.DrawRectShader(width, height, shader, op)

	uniforms = map[string]any{
		"Direction": [2]float32{0, 1},
		"Dither":    float32(1),
	}
	op = &ebiten.DrawRectShaderOptions{}
	op.Uniforms = uniforms
	op.Images[0] = img1
	img2.DrawRectShader(width, height, shader, op)
	return img2
}
