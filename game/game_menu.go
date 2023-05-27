package game

import (
	"math"

	"github.com/ebitenui/ebitenui"
	"github.com/hajimehoshi/ebiten/v2"
)

type GameMenu struct {
	*MenuModel
	preSelectedPage int
}

func createGameMenu(g *Game) *GameMenu {
	var ui *ebitenui.UI = &ebitenui.UI{}

	menu := &GameMenu{
		MenuModel: &MenuModel{
			game:        g,
			ui:          ui,
			active:      false,
			fontScale:   1.0,
			resolutions: generateMenuResolutions(),
		},
	}

	menu.initResources()
	menu.initMenu()

	return menu
}

func (m *GameMenu) initMenu() {
	m.MenuModel.initMenu()

	// window title
	titleBar := gameMenuTitleContainer(m)
	m.root.AddChild(titleBar)

	// settings pages
	settings := settingsContainer(m)
	m.root.AddChild(settings)
}

func (g *Game) openMenu() {
	g.paused = true
	g.mouseMode = MouseModeCursor
	ebiten.SetCursorMode(ebiten.CursorModeVisible)

	gMenu, ok := g.menu.(*GameMenu)
	if ok {
		gMenu.initMenu()
		gMenu.active = true
	}
}

func (g *Game) closeMenu() {
	g.mouseMode = MouseModeTurret
	g.mouseX, g.mouseY = math.MinInt32, math.MinInt32

	gMenu, ok := g.menu.(*GameMenu)
	if ok {
		gMenu.active = false
		gMenu.closing = true
	}

	g.paused = false
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)
}

func (m *GameMenu) Update() {
	if !m.active {
		return
	}

	m.ui.Update()
}

func (m *GameMenu) Draw(screen *ebiten.Image) {
	if !m.active {
		return
	}

	m.ui.Draw(screen)
}
