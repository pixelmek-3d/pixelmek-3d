package weapon

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pixelmek-3d/pixelmek-3d/game"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	reportCmd.Flags().StringVarP(&outPath, "output", "o", "", "[required] report output path")
}

var (
	g         *game.Game
	outPath   string
	reportCmd = &cobra.Command{
		Use:   "report [UNIT_FILE]",
		Short: "Create weapon statistics CSV report",
		Run: func(cmd *cobra.Command, args []string) {
			// initialize game resources without running the actual game loop
			resources.InitResources()
			r, err := model.LoadModelResources()
			if err != nil {
				log.Fatal(err)
			}

			// expand tilde as home directory
			if strings.HasPrefix(outPath, "~/") {
				dirname, _ := os.UserHomeDir()
				outPath = filepath.Join(dirname, outPath[2:])
			}
			if err := os.MkdirAll(filepath.Dir(outPath), 0644); err != nil {
				log.Fatal(err)
			}

			// gather all weapons by type
			weaponListByType := map[model.WeaponType][]model.Weapon{
				model.ENERGY:    make([]model.Weapon, 0, len(r.EnergyWeapons)),
				model.BALLISTIC: make([]model.Weapon, 0, len(r.BallisticWeapons)),
				model.MISSILE:   make([]model.Weapon, 0, len(r.MissileWeapons)),
			}
			for _, r := range r.EnergyWeapons {
				w := model.EnergyWeaponModel(r)
				weaponListByType[w.Type()] = append(weaponListByType[w.Type()], &w)
			}
			for _, r := range r.BallisticWeapons {
				w := model.BallisticWeaponModel(r)
				weaponListByType[w.Type()] = append(weaponListByType[w.Type()], &w)
			}
			for _, r := range r.MissileWeapons {
				w := model.MissileWeaponModel(r)
				weaponListByType[w.Type()] = append(weaponListByType[w.Type()], &w)
			}

			for _, wType := range []model.WeaponType{model.ENERGY, model.BALLISTIC, model.MISSILE} {
				wList := weaponListByType[wType]
				slices.SortFunc(wList, func(a, b model.Weapon) int {
					// TODO: sort within weapon sizes so AC5 comes before AC10 and AC20
					return strings.Compare(a.Name(), b.Name())
				})
				for _, w := range wList {
					log.Infof("[%v] %s", wType, w.Name())
				}
			}

			// TODO: create CSV report with weapon stats
		},
	}
)
