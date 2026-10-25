package colors

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
)

const (
	SHADER_GRADIENT_VERTICAL = "shaders/gradient_vertical.kage"
	SHADER_GRADIENT_RADIAL   = "shaders/gradient_radial.kage"
)

func NewGradient(width, height int, topColor, bottomColor color.NRGBA) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	shader, err := resources.NewShaderFromFile(SHADER_GRADIENT_VERTICAL)
	if err != nil {
		log.Fatalf("error loading gradient shader: %s", err)
	}

	uniforms := map[string]any{
		"TopColor":    colorToVec4(topColor),
		"BottomColor": colorToVec4(bottomColor),
		"CanvasSize":  []float32{float32(width), float32(height)},
	}
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = uniforms
	img.DrawRectShader(width, height, shader, op)
	return img
}

func NewRadialGradient(width, height, centerX, centerY int, innerColor, outerColor color.NRGBA, radius int) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	shader, err := resources.NewShaderFromFile(SHADER_GRADIENT_RADIAL)
	if err != nil {
		log.Fatalf("error loading gradient radial shader: %s", err)
	}

	uniforms := map[string]any{
		"InnerColor": colorToVec4(innerColor),
		"OuterColor": colorToVec4(outerColor),
		"Center":     []float32{float32(centerX), float32(centerY)},
		"Radius":     float32(radius),
	}
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = uniforms
	img.DrawRectShader(width, height, shader, op)
	return img
}
