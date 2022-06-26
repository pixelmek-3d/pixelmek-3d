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

func NewTextureHandler(mapObj *model.Map, textureCapacity int) *TextureHandler {
	t := &TextureHandler{
		mapObj:         mapObj,
		textures:       make([]*ebiten.Image, textureCapacity),
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
		texNum = mapLevel[x][y] - 1 // 1 subtracted from it so that texture 0 can be used
	}

	if side == 0 {
		//--some supid hacks to make the houses render correctly--//
		// this corrects textures on two sides of house since the textures are not symmetrical
		if texNum == 3 {
			texNum = 4
		} else if texNum == 4 {
			texNum = 3
		}

		if texNum == 1 {
			texNum = 4
		} else if texNum == 2 {
			texNum = 3
		}

		// make the ebitengine splash only show on one side
		if texNum == 5 {
			texNum = 0
		}
	}

	if texNum < 0 {
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
