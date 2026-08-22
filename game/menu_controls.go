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
	modifiedHandler *input.Handler
	modifiedKeymap  input.Keymap
	keyScanner      *keyScanHandler
)

func controlsPage(m Menu) *settingsPage {
	c := newPageContentContainer()
	g := m.Game()
	res := m.Resources()

	// set modified handler and key scanner
	modifiedHandler = g.input.inputSystem.NewHandler(0, input.Keymap{})
	keyScanner = &keyScanHandler{keyScanner: input.NewKeyScanner(modifiedHandler)}

	page := &settingsPage{
		title:        "Controls",
		content:      c,
		tickUpdaters: []tickUpdater{keyScanner},
	}

	modifyControlsButton := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("Modify Control Binds", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			openEditKeysWindow(m, page, g.input.Controls())
		}),
	)
	c.AddChild(modifyControlsButton)
	return page
}

type keyScanHandler struct {
	keyScanner       *input.KeyScanner
	key              input.Key
	axes             input.Key
	scanningKey      bool
	scanningAxes     bool
	scanCompleteFunc func()
}

func (s *keyScanHandler) update() {
	if !s.scanningKey && !s.scanningAxes {
		return
	}

	if s.scanningKey {
		s.handleRemapKey()
	} else if s.scanningAxes {
		s.handleRemapAxes()
	}

}

func (s *keyScanHandler) handleRemapKey() {
	key, status := s.keyScanner.Scan()
	if status == input.KeyScanCompleted {
		s.key = key
		s.scanningKey = false
		s.scanCompleteFunc()
	}
}

func (s *keyScanHandler) handleRemapAxes() {
	axes, status := s.keyScanner.ScanAxes()
	if status == input.KeyScanCompleted {
		s.axes = axes
		s.scanningAxes = false
		s.scanCompleteFunc()
	}
}

func (s *keyScanHandler) startKeyScan(scanType keyScanType, scanCompleteFunc func()) {
	s.scanCompleteFunc = scanCompleteFunc
	switch scanType {
	case keyScanKeys:
		s.scanningKey = true
	case keyScanAxes:
		s.scanningAxes = true
	}
}

func openEditKeysWindow(m Menu, page *settingsPage, keymap input.Keymap) {
	var rmWindow widget.RemoveWindowFunc
	var window *widget.Window

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
		widget.TextOpts.Text("Modify Control Binds", res.text.titleFace, res.text.idleColor),
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

	content := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		widget.RowLayoutOpts.Spacing(20),
		widget.RowLayoutOpts.Padding(&widget.Insets{Top: 10, Bottom: 10}),
	)))

	// add control binds for all actions
	for action := range actionCount {
		binder := addControlBind(g, m, page, action)
		if binder != nil {
			content.AddChild(binder)
		}
	}

	scrollContainer := newScrollContainer(m, content)
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
			rmWindow()
			openEditKeysWindow(m, page, defaultControls())
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
			rmWindow()
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
			g.input.Remap(modifiedKeymap)
			g.input.saveControls()
			rmWindow()
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

	rmWindow = m.UI().AddWindow(window)
	m.AddWindow(window)
}

func addControlBind(g *Game, m Menu, page *settingsPage, action input.Action) *widget.Container {
	if action == ActionUnknown {
		return nil
	}
	actionStr := actionString(action)
	scanType := keyScanKeys
	if strings.Contains(actionStr, "_axes") {
		scanType = keyScanAxes
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

		keyNames := modifiedHandler.ActionKeyNames(action, input.AnyDevice)
		bindButton.SetText(strings.Join(keyNames, ", "))
		page.content.RequestRelayout()
	}

	res := m.Resources()
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(4),
				widget.GridLayoutOpts.Stretch([]bool{true, true, false, false}, []bool{false}),
				widget.GridLayoutOpts.Spacing(4, 2),
			),
		),
	)

	label := widget.NewLabel(widget.LabelOpts.Text(actionStr, res.fonts.face, res.label.text))
	c.AddChild(label)

	c.AddChild(newBlankSeparator(res, m.Padding(), widget.RowLayoutData{
		Stretch: true,
	}))

	keyNames := modifiedHandler.ActionKeyNames(action, input.AnyDevice)
	bindButton = widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text(strings.Join(keyNames, ", "), res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// open a modal window to wait for key/axes pressed update
			log.Debugf("[%s] opening rebind window", actionStr)
			openRebindWindow(m, action, scanType, keyScanCompleteFunc)
		}),
	)
	c.AddChild(bindButton)

	clearButton := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("clear", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			ClearAction(modifiedKeymap, action)
			bindButton.SetText("")
			page.content.RequestRelayout()
		}),
	)
	c.AddChild(clearButton)

	return c
}

func openRebindWindow(m Menu, action input.Action, scanType keyScanType, rebindCompleteFunc func()) {
	var rmWindow widget.RemoveWindowFunc
	var window *widget.Window

	var scanLabelStr string
	switch scanType {
	case keyScanKeys:
		scanLabelStr = "Press a Key or Button..."
	case keyScanAxes:
		scanLabelStr = "Move an Axes..."
	default:
		panic("unknown keyScanType: " + string(scanType))
	}

	game := m.Game()
	uiRect := game.uiRect()
	res := m.Resources()
	padding := m.Padding()
	spacing := m.Spacing()

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

	rmWindow = m.UI().AddWindow(window)
	keyScanner.startKeyScan(scanType, func() {
		rebindCompleteFunc()
		rmWindow()
	})
	m.AddWindow(window)
}
