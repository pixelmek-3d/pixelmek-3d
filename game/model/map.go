package model

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/harbdog/raycaster-go/geom"
	"gopkg.in/yaml.v3"
)

type Map struct {
	Levels   [][][]int             `yaml: levels`
	Textures map[string]MapTexture `yaml: textures`
	FloorBox MapTexture            `yaml: floorBox`
	SkyBox   MapTexture            `yaml: skyBox`
	Sprites  []MapSprite           `yaml: sprites`
}

type MapTexture struct {
	Image string `yaml: image`
}

type MapSprite struct {
	Image    string     `yaml: image`
	Position [2]float64 `yaml: position`
}

func (m *Map) NumLevels() int {
	return 3
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
	mapsPath := filepath.Join("game", "model", "maps", mapFile)

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
	if len(m.Textures) == 0 || len(m.Levels) == 0 {
		return m, fmt.Errorf("one or more entry in textures and levels are required")
	}

	return m, nil
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
