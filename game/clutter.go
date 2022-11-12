package game

import (
	"image/color"
	"math/rand"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type ClutterHandler struct {
	sprites           map[*render.Sprite]struct{}
	spritesByPosition map[int64][]*render.Sprite
}

func NewClutterHandler() *ClutterHandler {
	c := &ClutterHandler{
		sprites:           make(map[*render.Sprite]struct{}, 256),
		spritesByPosition: make(map[int64][]*render.Sprite, 256),
	}
	return c
}

func (c *ClutterHandler) Update(g *Game, forceUpdate bool) {
	if !g.player.Moved && !forceUpdate {
		// only update clutter position if camera moved or forceUpdate set
		return
	}

	numClutter := len(g.mission.Map().Clutter)
	if numClutter == 0 || g.clutterDistance <= 0 {
		return
	}

	pastPositions := make(map[int64]struct{}, len(c.spritesByPosition))
	for posId := range c.spritesByPosition {
		pastPositions[posId] = struct{}{}
	}

	// determine which cells are in view for clutter consideration
	camX, camY := g.player.Pos().X, g.player.Pos().Y
	viewAngle := g.player.Heading()
	viewFOV := geom.Radians(g.fovDegrees)
	for a := viewAngle - viewFOV/2; a <= viewAngle+viewFOV/2; a += viewFOV / 20 {
		for d := 0.0; d <= g.clutterDistance; d++ {
			line := geom.LineFromAngle(camX, camY, a, d)
			x, y := int64(line.X2), int64(line.Y2)
			if x < 0 || y < 0 || x >= int64(g.mapWidth) || y >= int64(g.mapHeight) {
				continue
			}

			// create x/y position ID for tracking and use in seed
			posId := (x-1)*int64(g.mapWidth) + y

			// make sure there's not a wall here
			if g.mission.Map().IsWallAt(0, int(x), int(y)) {
				continue
			}

			// remove entry from pastPositions so remainders can be cleaned after the loop
			delete(pastPositions, posId)

			if _, ok := c.spritesByPosition[posId]; ok {
				// clutter already loaded in view for position, move on
				continue
			}

			floorTexPath := g.tex.floorTexturePathAt(int(x), int(y))

			// store sprite objects by position ID to make it easy to remove clutter when it goes outside of view
			c.spritesByPosition[posId] = make([]*render.Sprite, numClutter)

			// use position based seed to produce consistent clutter positioning each time the coordinate is in view
			rand.Seed(g.mission.Map().Seed + posId)

			for i, clutter := range g.mission.Map().Clutter {
				// use floorPathMatch to determine if this clutter is for this coordinate based on floor texture
				if clutter.FloorPathMatch != nil && !clutter.FloorPathMatch.MatchString(floorTexPath) {
					continue
				}

				chanceToAppear := rand.Float64() <= clutter.Frequency
				if !chanceToAppear {
					continue
				}

				clutterImg := g.tex.texMap[clutter.Image]
				cSprite := render.NewSprite(
					model.BasicVisualEntity(float64(x)+rand.Float64(), float64(y)+rand.Float64(), 0, raycaster.AnchorBottom),
					clutter.Scale, clutterImg, color.RGBA{},
				)

				// store clutter sprites and which coordinate position id they are in
				c.sprites[cSprite] = struct{}{}
				c.spritesByPosition[posId][i] = cSprite
			}
		}
	}

	// clean up by posId sprites which are not in current view
	for posId := range pastPositions {
		if spriteList, ok := c.spritesByPosition[posId]; ok {
			for _, sprite := range spriteList {
				delete(c.sprites, sprite)
			}
		}
		delete(c.spritesByPosition, posId)
	}
}
