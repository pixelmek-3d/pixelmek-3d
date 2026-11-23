package unit

import (
	"image"
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
	animationCmd.Flags().StringVarP(&outImagePath, "output", "o", "", "[required] output animated image path (gif, apng)")
	animationCmd.MarkFlagRequired("output")

	animationCmd.Flags().IntVarP(&animationIndex, "index", "i", 0, "unit animation index")

	animationCmd.Flags().StringVarP(&camoMaskPath, "camo", "c", "", "camo mask image resource path")
}

var (
	g              *game.Game
	animationIndex int
	outImagePath   string
	camoMaskPath   string
	animationCmd   = &cobra.Command{
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

	var frames []*ebiten.Image
	var bounds image.Rectangle
	var msFrameDelay int = 200

	opts := game.UnitSpriteOptions{Camo: camoMaskPath}

	switch unitTypePrefix {
	case model.MechResourceType:
		if animationIndex >= int(sprites.NUM_MECH_ANIMATIONS) {
			log.Errorf("input animation index (%d) is higher than the number of animations present (%d)", animationIndex, sprites.NUM_MECH_ANIMATIONS)
			os.Exit(1)
		}

		r := &model.ModelMechResource{}
		err = yaml.Unmarshal(unitYaml, r)
		if err != nil {
			log.Error("error loading unit resource", unitFile, err.Error())
		}

		unit := model.NewMech(r)
		sprite := g.CreateUnitSprite(unit, opts).(*sprites.MechSprite)
		bounds = sprite.Texture().Bounds()

		msFrameDelay = 20 // shorter frame delay for mech animations
		sprite.SetMechAnimation(sprites.MechAnimationIndex(animationIndex), false)

		// collect sprite animation frames
		for sprite.LoopCounter() < 1 {
			frames = append(frames, sprite.Texture())
			sprite.Update(nil)
		}

	case model.VehicleResourceType:
		r := &model.ModelVehicleResource{}
		err = yaml.Unmarshal(unitYaml, r)
		if err != nil {
			log.Error("error loading unit resource", unitFile, err.Error())
		}

		unit := model.NewVehicle(r)
		sprite := g.CreateUnitSprite(unit, opts).(*sprites.VehicleSprite)
		bounds = sprite.Texture().Bounds()

		// collect sprite animation frames
		for sprite.LoopCounter() < 1 {
			frames = append(frames, sprite.Texture())
			sprite.Update(nil)
		}

	case model.VTOLResourceType:
		r := &model.ModelVTOLResource{}
		err = yaml.Unmarshal(unitYaml, r)
		if err != nil {
			log.Error("error loading unit resource", unitFile, err.Error())
		}

		unit := model.NewVTOL(r)
		sprite := g.CreateUnitSprite(unit, opts).(*sprites.VTOLSprite)
		bounds = sprite.Texture().Bounds()

		// collect sprite animation frames
		for sprite.LoopCounter() < 1 {
			frames = append(frames, sprite.Texture())
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

	err = render.SaveAnimation(frames, bounds, outImagePath, msFrameDelay)
	if err != nil {
		log.Error("error exporting unit animation: ", err)
		os.Exit(1)
	}

	log.Info("unit animation exported: ", outImagePath)
	os.Exit(0)
}
