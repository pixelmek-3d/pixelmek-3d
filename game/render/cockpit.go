package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
)

const cockpitScale float64 = 1.2

var DefaultCockpit Cockpit

type Cockpit struct {
	vectorBounds image.Rectangle
	vectorPaths  []VectorPath
}

type CockpitRenderOptions struct {
	StrideZ        float64
	StrideStompDir model.StrideStomp
}

func init() {
	DefaultCockpit = Cockpit{
		vectorBounds: image.Rect(0, 0, 1024, 768),
		vectorPaths: []VectorPath{
			{ // left
				Color: color.NRGBA{R: 60, G: 60, B: 60, A: 255},
				Vectors: [][2]float32{
					{0, 150},
					{415, 615},
					{360, 620},
					{0, 210},
				},
			},
			{ // bottom left
				Color: color.NRGBA{R: 30, G: 30, B: 30, A: 255},
				Vectors: [][2]float32{
					{80, 768},
					{360, 620},
					{415, 615},
					{250, 768},
				},
			},
			{ // right
				Color: color.NRGBA{R: 60, G: 60, B: 60, A: 255},
				Vectors: [][2]float32{
					{1024, 150},
					{609, 615},
					{664, 620},
					{1024, 210},
				},
			},
			{ // bottom right
				Color: color.NRGBA{R: 30, G: 30, B: 30, A: 255},
				Vectors: [][2]float32{
					{944, 768},
					{664, 620},
					{609, 615},
					{774, 768},
				},
			},
			{ // top left
				Color: color.NRGBA{R: 30, G: 30, B: 30, A: 255},
				Vectors: [][2]float32{
					{335, 0},
					{405, 115},
					{416, 166},
					{134, 0},
				},
			},
			{ // top right
				Color: color.NRGBA{R: 30, G: 30, B: 30, A: 255},
				Vectors: [][2]float32{
					{689, 0},
					{619, 115},
					{608, 166},
					{890, 0},
				},
			},
			{ // top center
				Color: color.NRGBA{R: 60, G: 60, B: 60, A: 255},
				Vectors: [][2]float32{
					{405, 115},
					{619, 115},
					{608, 166},
					{416, 166},
				},
			},
		},
	}
}

// converts all vector paths into path objects that ebitengine can render combined by color
func (c Cockpit) getRenderPaths(geoM ebiten.GeoM) map[color.Color]*vector.Path {
	renderPaths := make(map[color.Color]*vector.Path, len(c.vectorPaths))
	for _, cvp := range c.vectorPaths {
		if _, exists := renderPaths[cvp.Color]; !exists {
			renderPaths[cvp.Color] = &vector.Path{}
		}
		path := renderPaths[cvp.Color]
		path.AddPath(cvp.GetRenderPath(), &vector.AddPathOptions{GeoM: geoM})
	}
	return renderPaths
}

func DrawCockpit(screen *ebiten.Image, cockpit Cockpit, opts CockpitRenderOptions) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()

	// scale from given vector bounds to actual resolution
	geoM := ebiten.GeoM{}
	geoM.Scale(cockpitScale*float64(w)/float64(cockpit.vectorBounds.Dx()), cockpitScale*float64(h)/float64(cockpit.vectorBounds.Dy()))

	// translate to center since cockpit is overscaled to allow for movement without ending before screen edges
	maxOffX, maxOffY := float64(w)*(cockpitScale-1)/2, float64(h)*(cockpitScale-1)/2
	geoM.Translate(-maxOffX, -maxOffY)

	if opts.StrideZ != 0 {
		// apply stride offsets to move cockpit around, clamp to max offsets to avoid ending before screen edges
		strideY := geom.Clamp(opts.StrideZ*float64(h), -maxOffY, maxOffY)
		var strideX float64
		switch opts.StrideStompDir {
		case model.StrideStompLeft:
			strideX = strideY / 2
		case model.StrideStompRight:
			strideX = -strideY / 2
		}
		geoM.Translate(strideX, -strideY)
	}

	renderPaths := cockpit.getRenderPaths(geoM)
	for clr, path := range renderPaths {
		drawVectorPath(screen, path, clr)
	}
}

func drawVectorPath(screen *ebiten.Image, path *vector.Path, clr color.Color) {
	drawOp := &vector.DrawPathOptions{}
	drawOp.AntiAlias = false
	drawOp.ColorScale.ScaleWithColor(clr)
	vector.FillPath(screen, path, nil, drawOp)
}
