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

type MissionImageFlags struct {
	renderDropZone      bool
	renderNavPoints     bool
	renderEnemyUnits    bool
	renderFriendlyUnits bool
}

func init() {
	imageCmd.Flags().StringVarP(&outImagePath, "output", "o", "", "[required] output png image path")
	imageCmd.MarkFlagRequired("output")

	mapcmd.BindMapImageFlags(imageCmd, &mapImageFlags)

	imageCmd.Flags().BoolVar(&missionImageFlags.renderDropZone, "render-drop-zone", true, "render the drop zone")
	imageCmd.Flags().BoolVar(&missionImageFlags.renderNavPoints, "render-nav-points", true, "render the nav points")
	imageCmd.Flags().BoolVar(&missionImageFlags.renderEnemyUnits, "render-enemy units", true, "render all enemy units")
	imageCmd.Flags().BoolVar(&missionImageFlags.renderFriendlyUnits, "render-friendly-units", true, "render all friendly units")
}

var (
	outImagePath      string
	mapImageFlags     mapcmd.MapImageFlags
	missionImageFlags MissionImageFlags
	imageCmd          = &cobra.Command{
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
	res, err := model.LoadModelResources()
	if err != nil {
		log.Error("error loading model resources: ", res)
		os.Exit(1)
	}

	log.Debug("creating image from mission...")
	mapOpts := mapimage.MapImageOptions{
		PxPerCell:                 mapImageFlags.PxPerCell,
		RenderDefaultFloorTexture: mapImageFlags.RenderFloorTexture,
		FilterDefaultFloorTexture: mapImageFlags.FilterFloorTexture,
		RenderWallLines:           mapImageFlags.RenderWallLines,
		RenderGridLines:           mapImageFlags.RenderGridLines,
		GridCellDistance:          mapImageFlags.GridCellDistance,
	}
	missionOpts := missionimage.MissionImageOptions{
		RenderDropZone:      missionImageFlags.renderDropZone,
		RenderNavPoints:     missionImageFlags.renderNavPoints,
		RenderEnemyUnits:    missionImageFlags.renderEnemyUnits,
		RenderFriendlyUnits: missionImageFlags.renderFriendlyUnits,
	}
	image, err := missionimage.NewMissionImage(m, res, tex, mapOpts, missionOpts)
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
