package game

import (
	"fmt"
	"image/color"
	"math"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/harbdog/pixelmek-3d/game/model"
)

type pageContainer struct {
	widget    widget.PreferredSizeLocateableWidget
	titleText *widget.Text
	flipBook  *widget.FlipBook
}

type page struct {
	title   string
	content widget.PreferredSizeLocateableWidget
}

func gamePage(m *GameMenu) *page {
	c := newPageContentContainer()
	res := m.res

	resume := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Resume", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) { m.game.closeMenu() }),
	)
	c.AddChild(resume)

	if m.game.osType == osTypeBrowser {
		// exit in browser kills but freezes the application, users can just close the tab/window
	} else {
		// show in game exit button
		c.AddChild(m.newSeparator(res, widget.RowLayoutData{
			Stretch: true,
		}))

		exit := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			})),
			widget.ButtonOpts.Image(res.button.image),
			widget.ButtonOpts.Text("Exit", res.button.face, res.button.text),
			widget.ButtonOpts.TextPadding(res.button.padding),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) { exit(0) }),
		)
		c.AddChild(exit)
	}

	return &page{
		title:   "Game",
		content: c,
	}
}

func displayPage(m *GameMenu) *page {
	c := newPageContentContainer()
	res := m.res

	// resolution combo box and label
	resolutionRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(resolutionRow)

	resolutionLabel := widget.NewLabel(widget.LabelOpts.Text("Resolution", res.label.face, res.label.text))
	resolutionRow.AddChild(resolutionLabel)

	resolutions := []interface{}{}
	var selectedResolution interface{}
	for _, r := range m.resolutions {
		resolutions = append(resolutions, r)
		if m.game.screenWidth == r.width && m.game.screenHeight == r.height {
			selectedResolution = r
		}
	}

	if selectedResolution == nil {
		// generate custom entry to put at top of the list
		r := MenuResolution{
			width:  m.game.screenWidth,
			height: m.game.screenHeight,
		}
		resolutions = append([]interface{}{r}, resolutions...)
	}

	// TODO: figure out how to make Resolution dropdown snap to currently selected entry instead of first
	var fovSlider *widget.Slider
	resolutionCombo := newListComboButton(
		resolutions,
		selectedResolution,
		func(e interface{}) string {
			return fmt.Sprintf("%s", e)
		},
		func(e interface{}) string {
			return fmt.Sprintf("%s", e)
		},
		func(args *widget.ListComboButtonEntrySelectedEventArgs) {
			r := args.Entry.(MenuResolution)
			if m.game.screenWidth != r.width || m.game.screenHeight != r.height {
				m.game.setResolution(r.width, r.height)

				// pre-select ideal FOV for the aspect ratio
				m.game.setFovAngle(float64(r.aspectRatio.fov))

				// re-initialize the menu with the Display settings pre-selected
				m.preSelectedPage = 1
				m.initResources()
				m.initMenu()
			}
		},
		res)
	resolutionRow.AddChild(resolutionCombo)

	// horizontal FOV slider
	fovRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(fovRow)

	fovLabel := widget.NewLabel(widget.LabelOpts.Text("Horizontal FOV", res.label.face, res.label.text))
	fovRow.AddChild(fovLabel)

	var fovValueText *widget.Label

	fovSlider = widget.NewSlider(
		widget.SliderOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(100, 6),
		),
		widget.SliderOpts.MinMax(60, 120),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			fovValueText.Label = fmt.Sprintf("%d", args.Current)
			m.game.setFovAngle(float64(args.Current))
		}),
	)
	fovSlider.Current = int(m.game.fovDegrees)
	fovRow.AddChild(fovSlider)

	fovValueText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", fovSlider.Current), res.label.face, res.label.text),
	)
	fovRow.AddChild(fovValueText)

	// render scaling combo box
	scalingRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(scalingRow)

	scalingLabel := widget.NewLabel(widget.LabelOpts.Text("Render Scaling", res.label.face, res.label.text))
	scalingRow.AddChild(scalingLabel)

	scalings := []interface{}{
		0.25,
		0.5,
		0.75,
		1.0,
	}

	var selectedScaling interface{}
	for _, s := range scalings {
		if s == m.game.renderScale {
			selectedScaling = s
		}
	}

	scalingCombo := newListComboButton(
		scalings,
		selectedScaling,
		func(e interface{}) string {
			return fmt.Sprintf("%0.0f%%", e.(float64)*100)
		},
		func(e interface{}) string {
			return fmt.Sprintf("%0.0f%%", e.(float64)*100)
		},
		func(args *widget.ListComboButtonEntrySelectedEventArgs) {
			s := args.Entry.(float64)
			m.game.setRenderScale(s)
		},
		res)
	scalingRow.AddChild(scalingCombo)

	// fullscreen checkbox
	fsCheckbox := newCheckbox("Fullscreen", m.game.fullscreen, func(args *widget.CheckboxChangedEventArgs) {
		m.game.setFullscreen(args.State == widget.WidgetChecked)
	}, res)
	c.AddChild(fsCheckbox)

	// vsync checkbox
	vsCheckbox := newCheckbox("Use VSync", m.game.vsync, func(args *widget.CheckboxChangedEventArgs) {
		m.game.setVsyncEnabled(args.State == widget.WidgetChecked)
	}, res)
	c.AddChild(vsCheckbox)

	// fps checkbox
	floorCheckbox := newCheckbox("Show FPS", m.game.fpsEnabled, func(args *widget.CheckboxChangedEventArgs) {
		m.game.fpsEnabled = args.State == widget.WidgetChecked
	}, res)
	c.AddChild(floorCheckbox)

	return &page{
		title:   "Display",
		content: c,
	}
}

func renderPage(m *GameMenu) *page {
	c := newPageContentContainer()
	res := m.res

	// render distance (meters) slider
	distanceRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(distanceRow)

	distanceLabel := widget.NewLabel(widget.LabelOpts.Text("Render Distance", res.label.face, res.label.text))
	distanceRow.AddChild(distanceLabel)

	var distanceValueText *widget.Label

	distanceSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(100, 6)),
		widget.SliderOpts.MinMax(-1, 4000),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			distanceValueText.Label = fmt.Sprintf("%dm", args.Current)
			m.game.setRenderDistance(float64(args.Current) / model.METERS_PER_UNIT)
		}),
	)
	distanceSlider.Current = int(m.game.renderDistance * model.METERS_PER_UNIT)
	distanceRow.AddChild(distanceSlider)

	distanceValueText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", distanceSlider.Current), res.label.face, res.label.text),
	)
	distanceRow.AddChild(distanceValueText)

	// floor/ground texturing checkbox
	floorCheckbox := newCheckbox("Ground Texturing", m.game.tex.renderFloorTex, func(args *widget.CheckboxChangedEventArgs) {
		m.game.tex.renderFloorTex = args.State == widget.WidgetChecked
	}, res)
	c.AddChild(floorCheckbox)

	return &page{
		title:   "Render",
		content: c,
	}
}

func hudPage(m *GameMenu) *page {
	c := newPageContentContainer()
	res := m.res

	// HUD alpha slider
	opacityRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(opacityRow)

	opacityLabel := widget.NewLabel(widget.LabelOpts.Text("Opacity", res.label.face, res.label.text))
	opacityRow.AddChild(opacityLabel)

	var opacityValueText *widget.Label

	opacitySlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(100, 6)),
		widget.SliderOpts.MinMax(0, 100),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			opacityValueText.Label = fmt.Sprintf("%d%%", args.Current)
			m.game.hudRGBA.A = uint8(math.Round(255 * float64(args.Current) / 100))
		}),
	)
	opacitySlider.Current = int(math.Round(100 * float64(m.game.hudRGBA.A) / 255))
	opacityRow.AddChild(opacitySlider)

	opacityValueText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", opacitySlider.Current), res.label.face, res.label.text),
	)
	opacityRow.AddChild(opacityValueText)

	// custom HUD color checkbox
	var pickerMinRGB *widget.Container
	customCheckbox := newCheckbox("Use Custom Color", m.game.hudUseCustomColor, func(args *widget.CheckboxChangedEventArgs) {
		m.game.hudUseCustomColor = args.State == widget.WidgetChecked

		// regenerate nav sprites to pick up color change
		m.game.loadNavSprites()

		// disable RGB picker if not using custom color
		for _, cb := range pickerMinRGB.Children() {
			cb.GetWidget().Disabled = !m.game.hudUseCustomColor
		}
	}, res)
	c.AddChild(customCheckbox)

	// custom HUD RGB picker
	hudRGB := &color.NRGBA{
		R: m.game.hudRGBA.R, G: m.game.hudRGBA.G, B: m.game.hudRGBA.B, A: 255,
	}
	pickerMinRGB = m.newColorPickerRGB("Color", hudRGB, func(args *widget.SliderChangedEventArgs) {
		m.game.hudRGBA.R = hudRGB.R
		m.game.hudRGBA.G = hudRGB.G
		m.game.hudRGBA.B = hudRGB.B

		// regenerate nav sprites to pick up color change
		m.game.loadNavSprites()
	})
	c.AddChild(pickerMinRGB)

	if !m.game.hudUseCustomColor {
		// start with HUD color picker enabled only if using custom HUD color setting
		for _, cb := range pickerMinRGB.Children() {
			cb.GetWidget().Disabled = true
		}
	}

	return &page{
		title:   "HUD",
		content: c,
	}
}

func lightingPage(m *GameMenu) *page {
	c := newPageContentContainer()
	res := m.res

	// raycaster lighting options for debug mode only
	debugLabel := widget.NewLabel(widget.LabelOpts.Text("~Debug Mode Only", res.label.face, res.label.text))
	c.AddChild(debugLabel)
	c.AddChild(m.newSeparator(res, widget.RowLayoutData{
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
			m.game.setLightFalloff(float64(args.Current))
		}),
	)
	falloffSlider.Current = int(m.game.lightFalloff)
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
			m.game.setGlobalIllumination(float64(args.Current))
		}),
	)
	globalSlider.Current = int(m.game.globalIllumination)
	globalRow.AddChild(globalSlider)

	globalValueText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", globalSlider.Current), res.label.face, res.label.text),
	)
	globalRow.AddChild(globalValueText)

	// min lighting RGB selection
	pickerMinRGB := m.newColorPickerRGB("Min Light", m.game.minLightRGB, func(args *widget.SliderChangedEventArgs) {
		m.game.setLightRGB(m.game.minLightRGB, m.game.maxLightRGB)
	})
	c.AddChild(pickerMinRGB)

	// max lighting RGB selection
	pickerMaxRGB := m.newColorPickerRGB("Max Light", m.game.maxLightRGB, func(args *widget.SliderChangedEventArgs) {
		m.game.setLightRGB(m.game.minLightRGB, m.game.maxLightRGB)
	})
	c.AddChild(pickerMaxRGB)

	return &page{
		title:   "~Lighting",
		content: c,
	}
}

func newPageContentContainer() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
		})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(10),
		)))
}

func newPageContainer(res *uiResources) *pageContainer {
	c := widget.NewContainer(
		// background image will instead be set based on which page is showing
		//widget.ContainerOpts.BackgroundImage(res.panel.image),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(res.panel.padding),
			widget.RowLayoutOpts.Spacing(15))),
	)

	titleText := widget.NewText(
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.TextOpts.Text("", res.text.titleFace, res.text.idleColor))
	c.AddChild(titleText)

	flipBook := widget.NewFlipBook(
		widget.FlipBookOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		}))),
	)
	c.AddChild(flipBook)

	return &pageContainer{
		widget:    c,
		titleText: titleText,
		flipBook:  flipBook,
	}
}

func (p *pageContainer) setPage(page *page) {
	p.titleText.Label = page.title
	p.flipBook.SetPage(page.content)
	p.flipBook.RequestRelayout()
}
