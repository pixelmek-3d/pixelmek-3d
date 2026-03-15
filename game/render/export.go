package render

import (
	"image"
	"image/color/palette"
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

func SaveAnimatedGIF(imgs []*ebiten.Image, bounds image.Rectangle, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	var frames []*image.Paletted
	for _, img := range imgs {
		// Read pixels from ebiten.Image into a RGBA image
		rgbaImg := image.NewRGBA(bounds)
		img.ReadPixels(rgbaImg.Pix)

		// Create the target paletted image
		pImg := image.NewPaletted(bounds, palette.WebSafe)

		// Draw with quantization to map colors
		draw.FloydSteinberg.Draw(pImg, bounds, rgbaImg, image.Point{})

		frames = append(frames, pImg)
	}

	gifOptions := gif.GIF{
		Image:     frames,
		Delay:     make([]int, len(frames)),
		LoopCount: 0,
	}

	// Set delay for each frame (1/60s = approx 1.67 centiseconds, use 2 for delay value)
	for i := range len(frames) {
		gifOptions.Delay[i] = 2 // 2*1/100 second = 20ms delay per frame
	}

	err = gif.EncodeAll(f, &gifOptions)
	if err != nil {
		return err
	}
	return nil
}
