package render

import (
	"image"
	"image/draw"
	"image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

func SaveImageAsPNG(img *ebiten.Image, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// convert ebiten.Image to image.NRGBA for faster png encoding
	nrgba := image.NewNRGBA(img.Bounds())
	draw.Draw(nrgba, nrgba.Bounds(), img, img.Bounds().Min, draw.Src)

	return png.Encode(f, nrgba)
}
