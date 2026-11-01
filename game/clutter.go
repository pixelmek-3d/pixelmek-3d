package game

import (
	"fmt"
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/common"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/sprites"
)

type ClutterHandler struct {
	sprites           map[*sprites.Sprite]struct{}
	spritesByPosition map[string][]*sprites.Sprite
	rng               *model.Rand
}

func NewClutterHandler() *ClutterHandler {
	c := &ClutterHandler{
		sprites:           make(map[*sprites.Sprite]struct{}, 256),
		spritesByPosition: make(map[string][]*sprites.Sprite, 256),
		rng:               model.NewRNG(),
	}
	return c
}

func (c *ClutterHandler) Update(g *Game, forceUpdate bool) {
	if !g.player.moved && !forceUpdate {
		// only update clutter position if camera moved or forceUpdate set
		return
	}

	numClutter := len(g.mission.Map().Clutter)
	if numClutter == 0 || g.clutterDistance <= 0 {
		return
	}

	pastPositions := make(map[string]struct{}, len(c.spritesByPosition))
	for posKey := range c.spritesByPosition {
		pastPositions[posKey] = struct{}{}
	}

	// determine which cells are in view for clutter consideration
	camPos, _, camHeading, _ := g.player.CameraPosition()
	camX, camY := camPos.X, camPos.Y

	cellRadius := math.Sqrt(0.5)
	halfFOV := geom.Radians(g.fovDegrees) / 2
	minX := int64(math.Floor(camX-g.clutterDistance)) - 1
	maxX := int64(math.Ceil(camX+g.clutterDistance)) + 1
	minY := int64(math.Floor(camY-g.clutterDistance)) - 1
	maxY := int64(math.Ceil(camY+g.clutterDistance)) + 1

	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= int64(g.mapWidth) {
		maxX = int64(g.mapWidth - 1)
	}
	if maxY >= int64(g.mapHeight) {
		maxY = int64(g.mapHeight - 1)
	}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float64(x) + 0.5 - camX
			dy := float64(y) + 0.5 - camY
			dist := math.Hypot(dx, dy)
			if dist > g.clutterDistance+cellRadius {
				continue
			}

			if dist > cellRadius {
				angle := math.Atan2(dy, dx)
				delta := math.Remainder(angle-camHeading, 2*math.Pi)
				anglePadding := math.Asin(cellRadius / dist)
				if math.Abs(delta) > halfFOV+anglePadding {
					continue
				}
			}

			// make sure there's not a wall here
			if g.mission.Map().IsWallAt(0, int(x), int(y)) {
				continue
			}

			// create x/y position key for tracking and hash for use in seed
			posKey := fmt.Sprintf("%d_%d", x, y)
			posHash := common.GetFNVHash(posKey)

			// remove entry from pastPositions so remainders can be cleaned after the loop
			delete(pastPositions, posKey)

			if _, ok := c.spritesByPosition[posKey]; ok {
				// clutter already loaded in view for position, move on
				continue
			}

			floorTexPath := g.tex.FloorTexturePathAt(int(x), int(y))

			// store sprite objects by position ID to make it easy to remove clutter when it goes outside of view
			c.spritesByPosition[posKey] = make([]*sprites.Sprite, numClutter)

			// use position based seed to produce consistent clutter positioning each time the coordinate is in view
			c.rng.Seed(g.mission.Map().Seed + int64(posHash))

			for i, clutter := range g.mission.Map().Clutter {
				// use floorPathMatch to determine if this clutter is for this coordinate based on floor texture
				if clutter.FloorPathMatch != nil && !clutter.FloorPathMatch.MatchString(floorTexPath) {
					continue
				}

				chanceToAppear := c.rng.Float64() <= clutter.Frequency
				if !chanceToAppear {
					continue
				}

				clutterImg := g.tex.TextureImage(clutter.Image)
				clutterScale := clutter.Height / model.METERS_PER_UNIT
				cSprite := sprites.NewSprite(
					model.BasicVisualEntity(
						float64(x)+c.rng.Float64(),
						float64(y)+c.rng.Float64(),
						0,
						raycaster.AnchorBottom,
					),
					clutterScale,
					clutterImg,
				)

				// store clutter sprites and which coordinate position id they are in
				c.sprites[cSprite] = struct{}{}
				c.spritesByPosition[posKey][i] = cSprite
			}
		}
	}

	// clean up by posId sprites which are not in current view
	for posKey := range pastPositions {
		if spriteList, ok := c.spritesByPosition[posKey]; ok {
			for _, sprite := range spriteList {
				delete(c.sprites, sprite)
			}
		}
		delete(c.spritesByPosition, posKey)
	}
}
