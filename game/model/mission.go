package model

import (
	"fmt"
	"path"

	"github.com/pixelmek-3d/pixelmek-3d/game/resources"

	"github.com/go-playground/validator/v10"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

type Mission struct {
	missionMap   *Map
	Pathing      *Pathing            `yaml:"-"`
	Title        string              `yaml:"title" validate:"required"`
	Briefing     string              `yaml:"briefing" validate:"required"`
	MapPath      string              `yaml:"map" validate:"required"`
	MusicPath    string              `yaml:"music"`
	DropZone     *MissionDropZone    `yaml:"dropZone" validate:"required"`
	Lighting     *MapLighting        `yaml:"lighting,omitempty"`
	FloorBox     *MapTexture         `yaml:"floorBox,omitempty"`
	SkyBox       *MapTexture         `yaml:"skyBox,omitempty"`
	NavPoints    []*NavPoint         `yaml:"navPoints"`
	Objectives   *MissionObjectives  `yaml:"objectives" validate:"required"`
	Mechs        []MissionUnit       `yaml:"mechs"`
	Vehicles     []MissionUnit       `yaml:"vehicles"`
	Infantry     []MissionUnit       `yaml:"infantry"`
	VTOLs        []MissionFlyingUnit `yaml:"vtols"`
	Emplacements []MissionStaticUnit `yaml:"emplacements"`
}

func (m *Mission) Map() *Map {
	return m.missionMap
}

type MissionDropZone struct {
	Position [2]float64 `yaml:"position" validate:"required"`
	Heading  float64    `yaml:"heading"`
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

type MissionGuardArea struct {
	Position [2]float64 `yaml:"position"`
	Radius   float64    `yaml:"radius"`
}

type MissionUnitInterface interface {
	GetUnit() string
	GetPosition() geom.Vector2
}

type AllMissionUnitModels interface {
	Mech | Vehicle | Infantry | VTOL | Emplacement
}

type MissionUnitModels interface {
	Mech | Vehicle | Infantry
}

type MissionUnit struct {
	ID         string           `yaml:"id"`
	Team       int              `yaml:"team"`
	Unit       string           `yaml:"unit" validate:"required"`
	Position   [2]float64       `yaml:"position" validate:"required"`
	Heading    float64          `yaml:"heading"`
	PatrolPath [][2]float64     `yaml:"patrolPath"`
	GuardArea  MissionGuardArea `yaml:"guardArea"`
	GuardUnit  string           `yaml:"guardUnit"`
}

func (m MissionUnit) GetUnit() string {
	return m.Unit
}

func (m MissionUnit) GetPosition() geom.Vector2 {
	return geom.Vector2{X: m.Position[0], Y: m.Position[1]}
}

type MissionFlyingUnitModels interface {
	VTOL
}

type MissionFlyingUnit struct {
	ID         string           `yaml:"id"`
	Team       int              `yaml:"team"`
	Unit       string           `yaml:"unit" validate:"required"`
	Position   [2]float64       `yaml:"position" validate:"required"`
	ZPosition  float64          `yaml:"zPosition" validate:"required"`
	Heading    float64          `yaml:"heading"`
	PatrolPath [][2]float64     `yaml:"patrolPath"`
	GuardArea  MissionGuardArea `yaml:"guardArea"`
	GuardUnit  string           `yaml:"guardUnit"`
}

func (m MissionFlyingUnit) GetUnit() string {
	return m.Unit
}

func (m MissionFlyingUnit) GetPosition() geom.Vector2 {
	return geom.Vector2{X: m.Position[0], Y: m.Position[1]}
}

type MissionStaticUnitModels interface {
	Emplacement
}

type MissionStaticUnit struct {
	ID       string     `yaml:"id"`
	Team     int        `yaml:"team"`
	Unit     string     `yaml:"unit" validate:"required"`
	Position [2]float64 `yaml:"position" validate:"required"`
	Heading  float64    `yaml:"heading"`
}

func (m MissionStaticUnit) GetUnit() string {
	return m.Unit
}

func (m MissionStaticUnit) GetPosition() geom.Vector2 {
	return geom.Vector2{X: m.Position[0], Y: m.Position[1]}
}

type NavObjective int

const (
	NavNonObjective NavObjective = iota
	NavVisitObjective
	NavDustoffObjective
)

type NavPoint struct {
	Name      string     `yaml:"name" validate:"required"`
	Position  [2]float64 `yaml:"position" validate:"required"`
	image     *ebiten.Image
	visited   bool
	objective NavObjective
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

func (n *NavPoint) Objective() NavObjective {
	return n.objective
}

func (n *NavPoint) SetObjective(objective NavObjective) {
	n.objective = objective
}

func LoadMission(missionFile string) (*Mission, error) {
	v := validator.New()
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

	err = v.Struct(m)
	if err != nil {
		return nil, fmt.Errorf("[%s] %s", missionPath, err.Error())
	}

	// load mission map
	err = m.LoadMissionMap()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Mission) LoadMissionMap() error {
	var err error
	m.missionMap, err = LoadMap(m.MapPath)
	if err != nil {
		log.Error("Error loading map: ", m.MapPath)
		return err
	}

	// initialize map pathing
	m.Pathing = initPathing(m)

	// apply any defaults from map
	if m.MusicPath == "" && m.missionMap.MusicPath != "" {
		m.MusicPath = m.missionMap.MusicPath
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
	return nil
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
			// only folder with mission files expected
			continue
		}
		missionFilenames = append(missionFilenames, f.Name())
	}

	return missionFilenames, nil
}

func (o *MissionObjectives) Text() string {
	oText := ""
	if len(o.Destroy) > 0 {
		for _, destroy := range o.Destroy {
			if destroy.All {
				oText += "Destroy All Enemies\n"
				break
			}
			oText += "Destroy " + destroy.Unit + "\n"
		}
	}

	if len(o.Protect) > 0 {
		for _, protect := range o.Protect {
			oText += "Protect " + protect.Unit + "\n"
		}
	}

	if o.Nav != nil {
		if len(o.Nav.Visit) > 0 {
			for _, visit := range o.Nav.Visit {
				oText += "Visit Nav " + visit.Name + "\n"
			}
		}
		if len(o.Nav.Dustoff) > 0 {
			for _, dustoff := range o.Nav.Dustoff {
				oText += "Dustoff Nav " + dustoff.Name + "\n"
			}
		}
	}

	return oText
}
