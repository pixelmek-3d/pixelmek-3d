package weapon

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	reportCmd.Flags().StringVarP(&outPath, "output", "o", "", "[required] report output path")
}

const (
	NAME         = "Name"
	SNAME        = "Short Name"
	FILE         = "File"
	TYPE         = "Type"
	TECH         = "Tech"
	HEAT         = "Heat"
	DAMAGE       = "Damage"
	COOLDOWN     = "Cooldown"
	DPS          = "DPS"
	RANGE        = "Range"
	VELOCITY     = "Velocity"
	TONS         = "Tons"
	AMMO_PER_TON = "Ammo Per Ton"
)

var headers = []string{NAME, SNAME, FILE, TYPE, TECH, HEAT, DAMAGE, COOLDOWN, DPS, RANGE, VELOCITY, TONS, AMMO_PER_TON}

func weaponRowData(w model.Weapon) map[string]string {
	return map[string]string{
		NAME:         w.Name(),
		SNAME:        w.ShortName(),
		FILE:         w.File(),
		TYPE:         w.Type().String(),
		TECH:         w.Tech().String(),
		HEAT:         floatString(w.Heat()),
		DAMAGE:       floatString(w.Damage()),
		DPS:          floatString(w.Damage() / w.MaxCooldown()),
		COOLDOWN:     floatString(w.MaxCooldown()),
		RANGE:        floatString(w.Distance()),
		VELOCITY:     floatString(w.Velocity()),
		TONS:         floatString(w.Tonnage()),
		AMMO_PER_TON: fmt.Sprintf("%d", w.AmmoPerTon()),
	}
}

var (
	outPath          string
	weaponClassRegex = regexp.MustCompile(`(\D+)(\d+)`)
	reportCmd        = &cobra.Command{
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
			if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
				log.Fatal(err)
			}

			// create report file
			file, err := os.Create(filepath.Join(outPath, "weapon_report.csv"))
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			// initialize writer and write header row
			writer := csv.NewWriter(file)
			defer writer.Flush()
			if err := writer.Write(headers); err != nil {
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

			// gather weapon data to write to each row
			records := make([]map[string]string, 0, len(r.EnergyWeapons)+len(r.BallisticWeapons)+len(r.MissileWeapons))
			for _, wType := range []model.WeaponType{model.ENERGY, model.BALLISTIC, model.MISSILE} {
				wList := weaponListByType[wType]
				slices.SortFunc(wList, weaponSortFunc)
				for _, w := range wList {
					if strings.HasPrefix(w.File(), "_") {
						continue
					}
					rowData := weaponRowData(w)
					records = append(records, rowData)
				}
			}

			// write row data records
			for _, record := range records {
				row := make([]string, len(headers))
				for i, header := range headers {
					row[i] = record[header]
				}
				if err := writer.Write(row); err != nil {
					log.Fatal(err)
				}
			}

			log.Infof("Weapon report created: %s", file.Name())
		},
	}
)

func weaponSortFunc(a, b model.Weapon) int {
	// sort based on weapon classifications such that PPC comes after Lasers
	if a.Classification() != b.Classification() {
		return int(a.Classification()) - int(b.Classification())
	}
	// sort such that AC5 comes before AC10 and AC20
	aMatch, bMatch := weaponClassRegex.FindStringSubmatch(a.Name()), weaponClassRegex.FindStringSubmatch(b.Name())
	if aMatch != nil && bMatch != nil && aMatch[1] == bMatch[1] {
		aNumber, aErr := strconv.Atoi(aMatch[2])
		bNumber, bErr := strconv.Atoi(bMatch[2])
		if aErr == nil && bErr == nil {
			return aNumber - bNumber
		}
	}
	return strings.Compare(a.Name(), b.Name())
}

func floatString(f float64) string {
	s := fmt.Sprintf("%0.2f", f)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}
