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

	mapset "github.com/deckarep/golang-set/v2"
	log "github.com/sirupsen/logrus"
)

type Map struct {
	NumRaycastLevels int                `yaml:"numRaycastLevels"`
	Levels           [][][]int          `yaml:"levels"`
	GenerateLevels   MapGenerateLevels  `yaml:"generateLevels"`
	Lighting         MapLighting        `yaml:"lighting"`
	Textures         map[int]MapTexture `yaml:"textures"`
	FloorBox         MapTexture         `yaml:"floorBox"`
	SkyBox           MapTexture         `yaml:"skyBox"`
	Flooring         MapFlooring        `yaml:"flooring"`
	Clutter          []MapClutter       `yaml:"clutter"`
	Sprites          []MapSprite        `yaml:"sprites"`
	SpriteFill       []MapSpriteFill    `yaml:"spriteFill"`
	SpriteStamps     []MapSpriteStamp   `yaml:"spriteStamps"`
	Seed             int64              `yaml:"seed"`

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
	CollisionPxRadius float64      `yaml:"collisionRadius"`
	CollisionPxHeight float64      `yaml:"collisionHeight"`
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

type MapGeneratePrefabs struct {
	Name      string    `yaml:"name"`
	Levels    [][][]int `yaml:"levels"`
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

func (m *Map) generateFillerSprites() error {
	nSprites := make([]MapSprite, len(m.Sprites))
	copier.Copy(&nSprites, &m.Sprites)

	rng := NewRNG()

	for n, fill := range m.SpriteFill {
		rng.Seed(m.Seed + int64(n))

		x0, y0 := float64(fill.Rect[0][0]), float64(fill.Rect[0][1])
		x1, y1 := float64(fill.Rect[1][0]), float64(fill.Rect[1][1])

		for i := 0; i < fill.Quantity; i++ {
			fX, fY := RandFloat64In(x0, x1, rng), RandFloat64In(y0, y1, rng)

			var height float64
			if len(fill.HeightRange) == 2 {
				// generate random height value within height range
				height = RandFloat64In(fill.HeightRange[0], fill.HeightRange[1], rng)
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
	for i := 0; i < m.NumRaycastLevels; i++ {
		m.Levels[i] = make([][]int, mapSizeX)
		for x := 0; x < mapSizeX; x++ {
			m.Levels[i][x] = make([]int, mapSizeY)
		}
	}

	// if provided, create boundary wall
	if len(gen.BoundaryWall.Image) > 0 {
		// at this time boundary walls only supported on first elevation level
		level := m.Levels[0]

		// store boundary wall map texture as its own index (for now just very large int not likely to be in use)
		// TODO: create a function to generate unused index to make sure it's not in use?
		boundaryTex := math.MaxInt16
		m.Textures[boundaryTex] = gen.BoundaryWall

		for x := 0; x < mapSizeX; x++ {
			for y := 0; y < mapSizeY; y++ {
				if x == 0 || y == 0 || x == mapSizeX-1 || y == mapSizeY-1 {
					level[x][y] = boundaryTex
				}
			}
		}
	}

	// populate "prefab" structures
	for _, prefab := range gen.Prefabs {
		pLevels := len(prefab.Levels)
		if pLevels == 0 || len(prefab.Positions) == 0 {
			return fmt.Errorf("prefab must have at least one level and one position: %s", prefab.Name)
		}

		if pLevels > m.NumRaycastLevels {
			return fmt.Errorf(
				"prefab cannot have more levels (%d) than numRaycastLevels (%d): %s",
				pLevels, m.NumRaycastLevels, prefab.Name,
			)
		}

		pSizeX, pSizeY := len(prefab.Levels[0]), len(prefab.Levels[0][0])
		if pSizeX == 0 || pSizeY == 0 {
			return fmt.Errorf("prefab level X/Y length must both be greater than zero: %s", prefab.Name)
		}

		for _, pos := range prefab.Positions {
			posX, posY := pos[0], pos[1]

			for i := 0; i < pLevels; i++ {
				for x := 0; x < pSizeX; x++ {
					for y := 0; y < pSizeY; y++ {
						if x+posX >= mapSizeX || y+posY >= mapSizeY {
							continue
						}
						m.Levels[i][x+posX][y+posY] = prefab.Levels[i][x][y]
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
						for levelIndex := 0; levelIndex < height; levelIndex++ {
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

type KeyValuePair[K any, V any] struct {
	Key   K `json:"Key" form:"Key"`
	Value V `json:"Value" form:"Value"`
}

func (m *Map) GetCollisionLines(clipDistance float64) []*geom.Line {
	if len(m.Levels) == 0 || len(m.Levels[0]) == 0 || len(m.Levels[0][0]) == 0 {
		return []*geom.Line{}
	}

	firstLevel := m.Levels[0]
	lines := make([]*geom.Line, 0, 4*len(firstLevel))

	// border lines for edge of map
	rectLines := geom.Rect(clipDistance, clipDistance,
		float64(len(firstLevel))-2*clipDistance, float64(len(firstLevel[0]))-2*clipDistance)
	for _, li := range rectLines {
		lines = append(lines, &li)
	}

	// create list of coordinates for the corners of each wall cell
	// based on https://stackoverflow.com/a/73233163/854696
	wallCoords := make([][4]*geom.Vector2, 0, 4*len(firstLevel))

	for x, row := range firstLevel {
		for y, value := range row {
			if value > 0 {
				fX, fY := float64(x), float64(y)
				cellCoords := [4]*geom.Vector2{
					{X: fX, Y: fY},
					{X: fX + 1, Y: fY},
					{X: fX + 1, Y: fY + 1},
					{X: fX, Y: fY + 1},
				}
				wallCoords = append(wallCoords, cellCoords)
			}
		}
	}

	// compute edges of contiguous wall cells
	edges := mapset.NewThreadUnsafeSet[KeyValuePair[geom.Vector2, geom.Vector2]]()
	var firstEdge *KeyValuePair[geom.Vector2, geom.Vector2]
	for _, coords := range wallCoords {
		// Iterate over the coordinates to compute the edges
		// Using for loop to skip already processed edges
		for i, src := range coords {
			for j, dest := range coords {
				if i == j {
					continue
				}

				// Check the distance between them to filter out the diagonal edges
				coordDistance := geom.Distance(src.X, src.Y, dest.X, dest.Y)
				if !(math.Abs(coordDistance-1.0) < 0.001) {
					continue
				}

				edge := KeyValuePair[geom.Vector2, geom.Vector2]{*src, *dest}
				if firstEdge == nil {
					firstEdge = &edge
				}

				if edges.Contains(edge) {
					// If the edge already exists in the set,
					// it means its not part of the border
					edges.Remove(edge)
				} else {
					edges.Add(edge)
				}
			}
		}
	}

	// TODO: create ordered list of lines from the edge points
	coordsList := make([]geom.Vector2, 0, len(wallCoords))

	// Make a copy of the edges so we can remove items from it
	// without destroying the original collection
	cpEdges := edges.Clone()

	// Add the first pair before starting the loop
	previousEdge := KeyValuePair[geom.Vector2, geom.Vector2]{firstEdge.Key, firstEdge.Value}

	coordsList = append(coordsList, previousEdge.Key, previousEdge.Value)

	var currentEdge KeyValuePair[geom.Vector2, geom.Vector2]

	// While there is an edge that follows the previous one

	//  while (!(currentEdge = copy.FirstOrDefault(pair => pair.Key == previousEdge.Value))
	//  	.Equals(default(KeyValuePair<Vector3, Vector3>)))
	// {
	defaultEdge := KeyValuePair[geom.Vector2, geom.Vector2]{}
	currentEdge = setFirstOrDefaultWithKey(cpEdges, previousEdge.Value)
	for currentEdge != defaultEdge {
		// Our graph is not oriented but we want to ignores edges
		// that go back from where we went
		if currentEdge.Key.Equals(&previousEdge.Key) && currentEdge.Value.Equals(&previousEdge.Value) {
			cpEdges.Remove(currentEdge)
		}

		// Add the vertex to the list and continue
		coordsList = append(coordsList, currentEdge.Value)
		previousEdge = currentEdge

		// Remove traversed nodes
		cpEdges.Remove(currentEdge)

		currentEdge = setFirstOrDefaultWithKey(cpEdges, previousEdge.Value)
	}

	return lines
}

func setFirstOrDefaultWithKey(set mapset.Set[KeyValuePair[geom.Vector2, geom.Vector2]], key geom.Vector2) KeyValuePair[geom.Vector2, geom.Vector2] {
	for kv := range set.Iter() {
		if kv.Key.Equals(&key) {
			return kv
		}
	}
	return KeyValuePair[geom.Vector2, geom.Vector2]{}
}

func (m *Map) IsWallAt(levelNum, x, y int) bool {
	level := m.Level(levelNum)
	return level[x][y] > 0
}

func (m *Map) GetMapTexture(texIndex int) MapTexture {
	return m.Textures[texIndex]
}
