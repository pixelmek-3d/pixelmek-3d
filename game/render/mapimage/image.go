package mapimage

import (
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
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

	// first, draw floor texture layer
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
			op.GeoM.Translate(float64(x*pxPerCell), float64(y*pxPerCell))
			mapImage.DrawImage(cellImg, op)
		}
	}

	return mapImage, nil
}
