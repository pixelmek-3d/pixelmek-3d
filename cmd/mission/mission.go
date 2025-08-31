package mission

import (
	"github.com/spf13/cobra"
)

func init() {
	MissionCmd.AddCommand(launchCmd)

	// TODO: add --list param to list available mission files
}

var (
	MissionCmd = &cobra.Command{
		Use:   "mission",
		Short: "Commands related to Missions",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
		},
	}
)
