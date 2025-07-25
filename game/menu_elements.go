package game

import (
	"fmt"
	"image/color"
	"os"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

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

func newListComboButton(entries []any, selectedEntry any, buttonLabel widget.SelectComboButtonEntryLabelFunc, entryLabel widget.ListEntryLabelFunc,
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

func newTextArea(text string, res *uiResources, widgetOpts ...widget.WidgetOpt) *widget.TextArea {
	return widget.NewTextArea(
		widget.TextAreaOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widgetOpts...)),
		widget.TextAreaOpts.ScrollContainerOpts(widget.ScrollContainerOpts.Image(res.list.image)),
		widget.TextAreaOpts.SliderOpts(
			widget.SliderOpts.Images(res.list.track, res.list.handle),
			widget.SliderOpts.MinHandleSize(res.list.handleSize),
			widget.SliderOpts.TrackPadding(res.list.trackPadding),
		),
		widget.TextAreaOpts.ShowVerticalScrollbar(),
		// widget.TextAreaOpts.VerticalScrollMode(widget.PositionAtEnd),
		widget.TextAreaOpts.ProcessBBCode(true),
		widget.TextAreaOpts.FontFace(res.textArea.face),
		widget.TextAreaOpts.FontColor(color.NRGBA{R: 200, G: 100, B: 0, A: 255}),
		widget.TextAreaOpts.TextPadding(res.textArea.entryPadding),
		widget.TextAreaOpts.Text(text),
	)
}

func newSeparator(m Menu, ld any) widget.PreferredSizeLocateableWidget {
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

func newBlankSeparator(m Menu, ld any) widget.PreferredSizeLocateableWidget {
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
			MaxHeight: 4,
		})),
		widget.GraphicOpts.ImageNineSlice(image.NewNineSliceColor(res.backgroundColor)),
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
					// save config now in case settings changes were made
					game.saveConfig()

					if game.InProgress() && game.player.ejectionPod == nil {
						// destroy player to make them eject
						destroyEntity(game.player)
						game.closeMenu()
					} else {
						game.LeaveGame()
					}
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
					// save config now in case settings changes were made
					game.saveConfig()
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
	m.SetWindow(window)
}
