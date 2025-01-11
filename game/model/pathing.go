package model

import (
	"fmt"

	"github.com/harbdog/go-astar"
	"github.com/harbdog/raycaster-go/geom"

	log "github.com/sirupsen/logrus"
)

type Pathing struct {
	// TODO: refactor Map/Mission to use Tile/astar.Pather interface
	world TileWorld
}

// TileWorld is a two dimensional map of Tiles.
type TileWorld map[int]map[int]*Tile

type TileKind int

const (
	TileKindPlain TileKind = iota
	TileKindBlocker
)

type Tile struct {
	// Kind is the kind of tile, potentially affecting movement.
	Kind TileKind
	// X and Y are the coordinates of the tile.
	X, Y int
	// W is a reference to the World that the tile is a part of.
	W TileWorld
}

// Tile gets the tile at the given coordinates in the world.
func (w TileWorld) Tile(x, y int) *Tile {
	if w[x] == nil {
		return nil
	}
	return w[x][y]
}

// SetTile sets a tile at the given coordinates in the world.
func (w TileWorld) SetTile(t *Tile, x, y int) {
	if w[x] == nil {
		w[x] = map[int]*Tile{}
	}
	w[x][y] = t
	t.X = x
	t.Y = y
	t.W = w
}

// PathNeighbors returns the neighbors of the tile, excluding blockers and
// tiles off the edge of the board.
func (t *Tile) PathNeighbors() []astar.Pather {
	neighbors := []astar.Pather{}
	for _, offset := range [][]int{
		{-1, 0},
		{1, 0},
		{0, -1},
		{0, 1},
	} {
		if n := t.W.Tile(t.X+offset[0], t.Y+offset[1]); n != nil && n.Kind != TileKindBlocker {
			neighbors = append(neighbors, n)
		}
	}
	return neighbors
}

// PathNeighborCost returns the movement cost of the directly neighboring tile.
func (t *Tile) PathNeighborCost(to astar.Pather) float64 {
	return 1.0 // TODO: implement differing costs for non-plain tile kinds (i.e. water, rough terrain)
}

// PathEstimatedCost uses Manhattan distance to estimate orthogonal distance
// between non-adjacent nodes.
func (t *Tile) PathEstimatedCost(to astar.Pather) float64 {
	toT := to.(*Tile)
	absX := toT.X - t.X
	if absX < 0 {
		absX = -absX
	}
	absY := toT.Y - t.Y
	if absY < 0 {
		absY = -absY
	}
	return float64(absX + absY)
}

func initPathing(m *Mission) *Pathing {
	width, height := m.missionMap.Size()
	w := TileWorld{}

	level := m.missionMap.Level(0)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			cell := level[x][y]

			kind := TileKindPlain
			if cell != 0 {
				kind = TileKindBlocker
			}
			w.SetTile(&Tile{Kind: kind}, x, y)
		}
	}

	return &Pathing{world: w}
}

func PathToString(path []*geom.Vector2) string {
	var pathStr string
	pathCount := len(path)
	for i, pos := range path {
		pathStr += fmt.Sprintf("(%0.2f, %0.2f)", pos.X, pos.Y)
		if i < pathCount-1 {
			pathStr += ","
		}
	}
	return "[" + pathStr + "]"
}

func (p *Pathing) FindPath(startPos, finishPos *geom.Vector2) []*geom.Vector2 {
	startTile := p.world.Tile(int(startPos.X), int(startPos.Y))
	finishTile := p.world.Tile(int(finishPos.X), int(finishPos.Y))
	path, _, found := astar.Path(startTile, finishTile)

	steps := make([]*geom.Vector2, 0, len(path))
	if !found {
		log.Errorf("unable to find path for (%0.0f,%0.0f) -> (%0.0f,%0.0f)", startPos.X, startPos.Y, finishPos.X, finishPos.Y)
		return steps
	}

	// astar path returned in reverse order
	for i := len(path) - 1; i >= 0; i-- {
		t := path[i].(*Tile)
		x, y := float64(t.X)+0.5, float64(t.Y)+0.5
		steps = append(steps, &geom.Vector2{X: x, Y: y})
	}

	return steps
}
