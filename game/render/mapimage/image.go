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

type MapImageOptions struct {
	PxPerCell                 int
	RenderDefaultFloorTexture bool
	RenderWallLines           bool
}

func NewMapImage(m *model.Map, tex *texture.TextureHandler, opts MapImageOptions) (*ebiten.Image, error) {
	if m == nil || tex == nil {
		return nil, errors.New("map image called with nil map or texture handler")
	}
	pxPerCell := opts.PxPerCell
	if pxPerCell < 1 {
		pxPerCell = 1
	}
	mapWidth, mapHeight := m.Size()
	mapImage := ebiten.NewImage(mapWidth*pxPerCell, mapHeight*pxPerCell)

	// fill with basic floor color layer based on static floor image
	floorImg := resources.GetTextureFromFile(m.FloorBox.Image)
	if floorImg != nil {
		centerX, centerY := floorImg.Bounds().Dx()/2, floorImg.Bounds().Dy()/2
		mapImage.Fill(floorImg.At(centerX, centerY))
	}

	// draw floor texture layer
	defaultFloorTexturePath := tex.DefaultFloorTexturePath()
	texScale := float64(pxPerCell) / float64(resources.TexWidth)
	for x := range mapWidth {
		for y := range mapHeight {
			cellTexPath := tex.FloorTexturePathAt(x, y)
			if !opts.RenderDefaultFloorTexture && cellTexPath == defaultFloorTexturePath {
				continue
			}
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

	if opts.RenderWallLines {
		// draw collision lines around walls
		for _, line := range m.GenerateWallCollisionLines(0) {
			x1, x2 := line.X1*float64(pxPerCell), line.X2*float64(pxPerCell)
			y1, y2 := (float64(mapHeight)-line.Y1)*float64(pxPerCell), (float64(mapHeight)-line.Y2)*float64(pxPerCell)
			vector.StrokeLine(mapImage, float32(x1), float32(y1), float32(x2), float32(y2), 1, color.NRGBA{R: 255, G: 0, B: 0, A: 255}, false)
		}
	}

	// draw static map sprites
	for _, s := range m.Sprites {
		if len(s.Image) == 0 {
			continue
		}

		spriteImg := tex.TextureImage(s.Image)
		if spriteImg == nil {
			spriteImg = resources.GetSpriteFromFile(s.Image)
			tex.SetTextureImage(s.Image, spriteImg)
		}

		// convert sprite height to cell size and then use pixels per cell based on sprite height
		spriteWidth, spriteHeight := float64(spriteImg.Bounds().Dx()), float64(spriteImg.Bounds().Dy())
		scale := (s.Height / model.METERS_PER_UNIT) * (float64(pxPerCell) / spriteHeight)

		for _, position := range s.Positions {
			// adjust orientation as sprites are centered at X and bottomed at Y
			x := (position[0] * float64(pxPerCell)) - (spriteWidth*scale)/2
			y := ((float64(mapHeight) - position[1]) * float64(pxPerCell)) - (spriteHeight * scale)

			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterNearest
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(x, y)
			mapImage.DrawImage(spriteImg, op)
		}
	}

	return mapImage, nil
}
