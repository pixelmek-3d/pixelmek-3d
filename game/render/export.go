package render

import (
	"image"
	"image/draw"
	"image/gif"
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

func SaveAnimatedGIF(frames []*image.Paletted, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	numFrames := len(frames)

	gifOptions := gif.GIF{
		Image:     frames,
		Delay:     make([]int, numFrames),
		LoopCount: 0,
	}

	// Set delay for each frame (1/60s = approx 1.67 centiseconds, use 2 for delay value)
	for i := range numFrames {
		gifOptions.Delay[i] = 2 // 2*1/100 second = 20ms delay per frame
	}

	err = gif.EncodeAll(f, &gifOptions)
	if err != nil {
		return err
	}
	return nil
}
