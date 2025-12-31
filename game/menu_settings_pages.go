package game

import (
	"fmt"
	"image/color"
	"math"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
)

type settingsPageContainer struct {
	widget    widget.PreferredSizeLocateableWidget
	titleText *widget.Text
	flipBook  *widget.FlipBook
}

type settingsPage struct {
	title    string
	content  widget.PreferredSizeLocateableWidget
	updaters []settingsUpdater
}

type settingsUpdater interface {
	update(*Game)
}

func (s *settingsPage) update(g *Game) {
	for _, updater := range s.updaters {
		updater.update(g)
	}
}

func gameMissionPage(m Menu) *settingsPage {
	c := newPageContentContainer()
	res := m.Resources()
	g := m.Game()

	mContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Top:    0,
				Bottom: 0,
				Left:   0,
				Right:  0,
			}),
			// Spacing defines how much space to put between each column and row
			widget.GridLayoutOpts.Spacing(0, m.Spacing()))),
	)
	c.AddChild(mContainer)

	// show container with Exit/Resume buttons
	bContainer := widget.NewContainer(
		// TODO: fix exit/resume container buttons not stretching to fit width
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(3),
			widget.GridLayoutOpts.Stretch([]bool{false, true, false}, []bool{false}),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Left:   m.Padding(),
				Right:  m.Padding(),
				Top:    m.Padding(),
				Bottom: m.Padding(),
			}))))
	mContainer.AddChild(bContainer)

	if g.osType == osTypeBrowser {
		// exit in browser kills but freezes the application, users can just close the tab/window
		bContainer.AddChild(newBlankSeparator(m.Resources(), m.Padding(), widget.RowLayoutData{
			Stretch: true,
		}))
	} else {
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
		bContainer.AddChild(exit)
	}

	bContainer.AddChild(newBlankSeparator(m.Resources(), m.Padding(), widget.RowLayoutData{
		Stretch: true,
	}))

	resume := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Resume", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) { g.closeMenu() }),
	)
	bContainer.AddChild(resume)

	// show mission objectives
	missionCard := createMissionCard(g, res, g.mission, MissionCardGame)
	mContainer.AddChild(missionCard)

	return &settingsPage{
		title:    "Mission",
		content:  c,
		updaters: []settingsUpdater{missionCard},
	}
}

func gameUnitPage(m Menu) *settingsPage {
	c := newPageContentContainer()
	res := m.Resources()
	g := m.Game()

	mContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Top:    0,
				Bottom: 0,
				Left:   0,
				Right:  0,
			}),
			// Spacing defines how much space to put between each column and row
			widget.GridLayoutOpts.Spacing(0, m.Spacing()))),
	)
	c.AddChild(mContainer)

	// show player unit card
	var playerUnit model.Unit
	if g.player != nil {
		playerUnit = g.player.Unit
	}
	unitCard := createUnitCard(g, res, playerUnit, UnitCardGame)
	mContainer.AddChild(unitCard)

	return &settingsPage{
		title:    "Unit",
		content:  c,
		updaters: []settingsUpdater{unitCard},
	}
}

func displayPage(m Menu) *settingsPage {
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

	resolutions := []any{}
	var selectedResolution any
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
		resolutions = append([]any{r}, resolutions...)
	}

	var fovSlider *widget.Slider
	resolutionCombo := newListComboButton(
		resolutions,
		selectedResolution,
		func(e any) string {
			return fmt.Sprintf("%s", e)
		},
		func(e any) string {
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
					// stop any scene transitions that may panic when resolution is changed before completion
					game.StopSceneTransition()

					// re-initialize the in-game menu with the Display settings pre-selected
					gameMenu.preSelectedPage = 2
					gameMenu.initResources()
					gameMenu.initMenu()
				case settingsMenu != nil:
					menuScene, ok := game.scene.(*MainMenuScene)
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

	scalings := []any{
		0.25,
		0.5,
		0.75,
		1.0,
	}

	var selectedScaling any
	for _, s := range scalings {
		if s == game.renderScale {
			selectedScaling = s
		}
	}

	scalingCombo := newListComboButton(
		scalings,
		selectedScaling,
		func(e any) string {
			return fmt.Sprintf("%0.0f%%", e.(float64)*100)
		},
		func(e any) string {
			return fmt.Sprintf("%0.0f%%", e.(float64)*100)
		},
		func(args *widget.ListComboButtonEntrySelectedEventArgs) {
			s := args.Entry.(float64)
			game.setRenderScale(s)
		},
		res)
	scalingRow.AddChild(scalingCombo)

	// fullscreen checkbox
	fsCheckbox := newCheckbox(m, "Fullscreen", game.fullscreen, func(args *widget.CheckboxChangedEventArgs) {
		game.setFullscreen(args.State == widget.WidgetChecked)
	})
	c.AddChild(fsCheckbox)

	// vsync checkbox
	vsCheckbox := newCheckbox(m, "Use VSync", game.vsync, func(args *widget.CheckboxChangedEventArgs) {
		game.setVsyncEnabled(args.State == widget.WidgetChecked)
	})
	c.AddChild(vsCheckbox)

	// fps checkbox
	fpsCheckbox := newCheckbox(m, "Show FPS", game.fpsEnabled, func(args *widget.CheckboxChangedEventArgs) {
		game.fpsEnabled = args.State == widget.WidgetChecked
	})
	c.AddChild(fpsCheckbox)

	return &settingsPage{
		title:   "Display",
		content: c,
	}
}

func audioPage(m Menu) *settingsPage {
	c := newPageContentContainer()
	res := m.Resources()
	game := m.Game()

	// background music volume slider
	bgmVolumeRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(bgmVolumeRow)

	bgmVolumeLabel := widget.NewLabel(widget.LabelOpts.Text("BGM Volume", res.label.face, res.label.text))
	bgmVolumeRow.AddChild(bgmVolumeLabel)

	var bgmValueText *widget.Label

	bgmVolumeSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(100, 6)),
		widget.SliderOpts.MinMax(0, 100),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			bgmValueText.Label = fmt.Sprintf("%d%%", args.Current)
			game.audio.SetMusicVolume(float64(args.Current) / 100)
		}),
	)
	bgmVolumeSlider.Current = int(bgmVolume * 100)
	bgmVolumeRow.AddChild(bgmVolumeSlider)

	bgmValueText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", bgmVolumeSlider.Current), res.label.face, res.label.text),
	)
	bgmVolumeRow.AddChild(bgmValueText)

	// sound effects volume slider
	sfxVolumeRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(sfxVolumeRow)

	sfxVolumeLabel := widget.NewLabel(widget.LabelOpts.Text("SFX Volume", res.label.face, res.label.text))
	sfxVolumeRow.AddChild(sfxVolumeLabel)

	var sfxValueText *widget.Label

	sfxVolumeSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(100, 6)),
		widget.SliderOpts.MinMax(0, 100),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			sfxValueText.Label = fmt.Sprintf("%d%%", args.Current)
			game.audio.SetSFXVolume(float64(args.Current) / 100)
		}),
	)
	sfxVolumeSlider.Current = int(sfxVolume * 100)
	sfxVolumeRow.AddChild(sfxVolumeSlider)

	sfxValueText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", sfxVolumeSlider.Current), res.label.face, res.label.text),
	)
	sfxVolumeRow.AddChild(sfxValueText)

	// sound effects channels slider
	sfxChannelsRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(sfxChannelsRow)

	sfxChannelsLabel := widget.NewLabel(widget.LabelOpts.Text("SFX Channels", res.label.face, res.label.text))
	sfxChannelsRow.AddChild(sfxChannelsLabel)

	var sfxChannelsText *widget.Label

	sfxChannelsSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(100, 6)),
		widget.SliderOpts.MinMax(8, 32),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			sfxChannelsText.Label = fmt.Sprintf("%d", args.Current)
			game.audio.SetSFXChannels(int(args.Current))
		}),
	)
	sfxChannelsSlider.Current = sfxChannels
	sfxChannelsRow.AddChild(sfxChannelsSlider)

	sfxChannelsText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", sfxChannelsSlider.Current), res.label.face, res.label.text),
	)
	sfxChannelsRow.AddChild(sfxChannelsText)

	return &settingsPage{
		title:   "Audio",
		content: c,
	}
}

func renderPage(m Menu) *settingsPage {
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
	floorCheckbox := newCheckbox(m, "Ground Texturing", game.tex.RenderFloorTex(), func(args *widget.CheckboxChangedEventArgs) {
		game.tex.SetRenderFloorTex(args.State == widget.WidgetChecked)
		game.initRenderFloorTex = game.tex.RenderFloorTex()
	})
	c.AddChild(floorCheckbox)

	// CRT shader checkbox
	crtCheckbox := newCheckbox(m, "CRT Shader", game.crtShader, func(args *widget.CheckboxChangedEventArgs) {
		game.crtShader = args.State == widget.WidgetChecked
	})
	c.AddChild(crtCheckbox)

	return &settingsPage{
		title:   "Render",
		content: c,
	}
}

func hudPage(m Menu) *settingsPage {
	c := newPageContentContainer()
	res := m.Resources()
	game := m.Game()

	// HUD scale slider
	scaleRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)
	c.AddChild(scaleRow)

	scaleLabel := widget.NewLabel(widget.LabelOpts.Text("Scale", res.label.face, res.label.text))
	scaleRow.AddChild(scaleLabel)

	var scaleValueText *widget.Label

	scaleSlider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}), widget.WidgetOpts.MinSize(100, 6)),
		widget.SliderOpts.MinMax(50, 100),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
		widget.SliderOpts.FixedHandleSize(res.slider.handleSize),
		widget.SliderOpts.TrackOffset(5),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			scaleValueText.Label = fmt.Sprintf("%d%%", args.Current)
			game.hudScale = float64(float64(args.Current) / 100)
		}),
	)
	scaleSlider.Current = int(math.Round(100 * float64(game.hudScale)))
	scaleRow.AddChild(scaleSlider)

	scaleValueText = widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(fmt.Sprintf("%d", scaleSlider.Current), res.label.face, res.label.text),
	)
	scaleRow.AddChild(scaleValueText)

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
	customCheckbox := newCheckbox(m, "Use Custom Color", game.hudUseCustomColor, func(args *widget.CheckboxChangedEventArgs) {
		game.hudUseCustomColor = args.State == widget.WidgetChecked

		// regenerate nav sprites to pick up color change
		game.loadNavSprites()

		// disable RGB picker if not using custom color
		for _, cb := range pickerMinRGB.Children() {
			cb.GetWidget().Disabled = !game.hudUseCustomColor
		}
	})
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

	// crosshair selection widget with graphical preview
	var crosshairPreview *widget.Graphic
	numCrosshairs := resources.CrosshairsSheet.Columns * resources.CrosshairsSheet.Rows
	crosshairLabel := widget.NewLabel(
		widget.LabelOpts.Text(
			fmt.Sprintf("Crosshair: %d/%d", game.hudCrosshairIndex+1, numCrosshairs),
			res.label.face,
			res.label.text,
		),
	)
	c.AddChild(crosshairLabel)

	cCrosshair := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(3),
			widget.GridLayoutOpts.Stretch([]bool{false, true, false}, []bool{false}),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Left:   m.Padding(),
				Right:  m.Padding(),
				Top:    m.Padding(),
				Bottom: m.Padding(),
			}))))

	crosshairSheet := resources.GetSpriteFromFile("hud/crosshairs_sheet.png")
	crosshairs := render.NewCrosshairs(
		crosshairSheet, resources.CrosshairsSheet.Columns, resources.CrosshairsSheet.Rows, game.hudCrosshairIndex,
	)
	crosshairSprite := crosshairs.Texture()
	imageH := float64(game.screenHeight) / 8
	spriteW, spriteH := float64(crosshairSprite.Bounds().Dx()), float64(crosshairSprite.Bounds().Dy())
	imageScale := imageH / spriteH

	unitImage := ebiten.NewImage(int(spriteW*imageScale), int(spriteH*imageScale))
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(imageScale, imageScale)
	unitImage.DrawImage(crosshairSprite, op)

	crosshairPreview = widget.NewGraphic(
		widget.GraphicOpts.Image(unitImage),
	)

	cPrev := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("<", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if game.hudCrosshairIndex > 0 {
				game.hudCrosshairIndex--
			}
			crosshairLabel.Label = fmt.Sprintf("Crosshair: %d/%d", game.hudCrosshairIndex+1, numCrosshairs)
			crosshairs := render.NewCrosshairs(
				crosshairSheet, resources.CrosshairsSheet.Columns, resources.CrosshairsSheet.Rows, game.hudCrosshairIndex,
			)
			cSprite := crosshairs.Texture()
			unitImage.Clear()
			unitImage.DrawImage(cSprite, op)

			if game.playerHUD != nil {
				game.playerHUD[HUD_CROSSHAIRS] = crosshairs
			}
		}),
	)

	cNext := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text(">", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if game.hudCrosshairIndex+1 < resources.CrosshairsSheet.Columns*resources.CrosshairsSheet.Rows {
				game.hudCrosshairIndex++
			}
			crosshairLabel.Label = fmt.Sprintf("Crosshair: %d/%d", game.hudCrosshairIndex+1, numCrosshairs)
			crosshairs := render.NewCrosshairs(
				crosshairSheet, resources.CrosshairsSheet.Columns, resources.CrosshairsSheet.Rows, game.hudCrosshairIndex,
			)
			cSprite := crosshairs.Texture()
			unitImage.Clear()
			unitImage.DrawImage(cSprite, op)

			if game.playerHUD != nil {
				game.playerHUD[HUD_CROSSHAIRS] = crosshairs
			}
		}),
	)

	cCrosshair.AddChild(cPrev)
	cCrosshair.AddChild(crosshairPreview)
	cCrosshair.AddChild(cNext)
	c.AddChild(cCrosshair)

	return &settingsPage{
		title:   "HUD",
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

func newSettingsPageContainer(m Menu) *settingsPageContainer {
	res := m.Resources()

	c := widget.NewContainer(
		// background image will instead be set based on which page is showing
		//widget.ContainerOpts.BackgroundImage(res.panel.image),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(res.panel.padding),
			widget.RowLayoutOpts.Spacing(m.Spacing()))),
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

	return &settingsPageContainer{
		widget:    c,
		titleText: titleText,
		flipBook:  flipBook,
	}
}

func (p *settingsPageContainer) setPage(page *settingsPage) {
	p.titleText.Label = page.title
	p.flipBook.SetPage(page.content)
	p.flipBook.RequestRelayout()
}
