package colors

import (
	"image/color"
	"math"

	"github.com/erparts/go-shapes"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	SHADER_GRADIENT = "shaders/gradient.kage"
)

var shapesRenderer = shapes.NewRenderer()

func NewGradient(width, height int, startColor, endColor color.NRGBA) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	startRGBA := color.RGBAModel.Convert(startColor).(color.RGBA)
	endRGBA := color.RGBAModel.Convert(endColor).(color.RGBA)
	shapesRenderer.SimpleGradient(img, startRGBA, endRGBA, math.Pi/2)
	return img
}

func NewRadialGradient(width, height, centerX, centerY int, innerColor, outerColor color.NRGBA, radius int) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	innerRGBA := color.RGBAModel.Convert(innerColor).(color.RGBA)
	outerRGBA := color.RGBAModel.Convert(outerColor).(color.RGBA)
	shapesRenderer.GradientRadial(img, float32(centerX), float32(centerY), innerRGBA, outerRGBA, 0.0, float32(radius), float32(radius*2), -1, 1.0)
	return img
}
