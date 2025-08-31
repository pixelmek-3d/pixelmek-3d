package mission

import (
	"fmt"
	"strings"

	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	MissionCmd.AddCommand(launchCmd)

	MissionCmd.Flags().BoolVar(&listMissions, "list", false, "lists all mission files")
}

var (
	listMissions bool
	MissionCmd   = &cobra.Command{
		Use:   "mission",
		Short: "Commands related to Missions",
		Run: func(cmd *cobra.Command, args []string) {
			if listMissions {
				resources.InitResources(0)
				missionFilenames := make([]string, 0, 64)
				missionFiles, err := resources.ReadDir("missions")
				if err != nil {
					log.Fatal(err)
				}
				for _, f := range missionFiles {
					if f.IsDir() {
						// only folder with mission files expected
						continue
					}
					missionFilenames = append(missionFilenames, f.Name())
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
