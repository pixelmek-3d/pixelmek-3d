package model

import (
	"path"

	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

type Mission struct {
	missionMap   *Map
	Title        string               `yaml:"title" validate:"required"`
	Briefing     string               `yaml:"briefing" validate:"required"`
	MapPath      string               `yaml:"map" validate:"required"`
	MusicPath    string               `yaml:"music"`
	DropZone     *MissionDropZone     `yaml:"dropZone" validate:"required"`
	Lighting     *MapLighting         `yaml:"lighting,omitempty"`
	FloorBox     *MapTexture          `yaml:"floorBox,omitempty"`
	SkyBox       *MapTexture          `yaml:"skyBox,omitempty"`
	NavPoints    []*NavPoint          `yaml:"navPoints"`
	Objectives   *MissionObjectives   `yaml:"objectives" validate:"required"`
	Mechs        []MissionMech        `yaml:"mechs"`
	Vehicles     []MissionVehicle     `yaml:"vehicles"`
	VTOLs        []MissionVTOL        `yaml:"vtols"`
	Infantry     []MissionInfantry    `yaml:"infantry"`
	Emplacements []MissionEmplacement `yaml:"emplacements"`
}

func (m *Mission) Map() *Map {
	return m.missionMap
}

type MissionDropZone struct {
	Position [2]float64 `yaml:"position" validate:"required"`
	Heading  float64    `yaml:"heading" validate:"required"`
}

type MissionObjectives struct {
	Destroy []*MissionDestroyObjectives `yaml:"destroy"`
	Protect []*MissionProtectObjectives `yaml:"protect"`
	Nav     *MissionNavObjectives       `yaml:"nav"`
}

type MissionDestroyObjectives struct {
	All  bool   `yaml:"all,omitempty"`
	Unit string `yaml:"unit,omitempty"`
}

type MissionProtectObjectives struct {
	Unit string `yaml:"unit,omitempty"`
}

type MissionNavObjectives struct {
	Visit   []*MissionNavVisit   `yaml:"visit,omitempty"`
	Dustoff []*MissionNavDustoff `yaml:"dustoff,omitempty"`
}

type MissionNavVisit struct {
	Name string `yaml:"name" validate:"required"`
}

type MissionNavDustoff struct {
	Name string `yaml:"name" validate:"required"`
}

type MissionMech struct {
	ID         string       `yaml:"id"`
	Unit       string       `yaml:"unit" validate:"required"`
	Position   [2]float64   `yaml:"position" validate:"required"`
	PatrolPath [][2]float64 `yaml:"patrolPath"`
}

type MissionVehicle struct {
	ID         string       `yaml:"id"`
	Unit       string       `yaml:"unit" validate:"required"`
	Position   [2]float64   `yaml:"position" validate:"required"`
	PatrolPath [][2]float64 `yaml:"patrolPath"`
}

type MissionVTOL struct {
	ID         string       `yaml:"id"`
	Unit       string       `yaml:"unit" validate:"required"`
	Position   [2]float64   `yaml:"position" validate:"required"`
	ZPosition  float64      `yaml:"zPosition" validate:"required"`
	PatrolPath [][2]float64 `yaml:"patrolPath"`
}

type MissionInfantry struct {
	ID         string       `yaml:"id"`
	Unit       string       `yaml:"unit" validate:"required"`
	Position   [2]float64   `yaml:"position" validate:"required"`
	PatrolPath [][2]float64 `yaml:"patrolPath"`
}

type MissionEmplacement struct {
	ID       string     `yaml:"id"`
	Unit     string     `yaml:"unit" validate:"required"`
	Position [2]float64 `yaml:"position" validate:"required"`
}

type NavPoint struct {
	Name        string     `yaml:"name" validate:"required"`
	Position    [2]float64 `yaml:"position" validate:"required"`
	image       *ebiten.Image
	visited     bool
	isObjective bool
	isDustoff   bool
}

func (n *NavPoint) Pos() geom.Vector2 {
	return geom.Vector2{X: n.Position[0], Y: n.Position[1]}
}

func (n *NavPoint) Image() *ebiten.Image {
	return n.image
}

func (n *NavPoint) SetImage(image *ebiten.Image) {
	n.image = image
}

func (n *NavPoint) Visited() bool {
	return n.visited
}

func (n *NavPoint) SetVisited(visited bool) {
	n.visited = visited
}

func (n *NavPoint) IsObjective() bool {
	return n.isObjective
}

func (n *NavPoint) SetIsObjective(isObjective bool) {
	n.isObjective = isObjective
}

func (n *NavPoint) IsDustoff() bool {
	return n.isDustoff
}

func (n *NavPoint) SetIsDustoff(isDustoff bool) {
	n.isDustoff = isDustoff
}

func LoadMission(missionFile string) (*Mission, error) {
	missionPath := path.Join("missions", missionFile)

	missionYaml, err := resources.ReadFile(missionPath)
	if err != nil {
		return nil, err
	}

	m := &Mission{}
	err = yaml.Unmarshal(missionYaml, m)
	if err != nil {
		return nil, err
	}

	// load mission map
	m.missionMap, err = LoadMap(m.MapPath)
	if err != nil {
		log.Error("Error loading map", m.MapPath)
		return nil, err
	}

	// apply optional overrides to map
	if m.Lighting != nil {
		m.missionMap.Lighting = *m.Lighting
	}
	if m.FloorBox != nil {
		m.missionMap.FloorBox = *m.FloorBox
	}
	if m.SkyBox != nil {
		m.missionMap.SkyBox = *m.SkyBox
	}

	return m, nil
}

func ListMissionFilenames() ([]string, error) {
	missionFilenames := make([]string, 0, 64)
	missionsPath := "missions"
	missionFiles, err := resources.ReadDir(missionsPath)
	if err != nil {
		return missionFilenames, err
	}

	for _, f := range missionFiles {
		if f.IsDir() {
			// only folders with unit type name expected
			continue
		}

		missionFilenames = append(missionFilenames, f.Name())
	}

	return missionFilenames, nil
}
