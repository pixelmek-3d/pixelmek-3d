package weapon

import (
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

			log.Info("--- ENERGY ---")
			for _, w := range r.EnergyWeapons {
				log.Infof("[%s] %s - %s", w.File, w.ShortName, w.Name)
			}
			log.Info("--- BALLISTIC ---")
			for _, w := range r.BallisticWeapons {
				log.Infof("[%s] %s - %s", w.File, w.ShortName, w.Name)
			}
			log.Info("--- MISSILE ---")
			for _, w := range r.MissileWeapons {
				log.Infof("[%s] %s - %s", w.File, w.ShortName, w.Name)
			}

			// expand tilde as home directory
			// if strings.HasPrefix(outPath, "~/") {
			// 	dirname, _ := os.UserHomeDir()
			// 	outPath = filepath.Join(dirname, outPath[2:])
			// }

			// if err := os.MkdirAll(filepath.Dir(outPath), 0644); err != nil {
			// 	log.Fatal(err)
			// }

			// TODO: create CSV report with weapon stats
		},
	}
)
