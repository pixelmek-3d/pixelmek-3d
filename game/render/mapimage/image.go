package mapimage

import (
	"errors"
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	"github.com/pixelmek-3d/pixelmek-3d/game/texture"
)

type MapImageOptions struct {
	PxPerCell                 int
	RenderDefaultFloorTexture bool
	FilterDefaultFloorTexture bool
	RenderWallLines           bool
	RenderGridLines           bool
	GridCellDistance          int
}

func NewMapImage(m *model.Map, tex *texture.TextureHandler, opts MapImageOptions) (*ebiten.Image, error) {
	if m == nil || tex == nil {
		return nil, errors.New("map image called with nil map or texture handler")
	}
	pxPerCell := max(opts.PxPerCell, 1)
	texScale := float64(pxPerCell) / float64(resources.TexSize)
	mapWidth, mapHeight := m.Size()
	mapImage := ebiten.NewImage(mapWidth*pxPerCell, mapHeight*pxPerCell)

	// fill with basic floor color layer based on static floor image
	floorImg := resources.GetTextureFromFile(m.FloorBox.Image)
	if floorImg != nil {
		centerX, centerY := floorImg.Bounds().Dx()/2, floorImg.Bounds().Dy()/2
		mapImage.Fill(floorImg.At(centerX, centerY))
	}

	if opts.RenderDefaultFloorTexture {
		if opts.FilterDefaultFloorTexture {
			// draw default floor texture layer at higher resolution then scale down to reduce hatch grid effect
			hiResFloorPxPerCell := geom.ClampInt(pxPerCell*int(math.Pow(2, 3)), 1, 64)
			hiResFloorImg := ebiten.NewImage(mapWidth*hiResFloorPxPerCell, mapHeight*hiResFloorPxPerCell)
			drawFloorTextures(m, tex, hiResFloorImg, drawFloorTexturesOptions{
				drawDefaultFloorTexture: true,
				pxPerCell:               hiResFloorPxPerCell,
				filter:                  ebiten.FilterLinear,
			})

			// draw downscaled default floor texture image to fit target
			hiResDownScale := float64(hiResFloorPxPerCell) / float64(pxPerCell)
			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterLinear
			op.GeoM.Scale(hiResDownScale, hiResDownScale)
			mapImage.DrawImage(hiResFloorImg, op)
		} else {
			// draw default first floor textures
			drawFloorTextures(m, tex, mapImage, drawFloorTexturesOptions{
				drawDefaultFloorTexture: true,
				pxPerCell:               pxPerCell,
				filter:                  ebiten.FilterNearest,
			})
		}
	}

	// draw non-default first floor textures
	drawFloorTextures(m, tex, mapImage, drawFloorTexturesOptions{
		drawOtherFloorTexture: true,
		pxPerCell:             pxPerCell,
		filter:                ebiten.FilterNearest,
	})

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

	if opts.RenderGridLines {
		kCells := opts.GridCellDistance
		if kCells <= 0 {
			// default drawing distance grid lines every 1km worth of map cells
			kCells = int(1000 / model.METERS_PER_UNIT)
		}
		for x := 0; x < mapWidth; x += kCells {
			x1, x2 := float32(x*pxPerCell), float32(x*pxPerCell)
			y1, y2 := float32(0), float32(mapHeight*pxPerCell)
			vector.StrokeLine(mapImage, x1, y1, x2, y2, 1, color.White, false)
		}
		for y := 0; y < mapHeight; y += kCells {
			x1, x2 := float32(0), float32(mapWidth*pxPerCell)
			y1, y2 := float32(y*pxPerCell), float32(y*pxPerCell)
			vector.StrokeLine(mapImage, x1, y1, x2, y2, 1, color.White, false)
		}
	}

	return mapImage, nil
}

type drawFloorTexturesOptions struct {
	pxPerCell               int
	filter                  ebiten.Filter
	drawDefaultFloorTexture bool
	drawOtherFloorTexture   bool
}

func drawFloorTextures(m *model.Map, tex *texture.TextureHandler, mapImage *ebiten.Image, opts drawFloorTexturesOptions) error {
	mapWidth, mapHeight := m.Size()
	defaultFloorTexturePath := tex.DefaultFloorTexturePath()
	texScale := float64(opts.pxPerCell) / float64(resources.TexSize)
	for x := range mapWidth {
		for y := range mapHeight {
			cellTexPath := tex.FloorTexturePathAt(x, y)
			if (!opts.drawDefaultFloorTexture && cellTexPath == defaultFloorTexturePath) ||
				(!opts.drawOtherFloorTexture && cellTexPath != defaultFloorTexturePath) {
				continue
			}
			cellImg := tex.TextureImage(cellTexPath)
			if cellImg == nil {
				return fmt.Errorf("map image failed to load cell texture at (%d,%d): %s", x, y, cellTexPath)
			}
			op := &ebiten.DrawImageOptions{}
			op.Filter = opts.filter
			op.GeoM.Scale(texScale, texScale)
			op.GeoM.Translate(float64(x*opts.pxPerCell), float64((mapHeight-y-1)*opts.pxPerCell))
			mapImage.DrawImage(cellImg, op)
		}
	}
	return nil
}
