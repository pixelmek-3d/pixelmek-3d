package game

import (
	"image"
	"strconv"

	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
)

type TextureHandler struct {
	mapObj         *model.Map
	floorTex       *image.RGBA
	renderFloorTex bool
	texMap         map[string]*ebiten.Image
}

func NewTextureHandler(mapObj *model.Map) *TextureHandler {
	t := &TextureHandler{
		mapObj:         mapObj,
		renderFloorTex: true,
	}
	return t
}

func (t *TextureHandler) textureImage(texturePath string) *ebiten.Image {
	if img, ok := t.texMap[texturePath]; ok {
		return img
	}
	return nil
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

	// check if it has a side texture
	texObj := t.mapObj.GetMapTexture(strconv.Itoa(texNum))

	return t.textureImage(texObj.GetImage(side))
}

func (t *TextureHandler) FloorTexture() *image.RGBA {
	if t.renderFloorTex {
		return t.floorTex
	}
	return nil
}
