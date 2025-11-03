package export

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	exportScreenWidth  = 480
	exportScreenHeight = 25
)

func NewExportLoop(exportFunc func()) ebiten.Game {
	ebiten.SetFullscreen(false)
	ebiten.SetWindowSize(exportScreenWidth, exportScreenHeight)
	return &exportLoop{
		exportFunc: exportFunc,
	}
}

type exportLoop struct {
	exportFunc    func()
	exportRunning bool
	exportCounter int
}

func (g *exportLoop) Update() error {
	if !g.exportRunning {
		go g.exportFunc()
		g.exportCounter = 1
		g.exportRunning = true
	} else {
		g.exportCounter++
	}

	if g.exportCounter > ebiten.TPS() {
		g.exportCounter = 1
	}
	return nil
}

func (g *exportLoop) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, strings.Repeat(".", g.exportCounter))
}

func (g *exportLoop) Layout(outsideWidth, outsideHeight int) (int, int) {
	return exportScreenWidth, exportScreenHeight
}
