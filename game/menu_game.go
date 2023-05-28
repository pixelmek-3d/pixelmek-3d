package game

import (
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

	// menu title
	titleBar := settingsTitleContainer(m)
	m.root.AddChild(titleBar)

	// settings pages
	settings := settingsContainer(m)
	m.root.AddChild(settings)
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
