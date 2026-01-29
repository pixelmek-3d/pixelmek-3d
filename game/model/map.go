package model

import (
	"fmt"
	"image/color"
	"math"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	"github.com/go-playground/validator/v10"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

type CardinalDirection byte

const (
	NORTH CardinalDirection = iota
	EAST
	SOUTH
	WEST
	NOWHERE
)

func (cd CardinalDirection) Next() CardinalDirection {
	next := cd + 1
	if next == NOWHERE {
		next = NORTH
	}
	return next
}

func (cd CardinalDirection) Prev() CardinalDirection {
	next := cd - 1
	if cd == NORTH {
		next = WEST
	}
	return next
}

func (cd CardinalDirection) Opposite() CardinalDirection {
	switch cd {
	case NORTH:
		return SOUTH
	case EAST:
		return WEST
	case SOUTH:
		return NORTH
	case WEST:
		return EAST
	}
	return NOWHERE
}

func (cd CardinalDirection) String() string {
	switch cd {
	case NORTH:
		return "N"
	case EAST:
		return "E"
	case SOUTH:
		return "S"
	case WEST:
		return "W"
	}
	return "?"
}

type Map struct {
	Name             string             `yaml:"name" validate:"required"`
	DropZone         DropZone           `yaml:"dropZone" validate:"required"`
	SpawnPoints      [][2]float64       `yaml:"spawnPoints"`
	NumRaycastLevels int                `yaml:"numRaycastLevels"`
	Levels           [][][]int          `yaml:"levels"`
	GenerateLevels   MapGenerateLevels  `yaml:"generateLevels"`
	Lighting         MapLighting        `yaml:"lighting" validate:"required"`
	Textures         map[int]MapTexture `yaml:"textures" validate:"gt=0"`
	FloorBox         MapTexture         `yaml:"floorBox" validate:"required"`
	SkyBox           MapTexture         `yaml:"skyBox" validate:"required"`
	Flooring         MapFlooring        `yaml:"flooring" validate:"required"`
	Clutter          []MapClutter       `yaml:"clutter"`
	Sprites          []MapSprite        `yaml:"sprites"`
	SpriteFill       []MapSpriteFill    `yaml:"spriteFill"`
	SpriteStamps     []MapSpriteStamp   `yaml:"spriteStamps"`
	Seed             int64              `yaml:"seed"`
	MusicPath        string             `yaml:"music"`

	// Sprite ID mapping is initialized when map data is being loaded
	spritesByID map[string]MapSprite `yaml:"-"`
}

type MapTexture struct {
	Image string `yaml:"image"`
	SideX string `yaml:"sideX"`
	SideY string `yaml:"sideY"`
}

func (m MapTexture) GetImage(side int) string {
	switch {
	case side == 0 && m.SideY != "":
		return m.SideY
	case side == 1 && m.SideX != "":
		return m.SideX
	default:
		return m.Image
	}
}

type MapFlooring struct {
	Default string            `yaml:"default"`
	Pathing []MapFloorPathing `yaml:"pathing"`
}

type MapFloorPathing struct {
	Image string      `yaml:"image"`
	Rects [][2][2]int `yaml:"rects"`
	Lines [][][2]int  `yaml:"lines"`
}

type MapClutter struct {
	Image          string  `yaml:"image"`
	FloorPathMatch *RegExp `yaml:"floorPathMatch"`
	Frequency      float64 `yaml:"frequency"`
	Height         float64 `yaml:"height"`
}

type SpriteAnchor struct {
	raycaster.SpriteAnchor
}

// Unmarshals into raycaster.SpriteAnchor
func (r *SpriteAnchor) UnmarshalText(b []byte) error {
	str := strings.Trim(string(b), `"`)

	top, center, bottom := "top", "center", "bottom"

	switch str {
	case top:
		r.SpriteAnchor = raycaster.AnchorTop
	case center:
		r.SpriteAnchor = raycaster.AnchorCenter
	case bottom, "":
		r.SpriteAnchor = raycaster.AnchorBottom
	default:
		return fmt.Errorf("unknown anchor value '%s', must be one of: [%s, %s, %s]", str, top, center, bottom)
	}

	return nil
}

type RegExp struct {
	*regexp.Regexp
}

// Unmarshals into compiled regex
func (r *RegExp) UnmarshalText(b []byte) error {
	regex, err := regexp.Compile(string(b))
	if err != nil {
		return err
	}

	r.Regexp = regex
	return nil
}

type MapSprite struct {
	ID                string       `yaml:"id"`
	Image             string       `yaml:"image"`
	Positions         [][2]float64 `yaml:"positions"`
	ZPosition         float64      `yaml:"zPosition"`
	CollisionPxRadius float64      `yaml:"collisionRadiusPx"`
	CollisionPxHeight float64      `yaml:"collisionHeightPx"`
	HitPoints         float64      `yaml:"hitPoints"`
	Height            float64      `yaml:"height"`
	Anchor            SpriteAnchor `yaml:"anchor"`
	Stamp             string       `yaml:"stamp"`
}

type MapSpriteFill struct {
	SpriteID    string     `yaml:"sprite"`
	Quantity    int        `yaml:"quantity"`
	HeightRange [2]float64 `yaml:"heightRange"`
	Rect        [2][2]int  `yaml:"rect"`
}

type MapSpriteStamp struct {
	Positions [][2]float64        `yaml:"positions"`
	Sprites   []MapSpriteStampRef `yaml:"sprites"`
}

type MapSpriteStampRef struct {
	SpriteID  string       `yaml:"sprite"`
	Height    float64      `yaml:"height"`
	Positions [][2]float64 `yaml:"positions"`
}

type MapLighting struct {
	Falloff      float64  `yaml:"falloff"`
	Illumination float64  `yaml:"illumination"`
	MinLightRGB  [3]uint8 `yaml:"minLightRGB"`
	MaxLightRGB  [3]uint8 `yaml:"maxLightRGB"`
}

func (m MapLighting) LightRGB() (*color.NRGBA, *color.NRGBA) {
	min := &color.NRGBA{
		R: m.MinLightRGB[0], G: m.MinLightRGB[1], B: m.MinLightRGB[2], A: 255,
	}
	max := &color.NRGBA{
		R: m.MaxLightRGB[0], G: m.MaxLightRGB[1], B: m.MaxLightRGB[2], A: 255,
	}
	return min, max
}

type MapGenerateLevels struct {
	MapSize      [2]int               `yaml:"mapSize"`
	BoundaryWall MapTexture           `yaml:"boundaryWall"`
	Prefabs      []MapGeneratePrefabs `yaml:"prefabs"`
	Walls        []MapGenerateWalls   `yaml:"walls"`
}

func (m MapGenerateLevels) HasBoundaryWall() bool {
	return len(m.BoundaryWall.Image) > 0
}

type MapGeneratePrefabs struct {
	Name      string    `yaml:"name"`
	Layers    [][][]int `yaml:"layers"`
	Positions [][2]int  `yaml:"positions"`
}

type MapGenerateWalls struct {
	Texture int        `yaml:"texture"`
	Height  int        `yaml:"height"`
	Lines   [][][2]int `yaml:"lines"`
}

func (m *Map) Size() (width int, height int) {
	if len(m.Levels) == 0 || len(m.Levels[0]) == 0 {
		return 0, 0
	}
	width, height = len(m.Levels[0]), len(m.Levels[0][0])
	return
}

func (m *Map) NumLevels() int {
	return m.NumRaycastLevels
}

func (m *Map) Level(levelNum int) [][]int {
	lenLevels := len(m.Levels)
	if levelNum < lenLevels {
		return m.Levels[levelNum]
	} else {
		return m.Levels[lenLevels-1] // if above highest level index just keep extending last one up
	}
}

func LoadMap(mapFile string) (*Map, error) {
	if filepath.Ext(mapFile) == "" {
		mapFile += YAMLExtension
	}

	v := validator.New()
	mapPath := path.Join("maps", mapFile)

	mapYaml, err := resources.ReadFile(mapPath)
	if err != nil {
		return nil, err
	}

	m := &Map{}
	err = yaml.Unmarshal(mapYaml, m)
	if err != nil {
		return nil, err
	}

	err = v.Struct(m)
	if err != nil {
		return nil, fmt.Errorf("[%s] %s", mapPath, err.Error())
	}

	if len(m.Textures) == 0 {
		return m, fmt.Errorf("one or more entry in textures is required")
	}

	if len(m.Levels) == 0 && len(m.GenerateLevels.MapSize) != 2 {
		return m, fmt.Errorf("levels or generateLevels is required")
	}

	if len(m.Levels) > 0 && len(m.GenerateLevels.MapSize) == 2 {
		return m, fmt.Errorf("use of levels or generateLevels is mutually exclusive")
	}

	// generate levels array
	if len(m.GenerateLevels.MapSize) == 2 {
		err := m.generateMapLevels()
		if err != nil {
			return m, err
		}
	}

	if m.NumRaycastLevels == 0 {
		// default to number of levels provided in levels array
		m.NumRaycastLevels = len(m.Levels)
	}

	// map sprites by ID for use in sprite fill/stamps
	m.spritesByID = make(map[string]MapSprite, len(m.Sprites))
	for _, mSprite := range m.Sprites {
		if len(mSprite.ID) == 0 {
			continue
		}
		if _, exists := m.spritesByID[mSprite.ID]; exists {
			log.Errorf("sprite with same ID is defined more than once: %s", mSprite.ID)
			continue
		}
		m.spritesByID[mSprite.ID] = mSprite
	}

	// generate additional sprites using sprite fill
	if len(m.SpriteFill) > 0 {
		err := m.generateFillerSprites()
		if err != nil {
			return m, err
		}
	}

	// generate additional sprites using sprite stamps
	if len(m.SpriteStamps) > 0 {
		err := m.generateSpritesFromStamps()
		if err != nil {
			return m, err
		}
	}

	return m, nil
}

func ListMapFilenames() ([]string, error) {
	mapFilenames := make([]string, 0, 64)
	mapsPath := "maps"
	mapsFiles, err := resources.ReadDir(mapsPath)
	if err != nil {
		return mapFilenames, err
	}

	for _, f := range mapsFiles {
		if f.IsDir() {
			// only folder with maps files expected
			continue
		}
		mapFilenames = append(mapFilenames, f.Name())
	}

	return mapFilenames, nil
}

func (m *Map) generateFillerSprites() error {
	nSprites := make([]MapSprite, len(m.Sprites))
	copier.Copy(&nSprites, &m.Sprites)

	rng := NewRNG()

	for n, fill := range m.SpriteFill {
		rng.Seed(m.Seed + int64(n))

		x0, y0 := float64(fill.Rect[0][0]), float64(fill.Rect[0][1])
		x1, y1 := float64(fill.Rect[1][0]), float64(fill.Rect[1][1])

		for i := 0; i < fill.Quantity; i++ {
			fX, fY := rng.RandFloat64In(x0, x1), rng.RandFloat64In(y0, y1)

			var height float64
			if len(fill.HeightRange) == 2 {
				// generate random height value within height range
				height = rng.RandFloat64In(fill.HeightRange[0], fill.HeightRange[1])
			}

			mapSprite, ok := m.spritesByID[fill.SpriteID]
			if !ok {
				log.Errorf("fill sprite not found with ID: %s", fill.SpriteID)
				continue
			}

			mapSprite.Positions = [][2]float64{{fX, fY}}
			mapSprite.Height = height

			nSprites = append(nSprites, mapSprite)
		}

	}

	m.Sprites = nSprites
	return nil
}

func (m *Map) generateSpritesFromStamps() error {
	nSprites := make([]MapSprite, len(m.Sprites))
	copier.Copy(&nSprites, &m.Sprites)

	for _, stamp := range m.SpriteStamps {
		for _, position := range stamp.Positions {
			x, y := position[0], position[1]
			for _, stampSprite := range stamp.Sprites {
				mapSprite, ok := m.spritesByID[stampSprite.SpriteID]
				if !ok {
					log.Errorf("stamp sprite not found with ID: %s", stampSprite.SpriteID)
					continue
				}

				mapPositions := make([][2]float64, len(stampSprite.Positions))
				for i, stampPosition := range stampSprite.Positions {
					mapPositions[i] = [2]float64{x + stampPosition[0], y + stampPosition[1]}
				}
				mapSprite.Positions = mapPositions
				mapSprite.Height = stampSprite.Height

				nSprites = append(nSprites, mapSprite)
			}
		}
	}

	m.Sprites = nSprites
	return nil
}

func (m *Map) generateMapLevels() error {
	gen := m.GenerateLevels
	mapSizeX, mapSizeY := gen.MapSize[0], gen.MapSize[1]

	if mapSizeX <= 0 || mapSizeY <= 0 {
		return fmt.Errorf("map X/Y size must both be greater than zero")
	}

	// initialize map level slices
	m.Levels = make([][][]int, m.NumRaycastLevels)
	for i := range m.NumRaycastLevels {
		m.Levels[i] = make([][]int, mapSizeX)
		for x := range mapSizeX {
			m.Levels[i][x] = make([]int, mapSizeY)
		}
	}

	// if provided, create boundary wall
	if gen.HasBoundaryWall() {
		// at this time boundary walls only supported on first elevation level
		level := m.Levels[0]

		// store boundary wall map texture as its own index (for now just very large int not likely to be in use)
		// TODO: create a function to generate unused index to make sure it's not in use?
		boundaryTex := math.MaxInt16
		m.Textures[boundaryTex] = gen.BoundaryWall

		for x := range mapSizeX {
			for y := range mapSizeY {
				if x == 0 || y == 0 || x == mapSizeX-1 || y == mapSizeY-1 {
					level[x][y] = boundaryTex
				}
			}
		}
	}

	// populate "prefab" structures
	for _, prefab := range gen.Prefabs {
		pLayers := len(prefab.Layers)
		if pLayers == 0 || len(prefab.Positions) == 0 {
			return fmt.Errorf("prefab must have at least one layer and one position: %s", prefab.Name)
		}

		if pLayers > m.NumRaycastLevels {
			return fmt.Errorf(
				"prefab cannot have more layers (%d) than numRaycastLevels (%d): %s",
				pLayers, m.NumRaycastLevels, prefab.Name,
			)
		}

		if len(prefab.Layers[0]) == 0 || len(prefab.Layers[0][0]) == 0 {
			return fmt.Errorf("prefab layer Y/X length must both be greater than zero: %s", prefab.Name)
		}

		for _, pos := range prefab.Positions {
			posX, posY := pos[0], pos[1]

			for i := range pLayers {
				// map level layer in revese so the layers appear on top of each other instead of under
				iReverse := pLayers - i - 1
				for y := range len(prefab.Layers[i]) {
					// map Y-layer in reverse so the layers do not appear mirrored
					yReverse := len(prefab.Layers[i]) - y - 1
					for x := range len(prefab.Layers[i][y]) {
						if x+posX >= mapSizeX || yReverse+posY >= mapSizeY {
							continue
						}
						m.Levels[iReverse][x+posX][yReverse+posY] = prefab.Layers[i][y][x]
					}
				}
			}
		}
	}

	// generate walls from lines
	for i, wall := range gen.Walls {
		tex := wall.Texture
		if tex < 0 {
			return fmt.Errorf("generated wall at index [%d] must have texture index of at least zero, found value: %d", i, tex)
		}

		height := wall.Height
		if height < 1 {
			return fmt.Errorf("generated wall at index [%d] must have level height of at least one, found value: %d", i, height)
		}
		if height > m.NumRaycastLevels {
			return fmt.Errorf(
				"generated wall at index [%d] cannot have more levels (%d) than numRaycastLevels (%d)",
				i, height, m.NumRaycastLevels,
			)
		}

		// create line segment paths
		for _, segments := range wall.Lines {
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
						for levelIndex := range height {
							level := m.Levels[levelIndex]
							level[int(nLine.X2)][int(nLine.Y2)] = tex
						}
					}
				}

				prevPoint = point
			}
		}
	}

	return nil
}

type wallLineGenerator struct {
	m     *Map
	cells [][]*cellBorder
}

type cellBorder struct {
	dirLines map[CardinalDirection]*wallLine
	visited  bool
}

type wallLine struct {
	geom.Line
	dir       CardinalDirection
	cancelled bool
	visited   bool
}

type wallGroup struct {
	lines []*wallLine
}

// cancelled returns true if all lines in the cell are cancelled out
func (cb *cellBorder) cancelled() bool {
	for _, l := range cb.dirLines {
		if !l.cancelled {
			return false
		}
	}
	return true
}

func (l *wallLine) String() string {
	return fmt.Sprintf("{%0.3f,%0.3f->%0.3f,%0.3f@%v}", l.X1, l.Y1, l.X2, l.Y2, l.dir)
}

func newWallLineGenerator(m *Map) *wallLineGenerator {
	w, h := m.Size()
	cells := make([][]*cellBorder, w)
	for i := range cells {
		cells[i] = make([]*cellBorder, h)
	}
	gen := &wallLineGenerator{
		m:     m,
		cells: cells,
	}
	gen.initializeCellWalls()
	return gen
}

func newCellBorders(x, y int) *cellBorder {
	return &cellBorder{
		dirLines: map[CardinalDirection]*wallLine{
			NORTH: createWallLine(x, y, NORTH),
			EAST:  createWallLine(x, y, EAST),
			SOUTH: createWallLine(x, y, SOUTH),
			WEST:  createWallLine(x, y, WEST),
		},
	}
}

func (g *wallLineGenerator) cellInDirection(x, y int, direction CardinalDirection) (*cellBorder, int, int) {
	i, j := x, y
	switch direction {
	case NORTH:
		j++
	case EAST:
		i++
	case SOUTH:
		j--
	case WEST:
		i--
	}

	w, h := g.m.Size()
	if i < 0 || j < 0 || i >= w || j >= h {
		return nil, x, y
	}

	return g.cells[i][j], i, j
}

func (g *wallLineGenerator) cellInDiagonalFromDirection(x, y int, direction CardinalDirection) (*cellBorder, int, int) {
	i, j := x, y
	switch direction {
	case NORTH:
		i--
		j++
	case EAST:
		i++
		j++
	case SOUTH:
		i++
		j--
	case WEST:
		i--
		j--
	}

	w, h := g.m.Size()
	if i < 0 || j < 0 || i >= w || j >= h {
		return nil, x, y
	}

	return g.cells[i][j], i, j
}

func (g *wallLineGenerator) initializeCellWalls() {
	// initialize border line segments for each wall cell
	w, h := g.m.Size()
	level := g.m.Levels[0]
	wallSet := make(map[geom.Line]*wallLine, w*h/4)

	for x := range w {
		for y := range h {
			if g.m.GenerateLevels.HasBoundaryWall() {
				// skip creating cells from boundary wall
				if x == 0 || y == 0 || x == w-1 || y == h-1 {
					continue
				}
			}
			if level[x][y] == 0 {
				continue
			}
			// for each wall cell, create directional line for each of its 4 borders
			cb := newCellBorders(x, y)
			g.cells[x][y] = cb

			// for each border, check if an opposite direction of the same segment exists to cancel each other out
			for _, dLine := range cb.dirLines {
				oppLine := LineOpposite(dLine.Line)
				if existingLine, found := wallSet[oppLine]; found {
					// cancel each other out
					dLine.cancelled = true
					existingLine.cancelled = true
				} else {
					wallSet[dLine.Line] = dLine
				}
			}
		}
	}
}

func (g *wallLineGenerator) generateWallGroups() []*wallGroup {
	// walk connected cell lines to form connected groups of wall line segments
	w, h := g.m.Size()

	wallGroups := make([]*wallGroup, 0, w*h/4)

	// walking in X direction before Y to start from the west and go clockwise
	for y := range h {
		for x := range w {
			cell := g.cells[x][y]
			if cell == nil || cell.visited || cell.cancelled() {
				continue
			}

			// start a new wallgroup of connecting lines
			wg := &wallGroup{lines: make([]*wallLine, 0, 4)}
			wallGroups = append(wallGroups, wg)

			// go clockwise starting North since coming in from the West
			lineDir := NORTH
			ogLine := cell.dirLines[lineDir]
			ogPoint := geom.Vector2{X: ogLine.X1, Y: ogLine.Y1}
			//fmt.Printf("> ogPoint: %v\n", ogPoint)

			i, j := x, y
			startDir := lineDir
			for {
				cell.visited = true
				line := cell.dirLines[lineDir]
				if line == nil || line.cancelled || line.visited {
					lineDir = lineDir.Next()
					if lineDir == startDir {
						log.Errorf("[%d, %d] unable to escape cell lines - breaking out of infinite loop", i, j)
						break
					}
					continue
				}

				startDir = lineDir
				wg.addLine(line)
				//fmt.Printf("+ add %v\n", line)

				// check adjacent cell in current direction if it continues in same direction
				peekCell, peekX, peekY := g.cellInDirection(i, j, lineDir)
				if peekCell != nil {
					peekLine := peekCell.dirLines[lineDir]
					if peekLine != nil && !peekLine.cancelled && !peekLine.visited {
						// move to the new cell with wall line in the same direction
						//fmt.Printf("  - peekLine: %v\n", peekLine)
						cell = peekCell
						i, j = peekX, peekY
						continue
					}
				}

				// check for lines that can continue in diagnonal instead of adjacent cells
				diagCell, diagX, diagY := g.cellInDiagonalFromDirection(i, j, lineDir)
				if diagCell != nil {
					diagDir := lineDir.Prev()
					diagLine := diagCell.dirLines[diagDir]
					if diagLine != nil && !diagLine.cancelled && !diagLine.visited {
						//fmt.Printf("  > diagLine: %v\n", diagLine)
						cell = diagCell
						lineDir = diagDir
						i, j = diagX, diagY
						continue
					}
				}

				// break when back to origin point of wallgroup
				if line.X2 == ogPoint.X && line.Y2 == ogPoint.Y {
					//fmt.Printf("< back to ogPoint: %v\n", ogPoint)
					break
				}
			}
		}
	}
	return wallGroups
}

func (wg *wallGroup) addLine(l *wallLine) {
	l.visited = true
	wg.lines = append(wg.lines, l)
}

func (wg *wallGroup) generateWallLines() []*geom.Line {
	lines := make([]*geom.Line, 0, len(wg.lines))
	if len(wg.lines) == 0 {
		return lines
	}

	line := wg.lines[0]
	lineDir := line.dir
	prevDir := lineDir
	lines = append(lines, &line.Line)

	for i := 1; i < len(wg.lines); i++ {
		wl := wg.lines[i]
		lineDir = wl.dir

		if prevDir == lineDir {
			// update current wall line in same direction as before
			line.X2, line.Y2 = wl.Line.X2, wl.Line.Y2
		} else {
			// start a new wall line
			line = wl
			lines = append(lines, &line.Line)
		}
		prevDir = lineDir
	}
	return lines
}

func (m *Map) GenerateWallCollisionLines(clipDistance float64) []*geom.Line {
	w, h := m.Size()
	if w == 0 || h == 0 {
		return []*geom.Line{}
	}

	level := m.Levels[0]
	lines := make([]*geom.Line, 0, 4*len(level))

	if m.GenerateLevels.HasBoundaryWall() {
		// create collision lines around inner border boundary wall
		rectLines := geom.Rect(1+clipDistance, 1+clipDistance,
			float64(w-2)-2*clipDistance, float64(h-2)-2*clipDistance)
		for i := range rectLines {
			lines = append(lines, &rectLines[i])
		}
	}

	// create collision lines around outer map border
	rectLines := geom.Rect(clipDistance, clipDistance,
		float64(w)-2*clipDistance, float64(h)-2*clipDistance)
	for i := range rectLines {
		lines = append(lines, &rectLines[i])
	}

	// Phase 1 - Create 4 border lines per cell with cardinal direction of clockwise movement
	//           to trace outlines of contiguous wall segments
	//         - Keep track of same line segments and cancel out those with opposite cardinal direction
	gen := newWallLineGenerator(m)

	// Phase 2 - Go over each cell, walking the wall line segments to other connected cells
	//           to put contiguous line segments in groups
	wallGroups := gen.generateWallGroups()

	// Phase 3 - Connect contiguous lines in the same direction as single line instead of segments
	for _, wg := range wallGroups {
		lines = append(lines, wg.generateWallLines()...)
	}

	return lines
}

// createWallLine starts a new outer line for a cell based on the given direction
func createWallLine(x, y int, direction CardinalDirection) *wallLine {
	x1, y1 := float64(x), float64(y)
	x2, y2 := float64(x), float64(y)

	// TODO: account for clipDistance
	switch direction {
	case NORTH:
		// start from bottom left of cell and go up
		y2++
	case EAST:
		// start from top left of cell and go right
		y1++
		x2++
		y2++
	case SOUTH:
		// start from top right of cell and go down
		x1++
		y1++
		x2++
	case WEST:
		// start from bottom right of cell and go left
		x1++
	default:
		return nil
	}

	return &wallLine{
		Line: geom.Line{X1: x1, Y1: y1, X2: x2, Y2: y2},
		dir:  direction,
	}
}

func (m *Map) IsWallAt(levelNum, x, y int) bool {
	level := m.Level(levelNum)
	return level[x][y] > 0
}

func (m *Map) GetMapTexture(texIndex int) MapTexture {
	return m.Textures[texIndex]
}
