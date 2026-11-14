package render

import (
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"path"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/setanarut/apng"
	log "github.com/sirupsen/logrus"
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

func SaveAnimation(imgs []*ebiten.Image, bounds image.Rectangle, filename string) error {
	extension := path.Ext(filename)
	switch extension {
	case ".apng", ".png":
		return SaveAnimatedPNG(imgs, bounds, filename)
	default:
		return SaveAnimatedGIF(imgs, bounds, filename)
	}
}

func SaveAnimatedGIF(imgs []*ebiten.Image, bounds image.Rectangle, filename string) error {
	log.Debug("export as animated GIF...")
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

func SaveAnimatedPNG(imgs []*ebiten.Image, bounds image.Rectangle, filename string) error {
	log.Debug("export as animated PNG...")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	frames := make([]image.Image, len(imgs))
	delays := make([]uint16, len(imgs))
	frameBounds := image.Rect(0, 0, bounds.Dx(), bounds.Dy())

	for i, img := range imgs {
		frame := image.NewRGBA(frameBounds)
		img.ReadPixels(frame.Pix)
		frames[i] = frame
		delays[i] = 2 // 2 centiseconds = 20 ms
	}

	return apng.EncodeAll(f, &apng.APNG{
		Images:    frames,
		Delays:    delays,
		LoopCount: 0,
	})
}
