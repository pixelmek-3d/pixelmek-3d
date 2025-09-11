package shapes

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// StrokeDiamond strokes a diamond shape with the specified center position (cx, cy), the directional radius (rx, ry), width and color.
//
// clr has be to be a solid (non-transparent) color.
func StrokeDiamond(dst *ebiten.Image, cx, cy, rx, ry, strokeWidth float32, clr color.Color, antialias bool) {
	minX, minY := cx-rx, cy-ry
	maxX, maxY := cx+rx, cy+ry
	midX, midY := cx, cy
	vector.StrokeLine(dst, minX, midY, midX, minY, strokeWidth, clr, antialias)
	vector.StrokeLine(dst, midX, minY, maxX, midY, strokeWidth, clr, antialias)
	vector.StrokeLine(dst, minX, midY, midX, maxY, strokeWidth, clr, antialias)
	vector.StrokeLine(dst, midX, maxY, maxX, midY, strokeWidth, clr, antialias)
}
