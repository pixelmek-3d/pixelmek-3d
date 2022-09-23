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
	Mechs    map[string]*ModelMech
	Vehicles map[string]*ModelVehicle
	VTOLs    map[string]*ModelVTOL
	Infantry map[string]*ModelInfantry
}

// TODO: add field validation - https://github.com/go-playground/validator
type ModelMech struct {
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

type ModelVehicle struct {
	Name              string    `yaml:"name" validate:"required"`
	Variant           string    `yaml:"variant" validate:"required"`
	Image             string    `yaml:"image" validate:"required"`
	Tech              ModelTech `yaml:"tech" validate:"required"`
	Tonnage           float64   `yaml:"tonnage" validate:"gt=0,lte=200"`
	Speed             float64   `yaml:"speed" validate:"gt=0,lte=250"`
	Armor             float64   `yaml:"armor" validate:"gte=0"`
	Structure         float64   `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64   `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64   `yaml:"collisionHeight" validate:"gt=0"`
	Scale             float64   `yaml:"scale" validate:"gt=0"`
}

type ModelVTOL struct {
	Name              string    `yaml:"name" validate:"required"`
	Variant           string    `yaml:"variant" validate:"required"`
	Image             string    `yaml:"image" validate:"required"`
	Tech              ModelTech `yaml:"tech" validate:"required"`
	Tonnage           float64   `yaml:"tonnage" validate:"gt=0,lte=100"`
	Speed             float64   `yaml:"speed" validate:"gt=0,lte=250"`
	Armor             float64   `yaml:"armor" validate:"gte=0"`
	Structure         float64   `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64   `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64   `yaml:"collisionHeight" validate:"gt=0"`
	Scale             float64   `yaml:"scale" validate:"gt=0"`
}

type ModelInfantry struct {
	Name              string    `yaml:"name" validate:"required"`
	Variant           string    `yaml:"variant" validate:"required"`
	Image             string    `yaml:"image" validate:"required"`
	Tech              ModelTech `yaml:"tech" validate:"required"`
	Speed             float64   `yaml:"speed" validate:"gt=0,lte=250"`
	JumpJets          int       `yaml:"jumpJets" validate:"gte=0,lte=20"`
	Armor             float64   `yaml:"armor" validate:"gte=0"`
	Structure         float64   `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64   `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64   `yaml:"collisionHeight" validate:"gt=0"`
	Scale             float64   `yaml:"scale" validate:"gt=0"`
}

type TechBase int

const (
	CLAN TechBase = iota
	IS
)

type ModelTech struct {
	TechBase
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
			resources.Mechs = make(map[string]*ModelMech, len(unitFiles))

		case "vehicles":
			resources.Vehicles = make(map[string]*ModelVehicle, len(unitFiles))

		case "vtols":
			resources.VTOLs = make(map[string]*ModelVTOL, len(unitFiles))

		case "infantry":
			resources.Infantry = make(map[string]*ModelInfantry, len(unitFiles))

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
				m := &ModelMech{}
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
				m := &ModelVehicle{}
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
				m := &ModelVTOL{}
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
				m := &ModelInfantry{}
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
