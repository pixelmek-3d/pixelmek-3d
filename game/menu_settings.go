package game

import (
	"github.com/ebitenui/ebitenui"
	"github.com/hajimehoshi/ebiten/v2"
)

type SettingsMenu struct {
	*MenuModel
	preSelectedPage int
}

func createSettingsMenu(g *Game) *SettingsMenu {
	var ui *ebitenui.UI = &ebitenui.UI{}

	menu := &SettingsMenu{
		MenuModel: &MenuModel{
			game:        g,
			ui:          ui,
			active:      true,
			fontScale:   1.0,
			resolutions: generateMenuResolutions(),
		},
	}

	menu.initResources()
	menu.initMenu()

	return menu
}

func (m *SettingsMenu) initMenu() {
	m.MenuModel.initMenu()

	// menu title
	titleBar := settingsTitleContainer(m)
	m.root.AddChild(titleBar)

	// settings pages
	items := settingsContainer(m)
	m.root.AddChild(items)
}

func (m *SettingsMenu) Update() {
	m.ui.Update()
}

func (m *SettingsMenu) Draw(screen *ebiten.Image) {
	m.ui.Draw(screen)
}
