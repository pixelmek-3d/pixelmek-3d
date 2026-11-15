package effects

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
)

const SHADER_MASK = "shaders/mask.kage"

// MaskImage draws base image with a mask overlay and returns as new image
func MaskImage(baseImg, overlayTex *ebiten.Image, targetColor [3]float32, threshhold, tintStrength float32) *ebiten.Image {
	width, height := baseImg.Bounds().Dx(), baseImg.Bounds().Dy()

	texWidth, texHeight := overlayTex.Bounds().Dx(), overlayTex.Bounds().Dy()
	if texWidth != width || texHeight != height {
		// mask texture must be same size as base image
		scaledTex := ebiten.NewImage(width, height)
		scaleX := float64(width) / float64(texWidth)
		scaleY := float64(height) / float64(texHeight)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scaleX, scaleY)
		op.Filter = ebiten.FilterNearest
		scaledTex.DrawImage(overlayTex, op)
		overlayTex = scaledTex
	}

	shader, err := resources.NewShaderFromFile(SHADER_MASK)
	if err != nil {
		log.Fatalf("error loading mask shader: %s", err)
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = baseImg
	op.Images[1] = overlayTex
	op.Uniforms = map[string]any{
		"TargetColor":  targetColor,
		"Threshold":    threshhold,
		"TintStrength": tintStrength,
	}

	img := ebiten.NewImage(width, height)
	img.DrawRectShader(width, height, shader, op)
	return img
}
