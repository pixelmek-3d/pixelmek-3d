package mission

import (
	"fmt"
	"strings"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	MissionCmd.AddCommand(launchCmd)
	MissionCmd.AddCommand(imageCmd)

	MissionCmd.Flags().BoolVar(&listMissions, "list", false, "lists all mission files")
}

var (
	listMissions bool
	missionFile  string
	MissionCmd   = &cobra.Command{
		Use:   "mission",
		Short: "Commands related to Missions",
		Run: func(cmd *cobra.Command, args []string) {
			if listMissions {
				resources.InitResources()
				missionFilenames, err := model.ListMissionFilenames()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Print("Mission List:\n", strings.Join(missionFilenames, "\n"))
				return
			}
			if len(args) == 0 {
				cmd.Help()
				return
			}
		},
	}
)
