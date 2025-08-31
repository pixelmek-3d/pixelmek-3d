package mapimage

import (
	"errors"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	"github.com/pixelmek-3d/pixelmek-3d/game/texture"
)

func NewMapImage(m *model.Map, tex *texture.TextureHandler, pxPerCell int) (*ebiten.Image, error) {
	if m == nil || tex == nil {
		return nil, errors.New("map image called with nil map or texture handler")
	}
	if pxPerCell < 1 {
		pxPerCell = 1
	}
	mapWidth, mapHeight := m.Size()
	mapImage := ebiten.NewImage(mapWidth*pxPerCell, mapHeight*pxPerCell)

	// draw floor texture layer
	texScale := float64(pxPerCell) / float64(resources.TexWidth)
	//level := m.Level(0)
	for x := range mapWidth {
		for y := range mapHeight {
			cellTexPath := tex.FloorTexturePathAt(x, y)
			cellImg := tex.TextureImage(cellTexPath)
			if cellImg == nil {
				return nil, fmt.Errorf("map image failed to load cell texture at (%d,%d): %s", x, y, cellTexPath)
			}
			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterNearest
			op.GeoM.Scale(texScale, texScale)
			op.GeoM.Translate(float64(x*pxPerCell), float64((mapHeight-y-1)*pxPerCell))
			mapImage.DrawImage(cellImg, op)
		}
	}

	// draw first level wall texture layer
	for x := range mapWidth {
		for y := range mapHeight {
			cellImg := tex.TextureAt(x, y, 0, 0)
			if cellImg == nil {
				continue
			}
			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterNearest
			op.GeoM.Scale(texScale, texScale)
			op.GeoM.Translate(float64(x*pxPerCell), float64((mapHeight-y-1)*pxPerCell))
			mapImage.DrawImage(cellImg, op)
		}
	}

	// draw collision lines around walls
	for _, line := range m.GenerateWallCollisionLines(0) {
		x1, x2 := line.X1*float64(pxPerCell), line.X2*float64(pxPerCell)
		y1, y2 := (float64(mapHeight)-line.Y1)*float64(pxPerCell), (float64(mapHeight)-line.Y2)*float64(pxPerCell)
		vector.StrokeLine(mapImage, float32(x1), float32(y1), float32(x2), float32(y2), 1, color.NRGBA{R: 255, G: 0, B: 0, A: 255}, false)
	}

	// TODO: draw static map sprites

	return mapImage, nil
}
