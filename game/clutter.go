package game

import (
	"image/color"
	"math/rand"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type ClutterHandler struct {
	sprites            map[*model.Sprite]struct{}
	spritesByPosition  map[int64][]*model.Sprite
	maxClutterDistance float64
}

func NewClutterHandler(maxClutterDistance float64) *ClutterHandler {
	c := &ClutterHandler{
		sprites:            make(map[*model.Sprite]struct{}, 256),
		spritesByPosition:  make(map[int64][]*model.Sprite, 256),
		maxClutterDistance: maxClutterDistance,
	}
	return c
}

func (c *ClutterHandler) Update(g *Game, forceUpdate bool) {
	if !g.player.Moved && !forceUpdate {
		// only update clutter position if camera moved or forceUpdate set
		return
	}

	numClutter := len(g.mapObj.Clutter)

	pastPositions := make(map[int64]struct{}, len(c.spritesByPosition))
	for posId := range c.spritesByPosition {
		pastPositions[posId] = struct{}{}
	}

	// determine which cells are in view for clutter consideration
	camX, camY := g.player.Position.X, g.player.Position.Y
	viewAngle := g.player.Angle
	viewFOV := geom.Radians(g.fovDegrees)
	for a := viewAngle - viewFOV/2; a <= viewAngle+viewFOV/2; a += viewFOV / 20 {
		for d := 0.0; d <= c.maxClutterDistance; d++ {
			line := geom.LineFromAngle(camX, camY, a, d)
			x, y := int64(line.X2), int64(line.Y2)

			// create x/y position ID for tracking and use in seed
			posId := (x-1)*int64(g.mapWidth) + y

			// remove entry from pastPositions so remainders can be cleaned after the loop
			delete(pastPositions, posId)

			if _, ok := c.spritesByPosition[posId]; ok {
				// clutter already loaded in view for position, move on
				continue
			}

			// store sprite objects by position ID to make it easy to remove clutter when it goes outside of view
			c.spritesByPosition[posId] = make([]*model.Sprite, numClutter)

			// position based seed to produce consistent clutter positioning each time the coordinate is in view
			rand.Seed(g.mapObj.Seed + posId)

			for i, clutter := range g.mapObj.Clutter {

				// TODO: look up floor texture name for floorPathMatch to determine if this clutter is appropriate for this coordinate

				chanceToAppear := rand.Float64() <= clutter.Frequency
				if !chanceToAppear {
					continue
				}

				clutterImg := g.tex.texMap[clutter.Image]
				cSprite := model.NewSprite(
					float64(x)+rand.Float64(), float64(y)+rand.Float64(), clutter.Scale, clutterImg,
					color.RGBA{}, raycaster.AnchorBottom, 0,
				)

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
