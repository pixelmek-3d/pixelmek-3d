package game

import (
	"image"

	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
)

type TextureHandler struct {
	mapObj         *model.Map
	textures       []*ebiten.Image
	floorTex       *image.RGBA
	renderFloorTex bool
}

func NewTextureHandler(mapObj *model.Map) *TextureHandler {
	t := &TextureHandler{
		mapObj:         mapObj,
		renderFloorTex: true,
	}
	return t
}

func (t *TextureHandler) TextureAt(x, y, levelNum, side int) *ebiten.Image {
	texNum := -1

	mapLevel := t.mapObj.Level(levelNum)
	if mapLevel == nil {
		return nil
	}

	mapWidth := len(mapLevel)
	if mapWidth == 0 {
		return nil
	}
	mapHeight := len(mapLevel[0])
	if mapHeight == 0 {
		return nil
	}

	if x >= 0 && x < mapWidth && y >= 0 && y < mapHeight {
		texNum = mapLevel[x][y]
	}

	if texNum <= 0 {
		return nil
	}
	return t.textures[texNum]
}

func (t *TextureHandler) FloorTexture() *image.RGBA {
	if t.renderFloorTex {
		return t.floorTex
	}
	return nil
}
