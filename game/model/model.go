package model

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type ModelResources struct {
	Mechs    map[string]*ModelMechResource
	Vehicles map[string]*ModelVehicleResource
	VTOLs    map[string]*ModelVTOLResource
	Infantry map[string]*ModelInfantryResource
}

type TechBase int

const (
	CLAN TechBase = iota
	IS
)

type ModelTech struct {
	TechBase
}

type ModelMechResource struct {
	Name              string    `yaml:"name" validate:"required"`
	Variant           string    `yaml:"variant" validate:"required"`
	Image             string    `yaml:"image" validate:"required"`
	Tech              ModelTech `yaml:"tech" validate:"required"`
	Tonnage           float64   `yaml:"tonnage" validate:"gt=0,lte=200"`
	Speed             float64   `yaml:"speed" validate:"gt=0,lte=250"`
	JumpJets          int       `yaml:"jumpJets" validate:"gte=0,lte=20"`
	Armor             float64   `yaml:"armor" validate:"gte=0"`
	Structure         float64   `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64   `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64   `yaml:"collisionHeight" validate:"gt=0"`
	Scale             float64   `yaml:"scale" validate:"gt=0"`
}

type ModelVehicleResource struct {
	Name              string                   `yaml:"name" validate:"required"`
	Variant           string                   `yaml:"variant" validate:"required"`
	Image             string                   `yaml:"image" validate:"required"`
	ImageSheet        *ModelResourceImageSheet `yaml:"imageSheet"`
	Tech              ModelTech                `yaml:"tech" validate:"required"`
	Tonnage           float64                  `yaml:"tonnage" validate:"gt=0,lte=200"`
	Speed             float64                  `yaml:"speed" validate:"gt=0,lte=250"`
	Armor             float64                  `yaml:"armor" validate:"gte=0"`
	Structure         float64                  `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64                  `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64                  `yaml:"collisionHeight" validate:"gt=0"`
	Scale             float64                  `yaml:"scale" validate:"gt=0"`
}

type ModelVTOLResource struct {
	Name              string                   `yaml:"name" validate:"required"`
	Variant           string                   `yaml:"variant" validate:"required"`
	Image             string                   `yaml:"image" validate:"required"`
	ImageSheet        *ModelResourceImageSheet `yaml:"imageSheet"`
	Tech              ModelTech                `yaml:"tech" validate:"required"`
	Tonnage           float64                  `yaml:"tonnage" validate:"gt=0,lte=100"`
	Speed             float64                  `yaml:"speed" validate:"gt=0,lte=250"`
	Armor             float64                  `yaml:"armor" validate:"gte=0"`
	Structure         float64                  `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64                  `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64                  `yaml:"collisionHeight" validate:"gt=0"`
	Scale             float64                  `yaml:"scale" validate:"gt=0"`
}

type ModelInfantryResource struct {
	Name              string                   `yaml:"name" validate:"required"`
	Variant           string                   `yaml:"variant" validate:"required"`
	Image             string                   `yaml:"image" validate:"required"`
	ImageSheet        *ModelResourceImageSheet `yaml:"imageSheet"`
	Tech              ModelTech                `yaml:"tech" validate:"required"`
	Speed             float64                  `yaml:"speed" validate:"gt=0,lte=250"`
	JumpJets          int                      `yaml:"jumpJets" validate:"gte=0,lte=20"`
	Armor             float64                  `yaml:"armor" validate:"gte=0"`
	Structure         float64                  `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64                  `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64                  `yaml:"collisionHeight" validate:"gt=0"`
	Scale             float64                  `yaml:"scale" validate:"gt=0"`
}

type ModelResourceImageSheet struct {
	Columns     int             `yaml:"columns" validate:"gt=0"`
	Rows        int             `yaml:"rows" validate:"gt=0"`
	AngleFacing map[float64]int `yaml:"angleFacing"`
}

// Unmarshals into TechBase
func (t *ModelTech) UnmarshalText(b []byte) error {
	str := strings.Trim(string(b), `"`)

	clan, is := "clan", "is"

	switch str {
	case clan:
		t.TechBase = CLAN
	case is:
		t.TechBase = IS
	default:
		return fmt.Errorf("unknown tech value '%s', must be one of: [%s, %s]", str, clan, is)
	}

	return nil
}

func LoadModels() (*ModelResources, error) {
	unitsPath := filepath.Join("game", "resources", "units")

	// TODO: load and validate all units
	v := validator.New()
	resources := &ModelResources{}

	unitsTypes, err := filesInPath(unitsPath)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	for _, t := range unitsTypes {
		if !t.IsDir() {
			// only folders with unit type name expected
			continue
		}

		unitType := t.Name()
		unitTypePath := filepath.Join(unitsPath, unitType)
		unitFiles, err := filesInPath(unitTypePath)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		// initialize map for each recognized unit type
		switch unitType {
		case "mechs":
			resources.Mechs = make(map[string]*ModelMechResource, len(unitFiles))

		case "vehicles":
			resources.Vehicles = make(map[string]*ModelVehicleResource, len(unitFiles))

		case "vtols":
			resources.VTOLs = make(map[string]*ModelVTOLResource, len(unitFiles))

		case "infantry":
			resources.Infantry = make(map[string]*ModelInfantryResource, len(unitFiles))

		}

		for _, u := range unitFiles {
			if u.IsDir() {
				// TODO: support recursive directory structure?
				continue
			}

			fileName := u.Name()
			unitYaml, err := ioutil.ReadFile(filepath.Join(unitTypePath, fileName))
			if err != nil {
				log.Fatal(err)
				return nil, err
			}

			switch unitType {
			case "mechs":
				m := &ModelMechResource{}
				err = yaml.Unmarshal(unitYaml, m)
				if err != nil {
					log.Fatal(err)
					return nil, err
				}

				err = v.Struct(m)
				if err != nil {
					log.Fatal(err)
					return nil, err
				}

				resources.Mechs[fileName] = m

			case "vehicles":
				m := &ModelVehicleResource{}
				err = yaml.Unmarshal(unitYaml, m)
				if err != nil {
					log.Fatal(err)
					return nil, err
				}

				err = v.Struct(m)
				if err != nil {
					log.Fatal(err)
					return nil, err
				}

				resources.Vehicles[fileName] = m

			case "vtols":
				m := &ModelVTOLResource{}
				err = yaml.Unmarshal(unitYaml, m)
				if err != nil {
					log.Fatal(err)
					return nil, err
				}

				err = v.Struct(m)
				if err != nil {
					log.Fatal(err)
					return nil, err
				}

				resources.VTOLs[fileName] = m

			case "infantry":
				m := &ModelInfantryResource{}
				err = yaml.Unmarshal(unitYaml, m)
				if err != nil {
					log.Fatal(err)
					return nil, err
				}

				err = v.Struct(m)
				if err != nil {
					log.Fatal(err)
					return nil, err
				}

				resources.Infantry[fileName] = m

			}
		}

	}

	return resources, nil
}

func (r *ModelResources) GetMechResource(unit string) *ModelMechResource {
	if m, ok := r.Mechs[unit]; ok {
		return m
	}
	return nil
}

func (r *ModelResources) GetVehicleResource(unit string) *ModelVehicleResource {
	if m, ok := r.Vehicles[unit]; ok {
		return m
	}
	return nil
}

func (r *ModelResources) GetVTOLResource(unit string) *ModelVTOLResource {
	if m, ok := r.VTOLs[unit]; ok {
		return m
	}
	return nil
}

func (r *ModelResources) GetInfantryResource(unit string) *ModelInfantryResource {
	if m, ok := r.Infantry[unit]; ok {
		return m
	}
	return nil
}

func filesInPath(path string) ([]fs.DirEntry, error) {
	top, err := os.Open(path)
	if err != nil {
		return []fs.DirEntry{}, err
	}

	files, err := top.ReadDir(0)
	if err != nil {
		return []fs.DirEntry{}, err
	}

	return files, nil
}
