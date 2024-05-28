package game

import (
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
)

type LaunchMenu struct {
	*MenuModel
	content *widget.Container
}

func createLaunchMenu(g *Game) *LaunchMenu {
	var ui *ebitenui.UI = &ebitenui.UI{}

	menu := &LaunchMenu{
		MenuModel: &MenuModel{
			game:   g,
			ui:     ui,
			active: true,
		},
	}

	menu.initResources()
	menu.initMenu()

	return menu
}

func (m *LaunchMenu) initMenu() {
	m.MenuModel.initMenu()
	m.root.BackgroundImage = m.Resources().background

	// menu title
	titleBar := launchTitleContainer(m)
	m.root.AddChild(titleBar)

	// pre-launch briefing
	m.content = launchMenuBriefingContainer(m)
	m.root.AddChild(m.content)

	// footer
	footer := launchMenuFooterContainer(m)
	m.root.AddChild(footer)
}

func (m *LaunchMenu) Update() {
	m.ui.Update()
}

func (m *LaunchMenu) Draw(screen *ebiten.Image) {
	m.ui.Draw(screen)
}

func launchTitleContainer(m *LaunchMenu) *widget.Container {
	res := m.Resources()

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:   m.Padding(),
				Right:  m.Padding(),
				Top:    m.Padding(),
				Bottom: m.Padding(),
			}))))

	c.AddChild(widget.NewText(
		widget.TextOpts.Text("Mission Launchpad", res.text.bigTitleFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	return c
}

func launchMenuFooterContainer(m *LaunchMenu) *widget.Container {
	game := m.Game()
	res := m.Resources()

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(3),
			widget.GridLayoutOpts.Stretch([]bool{false, true, false}, []bool{false}),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:   m.Padding(),
				Right:  m.Padding(),
				Top:    m.Padding(),
				Bottom: m.Padding(),
			}))))

	back := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Back", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if iScene, ok := game.scene.(*InstantActionScene); ok {
				iScene.back()
			}
		}),
	)
	c.AddChild(back)

	c.AddChild(newBlankSeparator(m, widget.RowLayoutData{
		Stretch: true,
	}))

	next := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Launch", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if iScene, ok := game.scene.(*InstantActionScene); ok {
				iScene.next()
			}
		}),
	)
	c.AddChild(next)

	return c
}

func launchMenuBriefingContainer(m *LaunchMenu) *widget.Container {
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:  m.Spacing(),
				Right: m.Spacing(),
			}),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(m.Spacing(), 0),
		)))

	return c
}

func (m *LaunchMenu) loadBriefing() {
	m.content.RemoveChildren()
	res := m.Resources()
	g := m.game

	// show mission card
	missionCard := createMissionCard(g, res, g.mission, MissionCardLaunch)
	m.content.AddChild(missionCard)

	// show player unit card
	var playerUnit model.Unit
	if g.player != nil {
		playerUnit = g.player.Unit
	}
	unitCard := createUnitCard(g, res, playerUnit, UnitCardLaunch)
	m.content.AddChild(unitCard)
}
