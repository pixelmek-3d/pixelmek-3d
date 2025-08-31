package mapcmd

import (
	"github.com/spf13/cobra"
)

func init() {
	MapCmd.AddCommand(imageCmd)

	// TODO: add --list param to list available mission files
}

var (
	MapCmd = &cobra.Command{
		Use:   "map",
		Short: "Commands related to Maps",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
		},
	}
)
