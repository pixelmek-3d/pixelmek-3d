package game

import (
	"fmt"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
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

func generateMenuResolutions() []MenuResolution {
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

func settingsTitleContainer(m Menu) *widget.Container {
	res := m.Resources()

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true}),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:   m.Padding(),
				Right:  m.Padding(),
				Top:    m.Padding(),
				Bottom: m.Padding(),
			}))))

	c.AddChild(widget.NewText(
		widget.TextOpts.Text("PixelMek 3D Settings", res.text.titleFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	c.AddChild(widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("X", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			m.Game().closeMenu()
		}),
		widget.ButtonOpts.TabOrder(99),
	))

	return c
}

func settingsContainer(m Menu) widget.PreferredSizeLocateableWidget {
	res := m.Resources()

	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:  m.Spacing(),
				Right: m.Spacing(),
			}),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(m.Spacing(), 0),
		)))

	gameMenu, _ := m.(*GameMenu)
	settingsMenu, _ := m.(*SettingsMenu)

	var gameSettings *page
	if gameMenu != nil {
		// only show the Resume/Exit buttons page in-mission
		gameSettings = gamePage(m)
	}

	displaySettings := displayPage(m)
	renderSettings := renderPage(m)
	hudSettings := hudPage(m)

	pages := make([]interface{}, 0, 8)
	if gameSettings != nil {
		pages = append(pages, gameSettings)
	}
	pages = append(pages, displaySettings)
	pages = append(pages, renderSettings)
	pages = append(pages, hudSettings)

	var lightingSettings *page
	if m.Game().debug {
		lightingSettings = lightingPage(m)
		pages = append(pages, lightingSettings)
	}

	pageContainer := newPageContainer(res)

	pageList := widget.NewList(
		widget.ListOpts.Entries(pages),
		widget.ListOpts.EntryLabelFunc(func(e interface{}) string {
			return e.(*page).title
		}),
		widget.ListOpts.ScrollContainerOpts(widget.ScrollContainerOpts.Image(res.list.image)),
		widget.ListOpts.SliderOpts(
			widget.SliderOpts.Images(res.list.track, res.list.handle),
			widget.SliderOpts.MinHandleSize(res.list.handleSize),
			widget.SliderOpts.TrackPadding(res.list.trackPadding),
		),
		widget.ListOpts.EntryColor(res.list.entry),
		widget.ListOpts.EntryFontFace(res.list.face),
		widget.ListOpts.EntryTextPadding(res.list.entryPadding),
		widget.ListOpts.HideHorizontalSlider(),

		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			nextPage := args.Entry.(*page)
			pageContainer.setPage(nextPage)
			if gameSettings != nil && (nextPage == hudSettings || (lightingSettings != nil && nextPage == lightingSettings)) {
				// for in-game HUD and lighting setting, apply custom background so can see behind while adjusting
				m.Root().BackgroundImage = nil
				pageContainer.widget.(*widget.Container).BackgroundImage = nil
				nextPage.content.(*widget.Container).BackgroundImage = res.panel.filled
			} else {
				m.Root().BackgroundImage = res.background
				pageContainer.widget.(*widget.Container).BackgroundImage = res.panel.image
			}
			m.Root().RequestRelayout()
		}))
	c.AddChild(pageList)

	c.AddChild(pageContainer.widget)

	pageList.SetSelectedEntry(pages[0])

	switch {
	case gameMenu != nil && gameMenu.preSelectedPage > 0:
		pageList.SetSelectedEntry(pages[gameMenu.preSelectedPage])
		// reset pre-selected page selection
		gameMenu.preSelectedPage = 0
	case settingsMenu != nil && settingsMenu.preSelectedPage > 0:
		pageList.SetSelectedEntry(pages[settingsMenu.preSelectedPage])
		// reset pre-selected page selection
		settingsMenu.preSelectedPage = 0
	}

	return c
}
