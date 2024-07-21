package model

import (
	"fmt"
	"math"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/quasilyte/pathing"
)

type Pathing struct {
	grid       *pathing.Grid
	pathfinder pathBuilder
	layer      pathing.GridLayer
}

type pathBuilder interface {
	BuildPath(g *pathing.Grid, from, to pathing.GridCoord, l pathing.GridLayer) pathing.BuildPathResult
}

func initPathing(m *Mission) *Pathing {
	cellSize := uint(METERS_PER_UNIT)
	width, height := m.missionMap.Size()
	grid := pathing.NewGrid(pathing.GridConfig{
		WorldWidth:  uint(width) * cellSize,
		WorldHeight: uint(height) * cellSize,
		CellWidth:   cellSize,
		CellHeight:  cellSize,
	})

	//pathfinder := pathing.NewGreedyBFS(pathing.GreedyBFSConfig{
	pathfinder := pathing.NewAStar(pathing.AStarConfig{
		NumCols: uint(grid.NumCols()),
		NumRows: uint(grid.NumRows()),
	})

	const (
		tilePassable = iota
		tileNotPassable
	)

	level := m.missionMap.Level(0)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			cell := level[x][y]

			var cellTile uint8 = tilePassable
			if cell != 0 {
				cellTile = tileNotPassable
			}
			grid.SetCellTile(pathing.GridCoord{X: x, Y: y}, cellTile)
		}
	}

	normalLayer := pathing.MakeGridLayer([4]uint8{
		tilePassable:    1, // passable
		tileNotPassable: 0, // not passable
	})

	return &Pathing{
		grid:       grid,
		pathfinder: pathfinder,
		layer:      normalLayer,
	}
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
	start := pathing.GridCoord{X: int(startPos.X), Y: int(startPos.Y)}
	finish := pathing.GridCoord{X: int(finishPos.X), Y: int(finishPos.Y)}
	path := p.pathfinder.BuildPath(p.grid, start, finish, p.layer)

	steps := make([]*geom.Vector2, 0, path.Steps.Len())
	prevStep := startPos.Copy()
	for path.Steps.HasNext() {
		direction := path.Steps.Next()

		nextStep := prevStep.Copy()

		// FIXME: units snagging on corner due to collision handling, fix collision to let units slide one way or the other until they escape?
		nextStep.X = math.Floor(nextStep.X) + 0.5
		nextStep.Y = math.Floor(nextStep.Y) + 0.5

		// convert directional steps to map coordinates
		switch direction {
		case pathing.DirUp:
			// go down: pathing grid uses reversed Y-coords
			nextStep.Y -= 1
		case pathing.DirDown:
			// go up: pathing grid uses reversed Y-coords
			nextStep.Y += 1
		case pathing.DirLeft:
			nextStep.X -= 1
		case pathing.DirRight:
			nextStep.X += 1
		}

		steps = append(steps, nextStep)
		prevStep = nextStep
	}

	return steps
}
