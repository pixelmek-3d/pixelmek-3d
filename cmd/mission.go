package cmd

import (
	"os"
	"sort"
	"strings"

	"github.com/pixelmek-3d/pixelmek-3d/game"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func init() {
	rootCmd.AddCommand(missionCmd)

	missionCmd.Flags().StringVarP(&missionFile, "file", "f", "", "mission file")
	missionCmd.MarkFlagRequired("file")

	missionCmd.Flags().StringVar(&mechFile, "mech", "", "mech file")
}

var (
	missionFile string
	mechFile    string
	missionCmd  = &cobra.Command{
		Use:   "mission",
		Short: "Load directly into a mission",
		Run: func(cmd *cobra.Command, args []string) {
			g := game.NewGame()

			// load the mission path specified
			_, err := g.LoadMission(missionFile)
			if err != nil {
				log.Error("Error loading mission file: ", missionFile)
				log.Error(err)

				missionList, _ := model.ListMissionFilenames()
				if len(missionList) > 0 {
					log.Error("Mission files available:\n", strings.Join(missionList[:], "\n"))
				}
				os.Exit(1)
			}

			// load the unit file specified, or random unit if not provided
			if len(mechFile) == 0 {
				g.SetPlayerUnit(g.RandomUnit(model.MechResourceType))
			} else {
				unit := g.LoadUnit(model.MechResourceType, mechFile)
				if unit == nil {
					log.Error("Error loading mech file: ", mechFile)
					unitList := make([]string, 0, len(g.Resources().Mechs))
					for k := range g.Resources().Mechs {
						unitList = append(unitList, k)
					}
					sort.Strings(unitList)
					if len(unitList) > 0 {
						log.Error("Mech files available:\n", strings.Join(unitList[:], "\n"))
					}
					os.Exit(1)
				} else {
					g.SetPlayerUnit(unit)
				}
			}

			// initialize the menu system
			g.SetScene(game.NewMenuScene(g))

			// jump straight to the game scene
			g.SetScene(game.NewGameScene(g))

			g.Run()
		},
	}
)
