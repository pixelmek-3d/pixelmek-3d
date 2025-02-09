package game

import (
	"fmt"

	"github.com/ebitenui/ebitenui/widget"
)

func debugOptionsPage(m Menu) *settingsPage {
	c := newPageContentContainer()
	res := m.Resources()
	game := m.Game()

	// raycaster lighting options for debug mode only
	debugLabel := widget.NewLabel(widget.LabelOpts.Text("~Debug Mode Only", res.label.face, res.label.text))
	c.AddChild(debugLabel)
	c.AddChild(newSeparator(m, widget.RowLayoutData{
		Stretch: true,
	}))

	// AI ignore player checkbox
	ignorePlayerCheckbox := newCheckbox(m, "AI Ignore Player", game.aiIgnorePlayer, func(args *widget.CheckboxChangedEventArgs) {
		game.aiIgnorePlayer = args.State == widget.WidgetChecked
	})
	c.AddChild(ignorePlayerCheckbox)

	return &settingsPage{
		title:   "~Options",
		content: c,
	}
}

func debugLightingPage(m Menu) *settingsPage {
	c := newPageContentContainer()
	res := m.Resources()
	game := m.Game()

	// raycaster lighting options for debug mode only
	debugLabel := widget.NewLabel(widget.LabelOpts.Text("~Debug Mode Only", res.label.face, res.label.text))
	c.AddChild(debugLabel)
	c.AddChild(newSeparator(m, widget.RowLayoutData{
		Stretch: true,
	}))

	// light falloff slider
	falloffRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(falloffRow)

	falloffLabel := widget.NewLabel(widget.LabelOpts.Text("Light Falloff", res.label.face, res.label.text))
	falloffRow.AddChild(falloffLabel)

	var falloffValueText *widget.Label

	falloffSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(100, 6)),
		widget.SliderOpts.MinMax(-500, 500),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			falloffValueText.Label = fmt.Sprintf("%d", args.Current)
			game.setLightFalloff(float64(args.Current))
		}),
	)
	falloffSlider.Current = int(game.lightFalloff)
	falloffRow.AddChild(falloffSlider)

	falloffValueText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", falloffSlider.Current), res.label.face, res.label.text),
	)
	falloffRow.AddChild(falloffValueText)

	// global illumination slider
	globalRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(globalRow)

	globalLabel := widget.NewLabel(widget.LabelOpts.Text("Illumination", res.label.face, res.label.text))
	globalRow.AddChild(globalLabel)

	var globalValueText *widget.Label

	globalSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(100, 6)),
		widget.SliderOpts.MinMax(0, 1000),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			globalValueText.Label = fmt.Sprintf("%d", args.Current)
			game.setGlobalIllumination(float64(args.Current))
		}),
	)
	globalSlider.Current = int(game.globalIllumination)
	globalRow.AddChild(globalSlider)

	globalValueText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", globalSlider.Current), res.label.face, res.label.text),
	)
	globalRow.AddChild(globalValueText)

	// min lighting RGB selection
	pickerMinRGB := newColorPickerRGB(m, "Min Light", game.minLightRGB, func(args *widget.SliderChangedEventArgs) {
		game.setLightRGB(game.minLightRGB, game.maxLightRGB)
	})
	c.AddChild(pickerMinRGB)

	// max lighting RGB selection
	pickerMaxRGB := newColorPickerRGB(m, "Max Light", game.maxLightRGB, func(args *widget.SliderChangedEventArgs) {
		game.setLightRGB(game.minLightRGB, game.maxLightRGB)
	})
	c.AddChild(pickerMaxRGB)

	return &settingsPage{
		title:   "~Lighting",
		content: c,
	}
}
