package model

import (
	"fmt"
	"image/color"
	"math"
	"path"
	"regexp"
	"strings"

	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	"github.com/go-playground/validator/v10"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
	"gopkg.in/yaml.v3"
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
	Scale          float64 `yaml:"scale"`
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
	Image             string       `yaml:"image"`
	Positions         [][2]float64 `yaml:"positions"`
	ZPosition         float64      `yaml:"zPosition"`
	CollisionPxRadius float64      `yaml:"collisionRadius"`
	CollisionPxHeight float64      `yaml:"collisionHeight"`
	HitPoints         float64      `yaml:"hitPoints"`
	Scale             float64      `default:"1.0" yaml:"scale,omitempty"`
	Anchor            SpriteAnchor `yaml:"anchor"`
	Stamp             string       `yaml:"stamp"`
}

type MapSpriteFill struct {
	Image             string     `yaml:"image"`
	Quantity          int        `yaml:"quantity"`
	CollisionPxRadius float64    `yaml:"collisionRadius"`
	CollisionPxHeight float64    `yaml:"collisionHeight"`
	HitPoints         float64    `yaml:"hitPoints"`
	ScaleRange        [2]float64 `yaml:"scaleRange"`
	Rect              [2][2]int  `yaml:"rect"`
}

type MapSpriteStamp struct {
	Id      string      `yaml:"id"`
	Sprites []MapSprite `yaml:"sprites"`
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
}

type MapGeneratePrefabs struct {
	Name      string    `yaml:"name"`
	Levels    [][][]int `yaml:"levels"`
	Positions [][2]int  `yaml:"positions"`
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
			scale := 1.0
			if len(fill.ScaleRange) == 2 {
				// generate random scale value within scale range
				scale = RandFloat64In(fill.ScaleRange[0], fill.ScaleRange[1], rng)
			}

			mapSprite := MapSprite{
				Image:             fill.Image,
				Positions:         [][2]float64{{fX, fY}},
				CollisionPxRadius: fill.CollisionPxRadius,
				CollisionPxHeight: fill.CollisionPxHeight,
				HitPoints:         fill.HitPoints,
				Scale:             scale,
			}
			nSprites = append(nSprites, mapSprite)
		}

	}

	m.Sprites = nSprites
	return nil
}

func (m *Map) generateSpritesFromStamps() error {
	nSprites := make([]MapSprite, len(m.Sprites))

	stampsById := make(map[string]MapSpriteStamp, len(m.SpriteStamps))
	for _, stamp := range m.SpriteStamps {
		stampsById[stamp.Id] = stamp
	}

	for _, sprite := range m.Sprites {
		if sprite.Image != "" {
			nSprites = append(nSprites, sprite)
		}
		if sprite.Stamp != "" {
			if stamp, ok := stampsById[sprite.Stamp]; ok {
				for _, position := range sprite.Positions {
					x, y := position[0], position[1]
					for _, stampSprite := range stamp.Sprites {
						mapPositions := make([][2]float64, len(stampSprite.Positions))
						for i, stampPosition := range stampSprite.Positions {
							mapPositions[i] = [2]float64{x + stampPosition[0], y + stampPosition[1]}
						}
						mapSprite := MapSprite{
							Image:             stampSprite.Image,
							Positions:         mapPositions,
							CollisionPxRadius: stampSprite.CollisionPxRadius,
							CollisionPxHeight: stampSprite.CollisionPxHeight,
							HitPoints:         stampSprite.HitPoints,
							Scale:             stampSprite.Scale,
						}
						nSprites = append(nSprites, mapSprite)
					}
				}
			} else {
				return fmt.Errorf("stamp id is not defined or is misspelled: \"%s\"", sprite.Stamp)
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
			return fmt.Errorf("prefab must have at least one level and one position: %v", prefab.Name)
		}

		if pLevels > m.NumRaycastLevels {
			return fmt.Errorf(
				"prefab cannot have more levels (%v) than numRaycastLevels (%v): %v",
				pLevels, m.NumRaycastLevels, prefab.Name,
			)
		}

		pSizeX, pSizeY := len(prefab.Levels[0]), len(prefab.Levels[0][0])
		if pSizeX == 0 || pSizeY == 0 {
			return fmt.Errorf("prefab level X/Y length must both be greater than zero: %v", prefab.Name)
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

	return nil
}

func (m *Map) GetCollisionLines(clipDistance float64) []*geom.Line {
	if len(m.Levels) == 0 || len(m.Levels[0]) == 0 || len(m.Levels[0][0]) == 0 {
		return []*geom.Line{}
	}

	firstLevel := m.Levels[0]
	lines := make([]*geom.Line, 0, 4*len(firstLevel))

	rectLines := geom.Rect(clipDistance, clipDistance,
		float64(len(firstLevel))-2*clipDistance, float64(len(firstLevel[0]))-2*clipDistance)
	for i := 0; i < len(rectLines); i++ {
		lines = append(lines, &rectLines[i])
	}

	for x, row := range firstLevel {
		for y, value := range row {
			if value > 0 {
				rectLines = geom.Rect(float64(x)-clipDistance, float64(y)-clipDistance,
					1.0+(2*clipDistance), 1.0+(2*clipDistance))
				for i := 0; i < len(rectLines); i++ {
					lines = append(lines, &rectLines[i])
				}
			}
		}
	}

	return lines
}

func (m *Map) IsWallAt(levelNum, x, y int) bool {
	level := m.Level(levelNum)
	return level[x][y] > 0
}

func (m *Map) GetMapTexture(texIndex int) MapTexture {
	return m.Textures[texIndex]
}
