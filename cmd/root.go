package cmd

import (
	"github.com/pixelmek-3d/pixelmek-3d/game"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	"github.com/spf13/cobra"
	globalViper "github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "pixelmek-3d",
	Short: "PixelMek 3D is an unofficial BattleTech raycasted game",
	Long: `PixelMek 3D is an unofficial BattleTech raycasted game using community contributed pixel mech artwork.
		   Available at https://github.com/pixelmek-3d/pixelmek-3d`,
	Run: func(cmd *cobra.Command, args []string) {
		// run the game
		g := game.NewGame()
		g.Run()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(resources.InitConfig)

	// global flags that are not persisted in config file
	rootCmd.PersistentFlags().Bool(game.PARAM_KEY_DEBUG, false, "developer debug mode")
	rootCmd.PersistentFlags().Bool(game.PARAM_KEY_DISABLE_SHADERS, false, "disable shaders")

	globalViper.BindPFlag(game.PARAM_KEY_DEBUG, rootCmd.PersistentFlags().Lookup(game.PARAM_KEY_DEBUG))
	globalViper.BindPFlag(game.PARAM_KEY_DISABLE_SHADERS, rootCmd.PersistentFlags().Lookup(game.PARAM_KEY_DISABLE_SHADERS))
	globalViper.SetDefault(game.PARAM_KEY_DEBUG, false)
	globalViper.SetDefault(game.PARAM_KEY_DISABLE_SHADERS, false)
}
