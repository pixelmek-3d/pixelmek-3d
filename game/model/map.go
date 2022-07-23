package model

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"

	"github.com/harbdog/raycaster-go/geom"
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
	FloorPathMatch string  `yaml:"floorPathMatch"`
	Frequency      float64 `yaml:"frequency"`
	Scale          float64 `yaml:"scale"`
}

type MapSprite struct {
	Image     string       `yaml:"image"`
	Positions [][2]float64 `yaml:"positions"`
}

type MapLighting struct {
	Falloff      float64  `yaml:"falloff"`
	Illumination float64  `yaml:"illumination"`
	MinLightRGB  [3]uint8 `yaml:"minLightRGB"`
	MaxLightRGB  [3]uint8 `yaml:"maxLightRGB"`
}

func (m MapLighting) LightRGB() (min, max color.NRGBA) {
	min.R, min.G, min.B = m.MinLightRGB[0], m.MinLightRGB[1], m.MinLightRGB[2]
	max.R, max.G, max.B = m.MaxLightRGB[0], m.MaxLightRGB[1], m.MaxLightRGB[2]
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
	mapsPath := filepath.Join("game", "resources", "maps", mapFile)

	mapsYaml, err := ioutil.ReadFile(mapsPath)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	m := &Map{}
	err = yaml.Unmarshal(mapsYaml, m)
	if err != nil {
		log.Fatal(err)
		return nil, err
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

	return m, nil
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

func (m *Map) GetCollisionLines(clipDistance float64) []geom.Line {
	if len(m.Levels) == 0 || len(m.Levels[0]) == 0 || len(m.Levels[0][0]) == 0 {
		return []geom.Line{}
	}

	firstLevel := m.Levels[0]
	lines := geom.Rect(clipDistance, clipDistance,
		float64(len(firstLevel))-2*clipDistance, float64(len(firstLevel[0]))-2*clipDistance)

	for x, row := range firstLevel {
		for y, value := range row {
			if value > 0 {
				lines = append(lines, geom.Rect(float64(x)-clipDistance, float64(y)-clipDistance,
					1.0+(2*clipDistance), 1.0+(2*clipDistance))...)
			}
		}
	}

	return lines
}

func (m *Map) GetMapTexture(texIndex int) MapTexture {
	return m.Textures[texIndex]
}
