package colors

import "image/color"

var (
	DefaultRed    = color.NRGBA{R: 225, G: 0, B: 0, A: 255}
	DefaultGreen  = color.NRGBA{R: 0, G: 214, B: 0, A: 255}
	DefaultBlue   = color.NRGBA{R: 0, G: 0, B: 203, A: 255}
	DefaultYellow = color.NRGBA{R: 255, G: 206, B: 0, A: 255}

	Enemy    = color.NRGBA{R: 255, G: 0, B: 12, A: 255}
	Friendly = color.NRGBA{R: 0, G: 255, B: 12, A: 255}
	NavPoint = DefaultYellow
)
