package game

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pixelmek-3d/pixelmek-3d/game/common"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	log "github.com/sirupsen/logrus"
)

const NumWeaponGroups = 5

var userWeaponGroups map[string]_unitWeaponGroups

type _unitWeaponGroups struct {
	// Groups stores weapon group index lists by weapon index
	WeaponGroups [][]uint `json:"weapon_groups"`
}

func init() {
	userWeaponGroups = make(map[string]_unitWeaponGroups, 0)
}

func restoreUserWeaponGroups() error {
	log.Debug("restoring weapon groups file ", resources.UserWeaponGroupsFile)
	if _, err := os.Stat(resources.UserWeaponGroupsFile); err != nil {
		// weapon groups file does not yet exist, handle without failure
		return nil
	}

	userGroupsFile, err := os.Open(resources.UserWeaponGroupsFile)
	if err != nil {
		log.Error(err)
		return err
	}
	defer userGroupsFile.Close()

	fileBytes, err := io.ReadAll(userGroupsFile)
	if err != nil {
		log.Error(err)
		return err
	}

	if len(fileBytes) == 0 {
		// handle empty file without error
		return nil
	}

	err = json.Unmarshal(fileBytes, &userWeaponGroups)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func saveUserWeaponGroups() error {
	log.Debug("saving weapon groups file ", resources.UserWeaponGroupsFile)

	userGroupsPath := filepath.Dir(resources.UserWeaponGroupsFile)
	if _, err := os.Stat(userGroupsPath); os.IsNotExist(err) {
		err = os.MkdirAll(userGroupsPath, os.ModePerm)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	userGroupsFile, err := os.Create(resources.UserWeaponGroupsFile)
	if err != nil {
		log.Error(err)
		return err
	}
	defer userGroupsFile.Close()

	weaponGroupsJson, _ := json.Marshal(userWeaponGroups)
	_, err = userGroupsFile.Write(weaponGroupsJson)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func setUnitWeaponGroups(unit model.Unit, weaponGroups [][]model.Weapon) {
	unitKey := _getUnitWeaponsKey(unit)

	unitWeaponGroups := make([][]uint, len(unit.Armament()))
	for i, w := range unit.Armament() {
		unitWeaponGroups[i] = model.GetGroupsForWeapon(w, weaponGroups)
	}
	userWeaponGroups[unitKey] = _unitWeaponGroups{
		WeaponGroups: unitWeaponGroups,
	}
}

func getUnitWeaponGroups(unit model.Unit) [][]model.Weapon {
	unitKey := _getUnitWeaponsKey(unit)

	weaponGroups := make([][]model.Weapon, NumWeaponGroups)
	for i := 0; i < cap(weaponGroups); i++ {
		weaponGroups[i] = make([]model.Weapon, 0, len(unit.Armament()))
	}

	if wg, ok := userWeaponGroups[unitKey]; ok {
		// restore saved weapon groups
		for weaponIndex, groups := range wg.WeaponGroups {
			weapon := unit.Armament()[weaponIndex]
			for _, groupIndex := range groups {
				weaponGroups[groupIndex] = append(weaponGroups[groupIndex], weapon)
			}
		}
	} else {
		// initialize all weapons as only in first weapon group
		weaponGroups[0] = append(weaponGroups[0], unit.Armament()...)
	}
	return weaponGroups
}

func _getUnitWeaponsKey(unit model.Unit) string {
	// generate checksum for the unique unit and weapons it has as key to store and retrieve weapon groups
	var weaponsStr string
	for i, weapon := range unit.Armament() {
		weaponsStr += fmt.Sprintf("%d:%s|", i, weapon.ShortName())
	}
	weaponsChecksum := common.GetMD5Sum(weaponsStr)

	unitKey := fmt.Sprintf("%s_%s_%s", unit.Name(), unit.Variant(), weaponsChecksum)
	unitKey = strings.ReplaceAll(unitKey, " ", "_")
	return unitKey
}
