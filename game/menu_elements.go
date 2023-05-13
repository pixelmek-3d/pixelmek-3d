package game

import (
	"fmt"
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

func titleBarContainer(m *GameMenu) *widget.Container {
	res := m.res

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true}), widget.GridLayoutOpts.Padding(widget.Insets{
			Left:   m.padding,
			Right:  m.padding,
			Top:    m.padding,
			Bottom: m.padding,
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
			m.game.closeMenu()
		}),
		widget.ButtonOpts.TabOrder(99),
	))

	return c
}

func footerContainer(m *GameMenu) *widget.Container {
	res := m.res

	c := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Padding(widget.Insets{
			Left:  m.spacing,
			Right: m.spacing,
		}),
	)))
	c.AddChild(widget.NewText(
		widget.TextOpts.Text("github.com/harbdog/pixelmek-3d", res.text.smallFace, res.text.disabledColor)))
	return c
}

func settingsContainer(m *GameMenu) widget.PreferredSizeLocateableWidget {
	res := m.res

	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:  m.spacing,
				Right: m.spacing,
			}),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(m.spacing, 0),
		)))

	pages := make([]interface{}, 0, 8)

	pages = append(pages, gamePage(m))
	pages = append(pages, displayPage(m))
	pages = append(pages, renderPage(m))

	if m.game.debug {
		pages = append(pages, lightingPage(m))
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
			pageContainer.setPage(args.Entry.(*page))
			m.root.RequestRelayout()
		}))
	c.AddChild(pageList)

	c.AddChild(pageContainer.widget)

	pageList.SetSelectedEntry(pages[m.preSelectedPage])
	if m.preSelectedPage != 0 {
		// reset pre-selected page selection
		m.preSelectedPage = 0
	}

	return c
}

func newCheckbox(label string, checked bool, changedHandler widget.CheckboxChangedHandlerFunc, res *uiResources) *widget.LabeledCheckbox {
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

func (m *GameMenu) newColorPickerRGB(label string, clr *color.NRGBA, f widget.SliderChangedHandlerFunc) *widget.Container {
	// create custom RGB selection group container
	res := m.res
	picker := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(4),
			widget.GridLayoutOpts.Stretch([]bool{true, true, true, true}, nil),
			widget.GridLayoutOpts.Spacing(m.padding, m.padding))))

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
			Top:    m.padding,
			Bottom: m.padding,
			Left:   m.padding,
			Right:  m.padding,
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

func (m *GameMenu) newSeparator(res *uiResources, ld interface{}) widget.PreferredSizeLocateableWidget {
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Top:    m.spacing,
				Bottom: m.spacing,
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
