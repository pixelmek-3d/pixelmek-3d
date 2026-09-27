package game

import (
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
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

	// load font
	titleFace, err := resources.LoadFont(fontFaceTitle, 32.0*m.dynamicFontScale)
	if err != nil {
		panic(err)
	}

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Left:   24,
				Right:  0,
				Top:    0,
				Bottom: 0,
			}))))

	c.AddChild(widget.NewText(
		widget.TextOpts.Text(title, &titleFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	))

	return c
}

func mainMenuItemsContainer(m *MainMenu) *widget.Container {
	res := m.Resources()
	game := m.Game()

	c := newPageContentContainer()

	instant := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: false,
		})),
		widget.ButtonOpts.Image(res.darkButton.image),
		widget.ButtonOpts.Text("Instant Action", res.text.titleFace, res.darkButton.text),
		widget.ButtonOpts.TextPadding(res.darkButton.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			game.scene = NewInstantActionScene(game)
		}),
	)
	c.AddChild(instant)

	missions := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: false,
		})),
		widget.ButtonOpts.Image(res.darkButton.image),
		widget.ButtonOpts.Text("Missions", res.text.titleFace, res.darkButton.text),
		widget.ButtonOpts.TextPadding(res.darkButton.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			game.scene = NewMissionScene(game)
		}),
	)
	c.AddChild(missions)

	settings := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: false,
		})),
		widget.ButtonOpts.Image(res.darkButton.image),
		widget.ButtonOpts.Text("Settings", res.darkButton.face, res.darkButton.text),
		widget.ButtonOpts.TextPadding(res.darkButton.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			mScene, ok := game.scene.(*MainMenuScene)
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
		c.AddChild(newBlankSeparator(m.Resources(), m.Spacing(), widget.RowLayoutData{
			Stretch: true,
		}))

		exit := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: false,
			})),
			widget.ButtonOpts.Image(res.darkButton.image),
			widget.ButtonOpts.Text("Exit", res.darkButton.face, res.darkButton.text),
			widget.ButtonOpts.TextPadding(res.darkButton.padding),
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

	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
		)))

	c.AddChild(widget.NewText(
		widget.TextOpts.Text("github.com/pixelmek-3d", res.text.smallFace, color.Black),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	))
	return c
}
