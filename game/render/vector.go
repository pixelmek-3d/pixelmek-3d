package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/vector"
)

type VectorPath struct {
	Color   color.NRGBA
	Vectors [][2]float32
}

func (v VectorPath) GetRenderPath() *vector.Path {
	path := &vector.Path{}
	for i, point := range v.Vectors {
		if i == 0 {
			path.MoveTo(point[0], point[1])
		} else {
			path.LineTo(point[0], point[1])
		}
	}
	path.Close()
	return path
}
