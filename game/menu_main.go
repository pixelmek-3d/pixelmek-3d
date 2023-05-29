package game

import (
	"github.com/ebitenui/ebitenui"
	"github.com/hajimehoshi/ebiten/v2"
)

type MainMenu struct {
	*MenuModel
}

func createMainMenu(g *Game) *MainMenu {
	var ui *ebitenui.UI = &ebitenui.UI{}

	menu := &MainMenu{
		MenuModel: &MenuModel{
			game:      g,
			ui:        ui,
			active:    true,
			fontScale: 1.0,
		},
	}

	menu.initResources()
	menu.initMenu()

	return menu
}

func (m *MainMenu) initMenu() {
	m.MenuModel.initMenu()

	m.root.BackgroundImage = m.Resources().background

	// menu title
	titleBar := mainMenuTitleContainer(m)
	m.root.AddChild(titleBar)

	// main menu items
	items := mainMenuItemsContainer(m)
	m.root.AddChild(items)

	// footer
	footer := mainMenuFooterContainer(m)
	m.root.AddChild(footer)
}

func (m *MainMenu) Update() {
	m.ui.Update()
}

func (m *MainMenu) Draw(screen *ebiten.Image) {
	m.ui.Draw(screen)
}
