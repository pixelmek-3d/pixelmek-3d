package colors

import "image/color"

var (
	Black         = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	White         = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	DefaultRed    = color.NRGBA{R: 225, G: 0, B: 0, A: 255}
	DefaultGreen  = color.NRGBA{R: 0, G: 214, B: 0, A: 255}
	DefaultBlue   = color.NRGBA{R: 0, G: 0, B: 203, A: 255}
	DefaultYellow = color.NRGBA{R: 255, G: 206, B: 0, A: 255}

	Enemy    = color.NRGBA{R: 255, G: 0, B: 12, A: 255}
	Friendly = color.NRGBA{R: 0, G: 255, B: 12, A: 255}
	NavPoint = DefaultYellow
)

func ColorToVec4(clr color.NRGBA) [4]float32 {
	return [4]float32{
		float32(clr.R) / 255,
		float32(clr.G) / 255,
		float32(clr.B) / 255,
		float32(clr.A) / 255,
	}
}

func RGBArrayToNRGBA(rgbArray [3]uint8) color.NRGBA {
	return color.NRGBA{
		R: rgbArray[0], G: rgbArray[1], B: rgbArray[2], A: 255,
	}
}
