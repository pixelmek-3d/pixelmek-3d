package mapcmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/export"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/mapimage"
	"github.com/pixelmek-3d/pixelmek-3d/game/texture"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type MapImageFlags struct {
	PxPerCell          int
	RenderFloorTexture bool
	FilterFloorTexture bool
	RenderWallLines    bool
	RenderGridLines    bool
	GridCellDistance   int
}

func init() {
	imageCmd.Flags().StringVarP(&outImagePath, "output", "o", "", "[required] output png image path")
	imageCmd.MarkFlagRequired("output")

	BindMapImageFlags(imageCmd, &mapImageFlags)
}

func BindMapImageFlags(cmd *cobra.Command, imageFlags *MapImageFlags) {
	cmd.Flags().IntVar(&imageFlags.PxPerCell, "px-per-cell", 16, "number of pixels per map cell to render in each direction")
	cmd.Flags().BoolVar(&imageFlags.RenderFloorTexture, "render-floor-texture", true, "render the default floor texture")
	cmd.Flags().BoolVar(&imageFlags.FilterFloorTexture, "filter-floor-texture", false, "use scaling filter for the default floor texture to reduce hatch grid effect")
	cmd.Flags().BoolVar(&imageFlags.RenderWallLines, "render-wall-lines", true, "render the visibility lines surrounding walls")
	cmd.Flags().BoolVar(&imageFlags.RenderGridLines, "render-grid-lines", true, "render grid lines")
	cmd.Flags().IntVar(&imageFlags.GridCellDistance, "grid-cell-distance", 0, "cells per grid line (default: 1km of cells)")
}

var (
	outImagePath  string
	mapImageFlags MapImageFlags
	imageCmd      = &cobra.Command{
		Use:   "image [MAP_FILE]",
		Short: "Export map file as an image",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			mapPath = args[0]

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
			ebiten.SetWindowTitle("Exporting map image " + mapPath + " ...")
			mapExport := export.NewExportLoop(doMapExport)
			if err := ebiten.RunGame(mapExport); err != nil {
				log.Fatal(err)
			}
		},
	}
)

func doMapExport() {
	log.Debug("loading map file ", mapPath, "...")
	m, err := model.LoadMap(mapPath)
	if err != nil {
		log.Error("error loading map file: ", mapPath)
		os.Exit(1)
	}

	log.Debug("loading map textures...")
	tex := texture.NewTextureHandler(m)

	log.Debug("creating image from map...")
	mapOpts := mapimage.MapImageOptions{
		PxPerCell:                 mapImageFlags.PxPerCell,
		RenderDefaultFloorTexture: mapImageFlags.RenderFloorTexture,
		FilterDefaultFloorTexture: mapImageFlags.FilterFloorTexture,
		RenderWallLines:           mapImageFlags.RenderWallLines,
		RenderGridLines:           mapImageFlags.RenderGridLines,
		GridCellDistance:          mapImageFlags.GridCellDistance,
	}
	image, err := mapimage.NewMapImage(m, tex, mapOpts)
	if err != nil {
		log.Error("error creating map image: ", err)
		os.Exit(1)
	}

	log.Debug("exporting image to file...")
	err = render.SaveImageAsPNG(image, outImagePath)
	if err != nil {
		log.Error("error exporting map image: ", err)
		os.Exit(1)
	}

	log.Info("map image exported: " + outImagePath)
	os.Exit(0)
}
