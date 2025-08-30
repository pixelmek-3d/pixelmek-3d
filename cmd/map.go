package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/pixelmek-3d/pixelmek-3d/game"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/mapimage"
	"github.com/pixelmek-3d/pixelmek-3d/game/texture"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	mapExportScreenWidth  = 480
	mapExportScreenHeight = 25
)

func init() {
	rootCmd.AddCommand(mapCmd)

	mapCmd.Flags().StringVarP(&mapFile, "file", "f", "", "map file name")
	mapCmd.MarkFlagRequired("file")

	mapCmd.Flags().StringVarP(&outImagePath, "output", "o", "", "output png image path")
	mapCmd.MarkFlagRequired("output")
}

var (
	exportRunning bool
	exportCounter int
	mapFile       string
	outImagePath  string
	mapCmd        = &cobra.Command{
		Use:   "map",
		Short: "Export map file as an image",
		Run: func(cmd *cobra.Command, args []string) {
			// initialize game resources without running the actual game loop
			game.NewGame()
			// TODO: mute game audio, show progress spinner

			// expand tilde as home directory
			if strings.HasPrefix(outImagePath, "~/") {
				dirname, _ := os.UserHomeDir()
				outImagePath = filepath.Join(dirname, outImagePath[2:])
			}

			// mock game loop required for certain offscreen ebitengine render functions
			ebiten.SetFullscreen(false)
			ebiten.SetWindowSize(mapExportScreenWidth, mapExportScreenHeight)
			ebiten.SetWindowTitle("Exporting map image " + mapFile + " ...")
			if err := ebiten.RunGame(&mapExportLoop{mapFile: mapFile}); err != nil {
				log.Fatal(err)
			}
		},
	}
)

func doMapExport() {
	m, err := model.LoadMap(mapFile)
	if err != nil {
		log.Error("error loading map file: ", mapFile)
		os.Exit(1)
	}

	tex := texture.NewTextureHandler(m)
	image, err := mapimage.NewMapImage(m, tex, 16) // TODO: make pixels per cell configurable with param
	if err != nil {
		log.Error("error creating map image: ", err)
		os.Exit(1)
	}

	err = render.SaveImageAsPNG(image, outImagePath) // TODO: make export file path configurable with param
	if err != nil {
		log.Error("error exporting map image: ", err)
		os.Exit(1)
	}

	log.Info("map image exported: " + outImagePath)
	os.Exit(0)
}

type mapExportLoop struct {
	mapFile string
}

func (g *mapExportLoop) Update() error {
	if !exportRunning {
		go doMapExport()
		exportCounter = 1
		exportRunning = true
	} else {
		exportCounter++
	}

	if exportCounter > ebiten.TPS() {
		exportCounter = 1
	}
	return nil
}

func (g *mapExportLoop) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, strings.Repeat(".", exportCounter))
}

func (g *mapExportLoop) Layout(outsideWidth, outsideHeight int) (int, int) {
	return mapExportScreenWidth, mapExportScreenHeight
}
