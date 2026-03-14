package unit

import (
	"image"
	"image/color/palette"
	"image/draw"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/export"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/sprites"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	animationCmd.Flags().StringVarP(&outImagePath, "output", "o", "", "[required] output animated gif image path")
}

var (
	g            *game.Game
	outImagePath string
	animationCmd = &cobra.Command{
		Use:   "animation [UNIT_FILE]",
		Short: "Export unit animation image",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			unitFile = args[0]

			// initialize game resources without running the actual game loop
			g = game.NewGame()
			g.Pause()

			// expand tilde as home directory
			if strings.HasPrefix(outImagePath, "~/") {
				dirname, _ := os.UserHomeDir()
				outImagePath = filepath.Join(dirname, outImagePath[2:])
			}

			if err := os.MkdirAll(filepath.Dir(outImagePath), 0755); err != nil {
				log.Fatal(err)
			}

			// mock game loop required for offscreen ebitengine render functions
			ebiten.SetWindowTitle("Exporting unit animation " + unitFile + " ...")
			mapExport := export.NewExportLoop(doUnitExport)
			if err := ebiten.RunGame(mapExport); err != nil {
				log.Fatal(err)
			}
		},
	}
)

func doUnitExport() {
	log.Debug("loading unit file ", unitFile, "...")

	unitTypePrefix := strings.Split(unitFile, "/")[0]
	unitYaml, err := resources.ReadFile(path.Join(model.UnitsResourceType, unitFile))
	if err != nil {
		log.Error("error reading unit file: ", err)
		os.Exit(1)
	}

	var frames []*image.Paletted

	switch unitTypePrefix {
	case model.MechResourceType:
		r := &model.ModelMechResource{}
		err = yaml.Unmarshal(unitYaml, r)
		if err != nil {
			log.Error("error loading unit resource", unitFile, err.Error())
		}

		unit := model.NewMech(r)
		sprite := g.CreateUnitSprite(unit).(*sprites.MechSprite)
		sprite.SetMechAnimation(sprites.MECH_ANIMATE_IDLE, false)
		sprite.ResetAnimation()
		bounds := sprite.Texture().Bounds()

		// collect sprite animation frames
		for sprite.LoopCounter() < 1 {
			currFrame := sprite.Texture()

			// Read pixels from ebiten.Image into a RGBA image
			rgbaImg := image.NewRGBA(bounds)
			currFrame.ReadPixels(rgbaImg.Pix)

			// Create the target paletted image
			pImg := image.NewPaletted(bounds, palette.Plan9)

			// Draw with quantization to map colors
			draw.FloydSteinberg.Draw(pImg, bounds, rgbaImg, image.Point{})

			frames = append(frames, pImg)
			sprite.Update(nil)
		}
	default:
		log.Error("unit type not currently supported: ", unitTypePrefix)
		os.Exit(1)
	}

	if len(frames) == 0 {
		log.Error("no image frames generated")
		os.Exit(1)
	}

	log.Debug("export animation to file...")
	err = render.SaveAnimatedGIF(frames, outImagePath)
	if err != nil {
		log.Error("error exporting unit animation: ", err)
		os.Exit(1)
	}

	log.Info("unit animation exported: ", outImagePath)
	os.Exit(0)
}
