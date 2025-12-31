package texture

import (
	"image"

	"github.com/harbdog/raycaster-go/geom"
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
		t.loadMapTextures()
	}
	return t
}

func (t *TextureHandler) IsHandlerForMap(mapObj *model.Map) bool {
	if t.mapObj == nil {
		return false
	}
	return t.mapObj == mapObj
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

func NewFloorTexture(texture string) *FloorTexture {
	f := &FloorTexture{
		image: resources.GetRGBAFromFile(texture),
		path:  texture,
	}
	return f
}

func (t *TextureHandler) SetDefaultFloorTexturePath(floorTexPath string) {
	t.floorTexDefault = NewFloorTexture(floorTexPath)
}

func (t *TextureHandler) DefaultFloorTexturePath() string {
	return t.floorTexDefault.path
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

func (t *TextureHandler) FloorTexturePathAt(x, y int) string {
	if len(t.floorTexMap) > 0 {
		tex := t.floorTexMap[x][y]
		if tex != nil {
			return tex.path
		}
	}
	return t.floorTexDefault.path
}

func (t *TextureHandler) loadMapTextures() {
	// load textured flooring
	if t.mapObj.Flooring.Default != "" {
		t.SetDefaultFloorTexturePath(t.mapObj.Flooring.Default)
		t.SetTextureImage(t.mapObj.Flooring.Default, resources.GetTextureFromFile(t.mapObj.Flooring.Default))
	}

	// load texture floor pathing
	if len(t.mapObj.Flooring.Pathing) > 0 {
		// create map grid of path image textures for the X/Y coords indicated
		for _, pathing := range t.mapObj.Flooring.Pathing {
			tex := NewFloorTexture(pathing.Image)
			t.SetTextureImage(pathing.Image, resources.GetTextureFromFile(pathing.Image))

			// create filled rectangle paths
			for _, rect := range pathing.Rects {
				x0, y0, x1, y1 := rect[0][0], rect[0][1], rect[1][0], rect[1][1]
				for x := x0; x <= x1; x++ {
					for y := y0; y <= y1; y++ {
						t.SetFloorTextureAt(x, y, tex)
					}
				}
			}

			// create line segment paths
			for _, segments := range pathing.Lines {
				var prevPoint *geom.Vector2
				for _, seg := range segments {
					point := &geom.Vector2{X: float64(seg[0]), Y: float64(seg[1])}

					if prevPoint != nil {
						// fill in path for line segment from previous to next point
						line := geom.Line{X1: prevPoint.X, Y1: prevPoint.Y, X2: point.X, Y2: point.Y}

						// use the angle of the line to then find every coordinate along the line path
						angle := line.Angle()
						dist := geom.Distance(line.X1, line.Y1, line.X2, line.Y2)
						for d := 0.0; d <= dist; d += 0.1 {
							nLine := geom.LineFromAngle(line.X1, line.Y1, angle, d)
							t.SetFloorTextureAt(int(nLine.X2), int(nLine.Y2), tex)
						}
					}

					prevPoint = point
				}
			}
		}
	}

	// load clutter sprites mapped by path
	if len(t.mapObj.Clutter) > 0 {
		for _, clutter := range t.mapObj.Clutter {
			if img := t.TextureImage(clutter.Image); img == nil {
				t.SetTextureImage(clutter.Image, resources.GetSpriteFromFile(clutter.Image))
			}
		}
	}

	// load textures mapped by path
	for _, tex := range t.mapObj.Textures {
		if tex.Image != "" {
			if img := t.TextureImage(tex.Image); img == nil {
				t.SetTextureImage(tex.Image, resources.GetTextureFromFile(tex.Image))
			}
		}

		if tex.SideX != "" {
			if img := t.TextureImage(tex.SideX); img == nil {
				t.SetTextureImage(tex.SideX, resources.GetTextureFromFile(tex.SideX))
			}
		}

		if tex.SideY != "" {
			if img := t.TextureImage(tex.SideY); img == nil {
				t.SetTextureImage(tex.SideY, resources.GetTextureFromFile(tex.SideY))
			}
		}
	}
}
