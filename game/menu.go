package game

import (
	"fmt"
	"math"
	"os"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
	log "github.com/sirupsen/logrus"
)

type GameMenu struct {
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

	resolutions     []MenuResolution
	preSelectedPage int
}

type MenuResolution struct {
	width, height int
	aspectRatio   MenuAspectRatio
}

type MenuAspectRatio struct {
	w, h, fov int
}

func (r MenuResolution) String() string {
	if r.aspectRatio.w == 0 || r.aspectRatio.h == 0 {
		return fmt.Sprintf("(*) %dx%d", r.width, r.height)
	}
	return fmt.Sprintf("(%d:%d) %dx%d", r.aspectRatio.w, r.aspectRatio.h, r.width, r.height)
}

func createMenu(g *Game) *GameMenu {
	var ui *ebitenui.UI = &ebitenui.UI{}

	menu := &GameMenu{
		game:        g,
		ui:          ui,
		active:      false,
		fontScale:   1.0,
		resolutions: g.generateMenuResolutions(),
	}

	menu.initResources()
	menu.initMenu()

	return menu
}

func (m *GameMenu) initResources() {
	// adjust menu and resource sizes based on window size
	minMenuAspectRatio, maxMenuAspectRatio := 1.0, 1.5
	screenW, screenH := float64(m.game.screenWidth), float64(m.game.screenHeight)
	screenAspectRatio := screenW / screenH

	var paddingX, paddingY, menuWidth, menuHeight int

	if screenAspectRatio > maxMenuAspectRatio {
		// ultra-wide aspect, constrict HUD width based on screen height
		paddingY = int(screenH * 0.02)
		menuHeight = int(screenH) - paddingY*2

		menuWidth = int(screenH * maxMenuAspectRatio)
		//paddingX = menuWidth * 0.02
	} else if screenAspectRatio < minMenuAspectRatio {
		// tall vertical aspect, constrict HUD height based on screen width
		paddingX = int(screenW * 0.02)
		menuWidth = int(screenW) - paddingX*2

		menuHeight = int(screenW / minMenuAspectRatio)
		//paddingY = menuHeight * 0.02
	} else {
		// use current aspect ratio
		paddingX, paddingY = int(screenW*0.02), int(screenH*0.02)
		menuWidth, menuHeight = int(screenW)-paddingX*2, int(screenH)-paddingY*2
	}

	menuSize := menuHeight
	if menuWidth < menuHeight {
		menuSize = menuWidth
	}

	m.fontScale = geom.Clamp(float64(menuSize)*0.002, 0.5, 2.0)

	m.marginX = (m.game.screenWidth - menuWidth) / 2
	m.marginY = (m.game.screenHeight - menuHeight) / 2
	m.spacing = int(float64(menuSize) * 0.02)
	m.padding = 4

	res, err := NewUIResources(m)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	m.res = res
}

func (m *GameMenu) initMenu() {
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
		widget.ContainerOpts.BackgroundImage(m.res.background),
	)

	// window title
	titleBar := titleBarContainer(m)
	m.root.AddChild(titleBar)

	// settings pages
	settings := settingsContainer(m)
	m.root.AddChild(settings)

	// footer
	footer := footerContainer(m)
	m.root.AddChild(footer)

	m.ui.Container = m.root
}

func (g *Game) generateMenuResolutions() []MenuResolution {
	resolutions := make([]MenuResolution, 0)

	ratios := []MenuAspectRatio{
		{5, 4, 64},
		{4, 3, 68},
		{3, 2, 74},
		{16, 9, 84},
		{21, 9, 100},
	}

	widths := []int{
		640,
		800,
		960,
		1024,
		1280,
		1440,
		1600,
		1920,
	}

	for _, r := range ratios {
		for _, w := range widths {
			h := (w / r.w) * r.h
			resolutions = append(
				resolutions,
				MenuResolution{width: w, height: h, aspectRatio: r},
			)
		}
	}

	return resolutions
}

func (g *Game) openMenu() {
	g.paused = true
	g.mouseMode = MouseModeCursor
	ebiten.SetCursorMode(ebiten.CursorModeVisible)

	g.menu.initMenu()
	g.menu.active = true
}

func (g *Game) closeMenu() {
	g.mouseMode = MouseModeTurret
	g.mouseX, g.mouseY = math.MinInt32, math.MinInt32
	g.menu.active = false
	g.menu.closing = true
	g.paused = false
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)
}

func (m *GameMenu) update(g *Game) {
	if !m.active {
		return
	}

	m.ui.Update()
}

func (m *GameMenu) draw(screen *ebiten.Image) {
	if !m.active {
		return
	}

	m.ui.Draw(screen)
}
