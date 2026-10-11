package colors

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
)

const SHADER_GRADIENT = "shaders/gradient.kage"

func NewGradient(width, height int, startColor, endColor color.NRGBA) (*ebiten.Image, error) {
	shader, err := resources.NewShaderFromFile(SHADER_GRADIENT)
	if err != nil {
		return nil, fmt.Errorf("error loading gradient shader: %w", err)
	}

	img := ebiten.NewImage(width, height)
	uniforms := map[string]any{
		"StartColor": colorToVec4(startColor),
		"EndColor":   colorToVec4(endColor),
		"CanvasSize": []float32{float32(width), float32(height)},
	}
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = uniforms
	img.DrawRectShader(width, height, shader, op)
	return img, nil
}
