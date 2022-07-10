package model

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"
	"strconv"

	"github.com/harbdog/raycaster-go/geom"
	"gopkg.in/yaml.v3"
)

type Map struct {
	NumRaycastLevels int                   `yaml:"numRaycastLevels"`
	Levels           [][][]int             `yaml:"levels"`
	GenerateLevels   MapGenerateLevels     `yaml:"generateLevels"`
	Lighting         MapLighting           `yaml:"lighting"`
	Textures         map[string]MapTexture `yaml:"textures"`
	FloorBox         MapTexture            `yaml:"floorBox"`
	SkyBox           MapTexture            `yaml:"skyBox"`
	Sprites          []MapSprite           `yaml:"sprites"`
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

type MapSprite struct {
	Image    string     `yaml:"image"`
	Position [2]float64 `yaml:"position"`
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
	Name      string       `yaml:"name"`
	Levels    [][][]int    `yaml:"levels"`
	Positions [][2]float64 `yaml:"positions"`
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
	sizeX, sizeY := gen.MapSize[0], gen.MapSize[1]

	if sizeX <= 0 || sizeY <= 0 {
		return fmt.Errorf("map X/Y size must both be greater than zero")
	}

	// initialize map level slices
	m.Levels = make([][][]int, m.NumRaycastLevels)
	for i := 0; i < m.NumRaycastLevels; i++ {
		m.Levels[i] = make([][]int, sizeX)
		for x := 0; x < sizeX; x++ {
			m.Levels[i][x] = make([]int, sizeY)
		}
	}

	// if provided, create boundary wall
	if len(gen.BoundaryWall.Image) > 0 {
		// at this time boundary walls only supported on first elevation level
		level := m.Levels[0]

		// store boundary wall map texture as its own index (for now just very large int not likely to be in use)
		// TODO: create a function to generate unused index to make sure its not in use
		boundaryTex := math.MaxInt16
		m.Textures[strconv.Itoa(boundaryTex)] = gen.BoundaryWall

		for x := 0; x < sizeX; x++ {
			for y := 0; y < sizeY; y++ {
				if x == 0 || y == 0 || x == sizeX-1 || y == sizeY-1 {
					level[x][y] = boundaryTex
				}
			}
		}
	}

	// TODO next: create "prefab" structures

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

func (m *Map) GetMapTexture(texIndex string) MapTexture {
	return m.Textures[texIndex]
}
