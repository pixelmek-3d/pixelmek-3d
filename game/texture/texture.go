package texture

import (
	"image"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	"github.com/hajimehoshi/ebiten/v2"
)

type TextureHandler struct {
	mapObj          *model.Map
	texMap          map[string]*ebiten.Image
	renderFloorTex  bool
	floorTexDefault *FloorTexture
	floorTexMap     [][]*FloorTexture
}

type FloorTexture struct {
	image *image.RGBA
	path  string
}

func NewTextureHandler(mapObj *model.Map) *TextureHandler {
	t := &TextureHandler{
		mapObj:         mapObj,
		texMap:         make(map[string]*ebiten.Image, 128),
		renderFloorTex: true,
	}
	if mapObj != nil {
		mapWidth, mapHeight := mapObj.Size()
		t.floorTexMap = make([][]*FloorTexture, mapWidth)
		for x := 0; x < mapWidth; x++ {
			t.floorTexMap[x] = make([]*FloorTexture, mapHeight)
		}
	}
	return t
}

func (t *TextureHandler) RenderFloorTex() bool {
	return t.renderFloorTex
}

func (t *TextureHandler) SetRenderFloorTex(renderFloorTex bool) {
	t.renderFloorTex = renderFloorTex
}

func (t *TextureHandler) TextureImage(texturePath string) *ebiten.Image {
	if img, ok := t.texMap[texturePath]; ok {
		return img
	}
	return nil
}

func (t *TextureHandler) SetTextureImage(texturePath string, textureImg *ebiten.Image) {
	t.texMap[texturePath] = textureImg
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
	texObj := t.mapObj.GetMapTexture(texNum)

	return t.TextureImage(texObj.GetImage(side))
}

func (t *TextureHandler) SetDefaultFloorTexture(floorTex *FloorTexture) {
	t.floorTexDefault = floorTex
}

func (t *TextureHandler) SetFloorTextureAt(x, y int, floorTex *FloorTexture) {
	t.floorTexMap[x][y] = floorTex
}

func (t *TextureHandler) FloorTextureAt(x, y int) *image.RGBA {
	if x < 0 || y < 0 {
		return nil
	}
	if t.renderFloorTex {
		if len(t.floorTexMap) > 0 {
			tex := t.floorTexMap[x][y]
			if tex != nil {
				return tex.image
			}
		}
		return t.floorTexDefault.image
	}
	return nil
}

func NewFloorTexture(texture string) *FloorTexture {
	f := &FloorTexture{
		image: resources.GetRGBAFromFile(texture),
		path:  texture,
	}
	return f
}

func (t *TextureHandler) FloorTexturePathAt(x, y int) string {
	if len(t.floorTexMap) > 0 {
		tex := t.floorTexMap[x][y]
		if tex != nil {
			return tex.path
		}
	}
	return t.floorTexDefault.path
}
