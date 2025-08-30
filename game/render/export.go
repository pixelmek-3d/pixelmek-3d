package render

import (
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
	return png.Encode(f, img)
}
