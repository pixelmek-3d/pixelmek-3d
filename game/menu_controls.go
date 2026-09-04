package game

import (
	"strings"

	"github.com/ebitenui/ebitenui/widget"
	input "github.com/quasilyte/ebitengine-input"
	log "github.com/sirupsen/logrus"
)

type keyScanType uint8

const (
	keyScanAxes keyScanType = iota
	keyScanKeys
)

var (
	modifiedKeymap       input.Keymap
	keyboardMouseScanner *keyScanHandler
	gamepadScanner       *keyScanHandler
)

func controlsPage(m Menu) *settingsPage {
	c := newPageContentContainer()
	g := m.Game()
	res := m.Resources()

	// make separate handlers and scanners for keyboard/mouse only and gamepad only inputs
	keyboardMouseScanner = NewKeyScanHandler(KeymapTypeKeyboardMouse)
	gamepadScanner = NewKeyScanHandler(KeymapTypeGamepad)

	page := &settingsPage{
		title:        "Controls",
		content:      c,
		tickUpdaters: []tickUpdater{keyboardMouseScanner, gamepadScanner},
	}

	keyboardMouseControls := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("Keyboard/Mouse Controls", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			openModifyControlsWindow(m, page, g.input.KeyboardMouseControls(), KeymapTypeKeyboardMouse)
		}),
	)
	gamepadControls := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("Gamepad Controls", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			openModifyControlsWindow(m, page, g.input.GamepadControls(), KeymapTypeGamepad)
		}),
	)
	c.AddChild(keyboardMouseControls)
	c.AddChild(gamepadControls)
	return page
}

func openModifyControlsWindow(m Menu, page *settingsPage, keymap input.Keymap, keymapType KeymapType) {
	var window *widget.Window

	var windowTitle string
	var modifiedHandler *input.Handler
	switch keymapType {
	case KeymapTypeKeyboardMouse:
		windowTitle = "Keyboard/Mouse Controls"
		modifiedHandler = keyboardMouseScanner.handler
	case KeymapTypeGamepad:
		windowTitle = "Gamepad Controls"
		modifiedHandler = gamepadScanner.handler
	}

	g := m.Game()
	uiRect := g.uiRect()
	res := m.Resources()
	padding := m.Padding()
	spacing := m.Spacing()

	// copy existing key mapping prior to modification
	modifiedKeymap = keymap.Clone()
	modifiedHandler.Remap(modifiedKeymap)

	titleBar := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true}),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Left:   padding,
				Right:  padding,
				Top:    padding,
				Bottom: padding,
			}))))

	titleBar.AddChild(widget.NewText(
		widget.TextOpts.Text(windowTitle, res.text.titleFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.image),
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(1),
				widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true, false, true}),
				widget.GridLayoutOpts.Padding(res.panel.padding),
				widget.GridLayoutOpts.Spacing(1, spacing),
			),
		),
	)

	// add control binds for all actions in a grid container
	controlsGrid := widget.NewContainer(
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(4),
				widget.GridLayoutOpts.Stretch([]bool{true, false, false, false}, []bool{false}),
				widget.GridLayoutOpts.Spacing(4, 4),
			),
		),
	)
	for action := range actionCount {
		addControlBind(m, controlsGrid, action, keymapType)
	}

	scrollContainer := newScrollContainer(m, controlsGrid)
	c.AddChild(scrollContainer)

	//  add footer section buttons below scroll container to set defaults, save and cancel
	footer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(4),
			widget.GridLayoutOpts.Stretch([]bool{false, true, false, false}, []bool{false}),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Left:   m.Padding(),
				Right:  m.Padding(),
				Top:    m.Padding(),
				Bottom: m.Padding(),
			}))))
	c.AddChild(footer)

	setDefaultsButton := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Set Defaults", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// set all controls to their defaults and update the current view
			log.Debug("setting default key bindings")
			window.Close()

			var defaultKeymap input.Keymap
			switch keymapType {
			case KeymapTypeKeyboardMouse:
				defaultKeymap = defaultKeyboardMouseControls()
			case KeymapTypeGamepad:
				defaultKeymap = defaultGamepadControls()
			}
			openModifyControlsWindow(m, page, defaultKeymap, keymapType)
		}),
	)
	footer.AddChild(setDefaultsButton)

	footer.AddChild(newBlankSeparator(m.Resources(), m.Padding(), widget.RowLayoutData{
		Stretch: true,
	}))

	cancelButton := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Cancel", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// cancel any modified keybind changes
			window.Close()
		}),
	)
	footer.AddChild(cancelButton)

	saveButton := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Save", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// current keybinds already applied to the runtime, save to disk
			log.Debugf("saving modified key bindings")
			keyboardMouseMap := g.input.keyboardMouseMap
			gamepadMap := g.input.gamepadMap
			switch keymapType {
			case KeymapTypeKeyboardMouse:
				keyboardMouseMap = modifiedKeymap
			case KeymapTypeGamepad:
				gamepadMap = modifiedKeymap
			}
			g.input.SetControls(keyboardMouseMap, gamepadMap)
			g.input.saveControls()
			window.Close()
		}),
	)
	footer.AddChild(saveButton)

	// open in separate menu window
	window = widget.NewWindow(
		widget.WindowOpts.Modal(),
		widget.WindowOpts.Contents(c),
		widget.WindowOpts.TitleBar(titleBar, uiRect.Dy()/12),
	)

	wRect := uiRect.Inset(padding)
	window.SetLocation(wRect)

	m.AddWindow(window)
}

func addControlBind(m Menu, gridContainer *widget.Container, action input.Action, keymapType KeymapType) {
	if action == ActionUnknown {
		return
	}

	var keyScanner *keyScanHandler
	var modifiedHandler *input.Handler
	switch keymapType {
	case KeymapTypeKeyboardMouse:
		keyScanner = keyboardMouseScanner
		modifiedHandler = keyboardMouseScanner.handler
	case KeymapTypeGamepad:
		keyScanner = gamepadScanner
		modifiedHandler = gamepadScanner.handler
	}

	actionStr := actionString(action)
	scanType := keyScanKeys
	if strings.Contains(actionStr, "_axes") {
		scanType = keyScanAxes
	}

	res := m.Resources()
	label := widget.NewLabel(widget.LabelOpts.Text(actionStr, res.fonts.face, res.label.text))
	gridContainer.AddChild(label)

	// create exactly two bind buttons, regardless of number of keys currently bound
	keyNames := modifiedHandler.ActionKeyNames(action, input.AnyDevice)
	var bindButtonWidgets [2]*widget.Button
	for i := range 2 {
		var kName string
		if i < len(keyNames) {
			kName = keyNames[i]
		}
		var bindButton *widget.Button
		keyScanCompleteFunc := func() {
			keyBind := keyScanner.key
			if scanType == keyScanAxes {
				keyBind = keyScanner.axes
			}
			log.Debugf("[%s] add keybind to '%s'", actionStr, keyBind.String())
			err := AddKeyBind(modifiedKeymap, action, keyBind)
			if err != nil {
				log.Errorf("%v - TODO: display error on screen!", err)
				return
			}

			bindButton.SetText(keyBind.String())
		}

		bindButton = widget.NewButton(
			widget.ButtonOpts.Image(res.button.image),
			widget.ButtonOpts.TextPadding(res.button.padding),
			widget.ButtonOpts.Text(kName, res.button.face, res.button.text),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				// open a modal window to wait for key/axes pressed update
				log.Debugf("[%s] opening rebind window", actionStr)
				openRebindWindow(m, action, scanType, keymapType, keyScanCompleteFunc)
			}),
		)
		gridContainer.AddChild(bindButton)
		bindButtonWidgets[i] = bindButton
	}

	clearButton := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("clear", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			ClearAction(modifiedKeymap, action)
			for _, bindButton := range bindButtonWidgets {
				bindButton.SetText("")
			}
		}),
	)
	gridContainer.AddChild(clearButton)
}

func openRebindWindow(m Menu, action input.Action, scanType keyScanType, keymapType KeymapType, rebindCompleteFunc func()) {
	var window *widget.Window

	var keyScanner *keyScanHandler
	switch keymapType {
	case KeymapTypeKeyboardMouse:
		keyScanner = keyboardMouseScanner
	case KeymapTypeGamepad:
		keyScanner = gamepadScanner
	}

	game := m.Game()
	uiRect := game.uiRect()
	res := m.Resources()
	padding := m.Padding()
	spacing := m.Spacing()

	scanLabelStr := "Press any Key or Button..."
	switch {
	case scanType == keyScanKeys && keymapType == KeymapTypeGamepad:
		scanLabelStr = "Press any Button..."
	case scanType == keyScanAxes:
		scanLabelStr = "Move any Axes..."
	}

	titleBar := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Left:   padding,
				Right:  padding,
				Top:    padding,
				Bottom: padding,
			}))))

	titleBar.AddChild(widget.NewText(
		widget.TextOpts.Text("Action: "+actionString(action), res.text.titleFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.image),
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(1),
				widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true, true, true}),
				widget.GridLayoutOpts.Padding(res.panel.padding),
				widget.GridLayoutOpts.Spacing(1, spacing),
			),
		),
	)

	c.AddChild(newBlankSeparator(res, 12, widget.RowLayoutData{
		Stretch: true,
	}))

	label := widget.NewLabel(widget.LabelOpts.Text(scanLabelStr, res.fonts.face, res.label.text))
	c.AddChild(label)

	c.AddChild(newBlankSeparator(res, 12, widget.RowLayoutData{
		Stretch: true,
	}))

	window = widget.NewWindow(
		widget.WindowOpts.Modal(),
		widget.WindowOpts.Contents(c),
		widget.WindowOpts.TitleBar(titleBar, uiRect.Dy()/12),
	)

	wRect := uiRect.Inset(uiRect.Dy() / 6)
	window.SetLocation(wRect)

	keyScanner.startKeyScan(scanType, func() {
		rebindCompleteFunc()
		window.Close()
	})
	m.AddWindow(window)
}
