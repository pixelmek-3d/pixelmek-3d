package common

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func ScaleImage(img *ebiten.Image, scale float64, filter ebiten.Filter) *ebiten.Image {
	w, h := float64(img.Bounds().Dx()), float64(img.Bounds().Dy())
	scaledImg := ebiten.NewImage(int(w*scale), int(h*scale))
	op := &ebiten.DrawImageOptions{Filter: filter}
	op.GeoM.Scale(scale, scale)
	scaledImg.DrawImage(img, op)
	return scaledImg
}

func ScaleImageToHeight(img *ebiten.Image, pxHeight int, filter ebiten.Filter) *ebiten.Image {
	scale := float64(pxHeight) / float64(img.Bounds().Dy())
	return ScaleImage(img, scale, filter)
}

func ScaleImageToWidth(img *ebiten.Image, pxWidth int, filter ebiten.Filter) *ebiten.Image {
	scale := float64(pxWidth) / float64(img.Bounds().Dx())
	return ScaleImage(img, scale, filter)
}
