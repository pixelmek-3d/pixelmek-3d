package game

import (
	"os"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
	log "github.com/sirupsen/logrus"
)

type Menu interface {
	Active() bool
	Closing() bool
	UI() *ebitenui.UI
	Root() *widget.Container
	Resources() *uiResources
	Game() *Game
	FontScale() float64
	MarginX() int
	MarginY() int
	Padding() int
	Spacing() int
	Resolutions() []MenuResolution
	Update()
	Draw(screen *ebiten.Image)
}

type MenuModel struct {
	active  bool
	closing bool
	ui      *ebitenui.UI
	root    *widget.Container
	res     *uiResources
	game    *Game

	fontScale float64
	marginX   int
	marginY   int
	padding   int
	spacing   int

	resolutions []MenuResolution
}

func (m *MenuModel) Active() bool {
	return m.active
}

func (m *MenuModel) Closing() bool {
	return m.closing
}

func (m *MenuModel) UI() *ebitenui.UI {
	return m.ui
}

func (m *MenuModel) Root() *widget.Container {
	return m.root
}

func (m *MenuModel) Resources() *uiResources {
	return m.res
}

func (m *MenuModel) Game() *Game {
	return m.game
}

func (m *MenuModel) FontScale() float64 {
	return m.fontScale
}

func (m *MenuModel) MarginX() int {
	return m.marginX
}

func (m *MenuModel) MarginY() int {
	return m.marginY
}

func (m *MenuModel) Padding() int {
	return m.padding
}

func (m *MenuModel) Spacing() int {
	return m.spacing
}

func (m *MenuModel) Resolutions() []MenuResolution {
	return m.resolutions
}

func (m *MenuModel) initResources() {
	// adjust menu and resource sizes based on window size
	menuRect := m.game.uiRect()
	menuWidth, menuHeight := menuRect.Dx(), menuRect.Dy()

	menuSize := menuHeight
	if menuWidth < menuHeight {
		menuSize = menuWidth
	}

	m.fontScale = geom.Clamp(float64(menuSize)*0.002, 0.5, 2.0)
	fonts, err := loadFonts(m.fontScale)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	m.marginX = (m.game.screenWidth - menuWidth) / 2
	m.marginY = (m.game.screenHeight - menuHeight) / 2
	m.spacing = int(float64(menuSize) * 0.02)
	m.padding = 4

	res, err := NewUIResources(fonts)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	m.res = res
}

func (m *MenuModel) initMenu() {
	m.root = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			// It is using a GridLayout with a single column
			widget.GridLayoutOpts.Columns(1),
			// It uses the Stretch parameter to define how the rows will be layed out.
			// - a fixed sized header
			// - a content row that stretches to fill all remaining space
			// - a fixed sized footer
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true, false}),
			// Padding defines how much space to put around the outside of the grid.
			widget.GridLayoutOpts.Padding(widget.Insets{
				Top:    m.marginY,
				Bottom: m.marginY,
				Left:   m.marginX,
				Right:  m.marginX,
			}),
			// Spacing defines how much space to put between each column and row
			widget.GridLayoutOpts.Spacing(0, m.spacing))),
		// background image will instead be set based on which page is showing
		//widget.ContainerOpts.BackgroundImage(m.res.background),
	)
	m.ui.Container = m.root
}
