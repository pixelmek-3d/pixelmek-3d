package unit

import (
	"fmt"
	"strings"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	UnitCmd.AddCommand(animationCmd)

	UnitCmd.Flags().BoolVar(&listUnits, "list", false, "lists all unit files")

	UnitCmd.Flags().BoolVar(&listCamo, "list-camo", false, "lists all camo files")
}

var (
	listUnits bool
	listCamo  bool
	unitFile  string
	UnitCmd   = &cobra.Command{
		Use:   "unit",
		Short: "Commands related to Units",
		Run: func(cmd *cobra.Command, args []string) {
			if listUnits || listCamo {
				resources.InitResources()
				if listUnits {
					unitFilenames, err := model.ListUnitFilenames()
					if err != nil {
						log.Fatal(err)
					}
					fmt.Print("Unit List:\n", strings.Join(unitFilenames, "\n"))
				}
				if listCamo {
					camoFilenames, err := model.ListCamoFilenames()
					if err != nil {
						log.Fatal(err)
					}
					fmt.Print("Camo List:\n", strings.Join(camoFilenames, "\n"))
				}
				return
			}

			if len(args) == 0 {
				cmd.Help()
			}
		},
	}
)
