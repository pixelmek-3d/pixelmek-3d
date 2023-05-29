package game

import (
	"fmt"
	"image/color"
	"os"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

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

func mainMenuTitleContainer(m Menu) *widget.Container {
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
		widget.TextOpts.Text(title, res.text.bigTitleFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	return c
}

func mainMenuItemsContainer(m *MainMenu) *widget.Container {
	res := m.Resources()
	game := m.Game()

	c := newPageContentContainer()

	instantAction := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Instant Action", res.text.titleFace, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			game.scene = NewMissionScene(game)
		}),
	)
	c.AddChild(instantAction)

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

func mainMenuFooterContainer(m Menu) *widget.Container {
	res := m.Resources()

	c := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Padding(widget.Insets{
			Left:  m.Spacing(),
			Right: m.Spacing(),
		}),
	)))
	c.AddChild(widget.NewText(
		widget.TextOpts.Text("github.com/harbdog/pixelmek-3d", res.text.smallFace, res.text.disabledColor)))
	return c
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

func newCheckbox(m Menu, label string, checked bool, changedHandler widget.CheckboxChangedHandlerFunc) *widget.LabeledCheckbox {
	res := m.Resources()
	c := widget.NewLabeledCheckbox(
		widget.LabeledCheckboxOpts.Spacing(res.checkbox.spacing),
		widget.LabeledCheckboxOpts.CheckboxOpts(
			widget.CheckboxOpts.ButtonOpts(widget.ButtonOpts.Image(res.checkbox.image)),
			widget.CheckboxOpts.Image(res.checkbox.graphic),
			widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
				if changedHandler != nil {
					changedHandler(args)
				}
			})),
		widget.LabeledCheckboxOpts.LabelOpts(widget.LabelOpts.Text(label, res.label.face, res.label.text)))

	if checked {
		c.SetState(widget.WidgetChecked)
	}

	return c
}

func newListComboButton(entries []interface{}, selectedEntry interface{}, buttonLabel widget.SelectComboButtonEntryLabelFunc, entryLabel widget.ListEntryLabelFunc,
	entrySelectedHandler widget.ListComboButtonEntrySelectedHandlerFunc, res *uiResources) *widget.ListComboButton {

	c := widget.NewListComboButton(
		widget.ListComboButtonOpts.SelectComboButtonOpts(
			widget.SelectComboButtonOpts.ComboButtonOpts(
				widget.ComboButtonOpts.ButtonOpts(
					widget.ButtonOpts.Image(res.comboButton.image),
					widget.ButtonOpts.TextPadding(res.comboButton.padding),
				),
			),
		),
		widget.ListComboButtonOpts.Text(res.comboButton.face, res.comboButton.graphic, res.comboButton.text),
		widget.ListComboButtonOpts.ListOpts(
			widget.ListOpts.Entries(entries),
			widget.ListOpts.ScrollContainerOpts(
				widget.ScrollContainerOpts.Image(res.list.image),
			),
			widget.ListOpts.SliderOpts(
				widget.SliderOpts.Images(res.list.track, res.list.handle),
				widget.SliderOpts.MinHandleSize(res.list.handleSize),
				widget.SliderOpts.TrackPadding(res.list.trackPadding)),
			widget.ListOpts.EntryFontFace(res.list.face),
			widget.ListOpts.EntryColor(res.list.entry),
			widget.ListOpts.EntryTextPadding(res.list.entryPadding),
		),
		widget.ListComboButtonOpts.EntryLabelFunc(buttonLabel, entryLabel),
		widget.ListComboButtonOpts.EntrySelectedHandler(entrySelectedHandler))

	if selectedEntry != nil {
		c.SetSelectedEntry(selectedEntry)
	}

	return c
}

func newColorPickerRGB(m Menu, label string, clr *color.NRGBA, f widget.SliderChangedHandlerFunc) *widget.Container {
	// create custom RGB selection group container
	res := m.Resources()
	padding := m.Padding()

	picker := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(4),
			widget.GridLayoutOpts.Stretch([]bool{true, true, true, true}, nil),
			widget.GridLayoutOpts.Spacing(padding, padding))))

	pickerLabel := widget.NewLabel(widget.LabelOpts.Text(label, res.label.face, res.label.text))
	var rText, gText, bText *widget.Label
	var rgbValue *widget.Container

	rSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(50, 6)),
		widget.SliderOpts.MinMax(0, 255),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			rText.Label = fmt.Sprintf("R: %d", args.Current)
			clr.R = uint8(args.Current)
			rgbValue.BackgroundImage = image.NewNineSliceColor(*clr)
		}),
		widget.SliderOpts.ChangedHandler(f),
	)
	rSlider.Current = int(clr.R)

	gSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(50, 6)),
		widget.SliderOpts.MinMax(0, 255),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			gText.Label = fmt.Sprintf("G: %d", args.Current)
			clr.G = uint8(args.Current)
			rgbValue.BackgroundImage = image.NewNineSliceColor(*clr)
		}),
		widget.SliderOpts.ChangedHandler(f),
	)
	gSlider.Current = int(clr.G)

	bSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(50, 6)),
		widget.SliderOpts.MinMax(0, 255),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			bText.Label = fmt.Sprintf("B: %d", args.Current)
			clr.B = uint8(args.Current)
			rgbValue.BackgroundImage = image.NewNineSliceColor(*clr)
		}),
		widget.SliderOpts.ChangedHandler(f),
	)
	bSlider.Current = int(clr.B)

	rText = widget.NewLabel(widget.LabelOpts.Text(fmt.Sprintf("R: %d", rSlider.Current), res.label.face, res.label.text))
	gText = widget.NewLabel(widget.LabelOpts.Text(fmt.Sprintf("G: %d", gSlider.Current), res.label.face, res.label.text))
	bText = widget.NewLabel(widget.LabelOpts.Text(fmt.Sprintf("B: %d", bSlider.Current), res.label.face, res.label.text))

	rgbBackground := image.NewNineSliceColor(clr)
	rgbValue = widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: false,
		})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Padding(widget.Insets{
			Top:    padding,
			Bottom: padding,
			Left:   padding,
			Right:  padding,
		}))),
		widget.ContainerOpts.BackgroundImage(rgbBackground),
	)

	// RGB row 1: labels
	picker.AddChild(pickerLabel)
	picker.AddChild(rText)
	picker.AddChild(gText)
	picker.AddChild(bText)

	// RGB row 2: color swatch and sliders
	picker.AddChild(rgbValue)
	picker.AddChild(rSlider)
	picker.AddChild(gSlider)
	picker.AddChild(bSlider)

	return picker
}

func newSeparator(m Menu, ld interface{}) widget.PreferredSizeLocateableWidget {
	res := m.Resources()
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Top:    m.Spacing(),
				Bottom: m.Spacing(),
			}))),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(ld)))

	c.AddChild(widget.NewGraphic(
		widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch:   true,
			MaxHeight: 2,
		})),
		widget.GraphicOpts.ImageNineSlice(image.NewNineSliceColor(res.separatorColor)),
	))

	return c
}

func openExitWindow(m Menu) {
	var rmWindow widget.RemoveWindowFunc
	var window *widget.Window

	game := m.Game()
	uiRect := game.uiRect()
	res := m.Resources()
	padding := m.Padding()
	spacing := m.Spacing()

	showLeave := false
	showExit := false

	if game.osType == osTypeBrowser {
		// only show leave battle button in browser
		// exit in browser kills but freezes the application, users can just close the tab/window
	} else {
		showExit = true
	}

	_, isGameMenu := m.(*GameMenu)
	if isGameMenu {
		// leave battle only applicable in mission/game
		showLeave = true
	}

	titleBar := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true}), widget.GridLayoutOpts.Padding(widget.Insets{
			Left:   padding,
			Right:  padding,
			Top:    padding,
			Bottom: padding,
		}))))

	titleBar.AddChild(widget.NewText(
		widget.TextOpts.Text("Embrace Cowardice?", res.text.titleFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	titleBar.AddChild(widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("X", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			rmWindow()
		}),
		widget.ButtonOpts.TabOrder(99),
	))

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.image),
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(1),
				widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, false, true}),
				widget.GridLayoutOpts.Padding(res.panel.padding),
				widget.GridLayoutOpts.Spacing(1, spacing),
			),
		),
	)

	cancel := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("Cancel", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			rmWindow()
		}),
	)
	c.AddChild(cancel)

	c.AddChild(newSeparator(m, widget.RowLayoutData{
		Stretch: true,
	}))

	numExitOptions := 0
	if showLeave {
		numExitOptions += 1
	}
	if showExit {
		numExitOptions += 1
	}

	if numExitOptions > 0 {
		bc := widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(res.panel.image),
			widget.ContainerOpts.Layout(
				widget.NewGridLayout(
					widget.GridLayoutOpts.Columns(numExitOptions),
					widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{true}),
					widget.GridLayoutOpts.Padding(res.panel.padding),
					widget.GridLayoutOpts.Spacing(1, spacing),
				),
			),
		)
		c.AddChild(bc)

		if showLeave {
			leave := widget.NewButton(
				widget.ButtonOpts.Image(res.button.image),
				widget.ButtonOpts.TextPadding(res.button.padding),
				widget.ButtonOpts.Text("Leave Battle", res.button.face, res.button.text),
				widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
					// go back to main menu
					game.scene = NewMenuScene(game)
				}),
			)
			bc.AddChild(leave)
		}

		if showExit {
			exit := widget.NewButton(
				widget.ButtonOpts.Image(res.button.image),
				widget.ButtonOpts.TextPadding(res.button.padding),
				widget.ButtonOpts.Text("Exit Game", res.button.face, res.button.text),
				widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
					os.Exit(0)
				}),
			)
			bc.AddChild(exit)
		}
	}

	window = widget.NewWindow(
		widget.WindowOpts.Modal(),
		widget.WindowOpts.Contents(c),
		widget.WindowOpts.TitleBar(titleBar, uiRect.Dy()/12),
	)

	wRect := uiRect.Inset(uiRect.Dy() / 6)
	window.SetLocation(wRect)

	rmWindow = m.UI().AddWindow(window)
}
