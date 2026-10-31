package colors

import (
	"image/color"
	"math"

	"github.com/erparts/go-shapes"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/effects"
)

const SHADER_BLUR = "shaders/blur.kage"

var shapesRenderer = shapes.NewRenderer()

func NewVerticalGradient(width, height int, topColor, bottomColor color.NRGBA) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	topRGBA := color.RGBAModel.Convert(topColor).(color.RGBA)
	bottomRGBA := color.RGBAModel.Convert(bottomColor).(color.RGBA)
	shapesRenderer.SimpleGradient(img, topRGBA, bottomRGBA, math.Pi/2)

	// apply blur shader to smooth gradient lines
	return effects.BlurImage(img)
}

func NewRadialGradient(width, height, centerX, centerY int, innerColor, outerColor color.NRGBA, radius int) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	innerRGBA := color.RGBAModel.Convert(innerColor).(color.RGBA)
	outerRGBA := color.RGBAModel.Convert(outerColor).(color.RGBA)
	shapesRenderer.GradientRadial(img, float32(centerX), float32(centerY), innerRGBA, outerRGBA, 0.0, float32(radius), float32(radius*2), -1, 1.0)

	// apply blur shader to smooth gradient lines
	return effects.BlurImage(img)
}
