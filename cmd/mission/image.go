package mission

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	mapcmd "github.com/pixelmek-3d/pixelmek-3d/cmd/map"
	"github.com/pixelmek-3d/pixelmek-3d/game"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/export"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/mapimage"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/missionimage"
	"github.com/pixelmek-3d/pixelmek-3d/game/texture"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	imageCmd.Flags().StringVarP(&outImagePath, "output", "o", "", "[required] output png image path")
	imageCmd.MarkFlagRequired("output")

	mapcmd.BindMapImageFlags(imageCmd, &mapImageFlags)
}

var (
	outImagePath  string
	mapImageFlags mapcmd.MapImageFlags
	imageCmd      = &cobra.Command{
		Use:   "image [MISSION_FILE]",
		Short: "Export mission file as an image",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			missionFile = args[0]

			// initialize game resources without running the actual game loop
			g := game.NewGame()
			g.Pause()

			// expand tilde as home directory
			if strings.HasPrefix(outImagePath, "~/") {
				dirname, _ := os.UserHomeDir()
				outImagePath = filepath.Join(dirname, outImagePath[2:])
			}

			if err := os.MkdirAll(filepath.Dir(outImagePath), 0755); err != nil {
				log.Fatal(err)
			}

			// mock game loop required for certain offscreen ebitengine render functions
			ebiten.SetWindowTitle("Exporting mission image " + missionFile + " ...")
			mapExport := export.NewExportLoop(doMissionExport)
			if err := ebiten.RunGame(mapExport); err != nil {
				log.Fatal(err)
			}
		},
	}
)

func doMissionExport() {
	log.Debug("loading mission file ", missionFile, "...")
	m, err := model.LoadMission(missionFile)
	if err != nil {
		log.Error("error loading mission file: ", missionFile)
		os.Exit(1)
	}

	log.Debug("loading mission map textures...")
	tex := texture.NewTextureHandler(m.Map())

	log.Debug("creating image from mission...")
	mapOpts := mapimage.MapImageOptions{
		PxPerCell:                 mapImageFlags.PxPerCell,
		RenderDefaultFloorTexture: mapImageFlags.RenderFloorTexture,
		RenderGridLines:           mapImageFlags.RenderGridLines,
		RenderWallLines:           mapImageFlags.RenderWallLines,
	}
	missionOpts := missionimage.MissionImageOptions{}
	image, err := missionimage.NewMissionImage(m, tex, mapOpts, missionOpts)
	if err != nil {
		log.Error("error creating mission image: ", err)
		os.Exit(1)
	}

	log.Debug("exporting image to file...")
	err = render.SaveImageAsPNG(image, outImagePath)
	if err != nil {
		log.Error("error exporting mission image: ", err)
		os.Exit(1)
	}

	log.Info("mission image exported: " + outImagePath)
	os.Exit(0)
}
