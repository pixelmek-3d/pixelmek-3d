package model

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Mission struct {
	missionMap *Map
	MapPath    string            `yaml:"map"`
	Mechs      []MissionMech     `yaml:"mechs"`
	Vehicles   []MissionVehicle  `yaml:"vehicles"`
	VTOLs      []MissionVTOL     `yaml:"vtols"`
	Infantry   []MissionInfantry `yaml:"infantry"`
}

func (m *Mission) Map() *Map {
	return m.missionMap
}

type MissionMech struct {
	Unit       string       `yaml:"unit"`
	Position   [2]float64   `yaml:"position"`
	PatrolPath [][2]float64 `yaml:"patrolPath"`
}

type MissionVehicle struct {
	Unit       string       `yaml:"unit"`
	Position   [2]float64   `yaml:"position"`
	PatrolPath [][2]float64 `yaml:"patrolPath"`
}

type MissionVTOL struct {
	Unit       string       `yaml:"unit"`
	Position   [2]float64   `yaml:"position"`
	ZPosition  float64      `yaml:"zPosition"`
	PatrolPath [][2]float64 `yaml:"patrolPath"`
}

type MissionInfantry struct {
	Unit       string       `yaml:"unit"`
	Position   [2]float64   `yaml:"position"`
	PatrolPath [][2]float64 `yaml:"patrolPath"`
}

func LoadMission(missionFile string) (*Mission, error) {
	missionPath := filepath.Join("game", "resources", "missions", missionFile)

	missionYaml, err := ioutil.ReadFile(missionPath)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	m := &Mission{}
	err = yaml.Unmarshal(missionYaml, m)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// load mission map
	m.missionMap, err = LoadMap(m.MapPath)
	if err != nil {
		log.Println("Error loading map", m.MapPath)
		return nil, err
	}

	return m, nil
}
