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

func gamePage(m Menu) *page {
	c := newPageContentContainer()
	res := m.Resources()
	game := m.Game()

	resume := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Resume", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) { game.closeMenu() }),
	)
	c.AddChild(resume)

	if game.osType == osTypeBrowser {
		// exit in browser kills but freezes the application, users can just close the tab/window
	} else {
		// show in game exit button
		c.AddChild(newSeparator(m, res, widget.RowLayoutData{
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
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) { exit(0) }),
		)
		c.AddChild(exit)
	}

	return &page{
		title:   "Game",
		content: c,
	}
}

func displayPage(m Menu) *page {
	c := newPageContentContainer()
	res := m.Resources()
	game := m.Game()

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
	for _, r := range m.Resolutions() {
		resolutions = append(resolutions, r)
		if game.screenWidth == r.width && game.screenHeight == r.height {
			selectedResolution = r
		}
	}

	if selectedResolution == nil {
		// generate custom entry to put at top of the list
		r := MenuResolution{
			width:  game.screenWidth,
			height: game.screenHeight,
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
			if game.screenWidth != r.width || game.screenHeight != r.height {
				game.setResolution(r.width, r.height)

				// pre-select ideal FOV for the aspect ratio
				game.setFovAngle(float64(r.aspectRatio.fov))

				gameMenu, _ := m.(*GameMenu)
				settingsMenu, _ := m.(*SettingsMenu)
				switch {
				case gameMenu != nil:
					// re-initialize the menu with the Display settings pre-selected
					gameMenu.preSelectedPage = 1
					gameMenu.initResources()
					gameMenu.initMenu()
				case settingsMenu != nil:
					menuScene, ok := game.scene.(*MenuScene)
					if ok {
						menuScene.settings.preSelectedPage = 0
						menuScene.settings.initResources()
						menuScene.settings.initMenu()
						menuScene.main.initResources()
						menuScene.main.initMenu()
					}
				}
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
			game.setFovAngle(float64(args.Current))
		}),
	)
	fovSlider.Current = int(game.fovDegrees)
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
		if s == game.renderScale {
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
			game.setRenderScale(s)
		},
		res)
	scalingRow.AddChild(scalingCombo)

	// fullscreen checkbox
	fsCheckbox := newCheckbox("Fullscreen", game.fullscreen, func(args *widget.CheckboxChangedEventArgs) {
		game.setFullscreen(args.State == widget.WidgetChecked)
	}, res)
	c.AddChild(fsCheckbox)

	// vsync checkbox
	vsCheckbox := newCheckbox("Use VSync", game.vsync, func(args *widget.CheckboxChangedEventArgs) {
		game.setVsyncEnabled(args.State == widget.WidgetChecked)
	}, res)
	c.AddChild(vsCheckbox)

	// fps checkbox
	floorCheckbox := newCheckbox("Show FPS", game.fpsEnabled, func(args *widget.CheckboxChangedEventArgs) {
		game.fpsEnabled = args.State == widget.WidgetChecked
	}, res)
	c.AddChild(floorCheckbox)

	return &page{
		title:   "Display",
		content: c,
	}
}

func renderPage(m Menu) *page {
	c := newPageContentContainer()
	res := m.Resources()
	game := m.Game()

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
			game.setRenderDistance(float64(args.Current) / model.METERS_PER_UNIT)
		}),
	)
	distanceSlider.Current = int(game.renderDistance * model.METERS_PER_UNIT)
	distanceRow.AddChild(distanceSlider)

	distanceValueText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", distanceSlider.Current), res.label.face, res.label.text),
	)
	distanceRow.AddChild(distanceValueText)

	// floor/ground texturing checkbox
	floorCheckbox := newCheckbox("Ground Texturing", game.tex.renderFloorTex, func(args *widget.CheckboxChangedEventArgs) {
		game.tex.renderFloorTex = args.State == widget.WidgetChecked
	}, res)
	c.AddChild(floorCheckbox)

	return &page{
		title:   "Render",
		content: c,
	}
}

func hudPage(m Menu) *page {
	c := newPageContentContainer()
	res := m.Resources()
	game := m.Game()

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
			game.hudRGBA.A = uint8(math.Round(255 * float64(args.Current) / 100))
		}),
	)
	opacitySlider.Current = int(math.Round(100 * float64(game.hudRGBA.A) / 255))
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
	customCheckbox := newCheckbox("Use Custom Color", game.hudUseCustomColor, func(args *widget.CheckboxChangedEventArgs) {
		game.hudUseCustomColor = args.State == widget.WidgetChecked

		// regenerate nav sprites to pick up color change
		game.loadNavSprites()

		// disable RGB picker if not using custom color
		for _, cb := range pickerMinRGB.Children() {
			cb.GetWidget().Disabled = !game.hudUseCustomColor
		}
	}, res)
	c.AddChild(customCheckbox)

	// custom HUD RGB picker
	hudRGB := &color.NRGBA{
		R: game.hudRGBA.R, G: game.hudRGBA.G, B: game.hudRGBA.B, A: 255,
	}
	pickerMinRGB = newColorPickerRGB(m, "Color", hudRGB, func(args *widget.SliderChangedEventArgs) {
		game.hudRGBA.R = hudRGB.R
		game.hudRGBA.G = hudRGB.G
		game.hudRGBA.B = hudRGB.B

		// regenerate nav sprites to pick up color change
		game.loadNavSprites()
	})
	c.AddChild(pickerMinRGB)

	if !game.hudUseCustomColor {
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

func lightingPage(m Menu) *page {
	c := newPageContentContainer()
	res := m.Resources()
	game := m.Game()

	// raycaster lighting options for debug mode only
	debugLabel := widget.NewLabel(widget.LabelOpts.Text("~Debug Mode Only", res.label.face, res.label.text))
	c.AddChild(debugLabel)
	c.AddChild(newSeparator(m, res, widget.RowLayoutData{
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
