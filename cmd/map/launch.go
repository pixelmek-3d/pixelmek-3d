package mapcmd

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
	launchCmd.Flags().StringVar(&mechFile, "mech", "", "mech file")
}

var (
	mechFile  string
	launchCmd = &cobra.Command{
		Use:   "launch [MAP_FILE]",
		Short: "Launch game directly into a map for instant action",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			mapFile = args[0]

			// load the map path specified
			g := game.NewGame()
			_, err := g.LoadInstantAction(mapFile)
			if err != nil {
				log.Error("Error loading map file: ", mapFile)
				log.Error(err)

				mapFilenames, _ := model.ListMapFilenames()
				if len(mapFilenames) > 0 {
					log.Error("Map files available:\n", strings.Join(mapFilenames[:], "\n"))
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

			// jump straight to the game scene
			g.SetInitialSceneFunc(game.NewGameScene)

			g.Run()
		},
	}
)
