package weapon

import (
	"fmt"
	"strings"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	WeaponCmd.AddCommand(reportCmd)

	WeaponCmd.Flags().BoolVar(&listWeapons, "list", false, "lists all weapon files")
}

var (
	listWeapons bool
	WeaponCmd   = &cobra.Command{
		Use:   "weapon",
		Short: "Commands related to Weapons",
		Run: func(cmd *cobra.Command, args []string) {
			if listWeapons {
				resources.InitResources()
				weaponFilenames, err := model.ListWeaponFilenames()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Print("Weapon List:\n", strings.Join(weaponFilenames, "\n"))
				return
			}
			if len(args) == 0 {
				cmd.Help()
			}
		},
	}
)
