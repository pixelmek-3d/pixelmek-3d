package game

import (
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
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
			fontScale: 1.5,
		},
	}

	menu.initResources()
	menu.initMenu()

	return menu
}

func (m *MainMenu) initMenu() {
	m.MenuModel.initMenu()
	m.root.SetBackgroundImage(m.Resources().background)

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

func mainMenuTitleContainer(m *MainMenu) *widget.Container {
	res := m.Resources()

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Left:   m.Padding(),
				Right:  m.Padding(),
				Top:    m.Padding(),
				Bottom: m.Padding(),
			}))))

	c.AddChild(widget.NewText(
		widget.TextOpts.Text(title, res.text.bigTitleFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	return c
}

func mainMenuItemsContainer(m *MainMenu) *widget.Container {
	res := m.Resources()
	game := m.Game()

	c := newPageContentContainer()

	missions := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Missions", res.text.titleFace, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			game.scene = NewMissionScene(game)
		}),
	)
	c.AddChild(missions)

	settings := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Settings", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			mScene, ok := game.scene.(*MenuScene)
			if ok {
				mScene.SetMenu(mScene.settings)
			}
		}),
	)
	c.AddChild(settings)

	if game.osType == osTypeBrowser {
		// exit in browser kills but freezes the application, users can just close the tab/window
	} else {
		// show in game exit button
		c.AddChild(newSeparator(m, widget.RowLayoutData{
			Stretch: true,
		}))

		// TODO: add pop up to confirm exit to main menu or exit application

		exit := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			})),
			widget.ButtonOpts.Image(res.button.image),
			widget.ButtonOpts.Text("Exit", res.button.face, res.button.text),
			widget.ButtonOpts.TextPadding(res.button.padding),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				openExitWindow(m)
			}),
		)
		c.AddChild(exit)
	}

	return c
}

func mainMenuFooterContainer(m *MainMenu) *widget.Container {
	res := m.Resources()

	c := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Padding(&widget.Insets{
			Left:  m.Spacing(),
			Right: m.Spacing(),
		}),
	)))
	c.AddChild(widget.NewText(
		widget.TextOpts.Text("github.com/pixelmek-3d/pixelmek-3d", res.text.smallFace, res.text.disabledColor)))
	return c
}
