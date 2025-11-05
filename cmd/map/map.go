package mapcmd

import (
	"fmt"
	"strings"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	MapCmd.AddCommand(launchCmd)
	MapCmd.AddCommand(imageCmd)

	MapCmd.Flags().BoolVar(&listMaps, "list", false, "lists all map files")
}

var (
	listMaps bool
	mapFile  string
	MapCmd   = &cobra.Command{
		Use:   "map",
		Short: "Commands related to Maps",
		Run: func(cmd *cobra.Command, args []string) {
			if listMaps {
				resources.InitResources()
				mapFilenames, err := model.ListMapFilenames()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Print("Map List:\n", strings.Join(mapFilenames, "\n"))
				return
			}
			if len(args) == 0 {
				cmd.Help()
				return
			}
		},
	}
)
