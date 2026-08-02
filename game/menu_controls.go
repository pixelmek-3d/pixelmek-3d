package game

import (
	"strings"

	"github.com/ebitenui/ebitenui/widget"
	input "github.com/quasilyte/ebitengine-input"
	log "github.com/sirupsen/logrus"
)

var pressAnyKeyScanner *keyScanHandler

func controlsPage(m Menu) *settingsPage {
	c := newPageContentContainer()
	res := m.Resources()

	// add key scan handler
	pressAnyKeyScanner = &keyScanHandler{keyScanner: input.NewKeyScanner(m.Game().input.Handler)}

	page := &settingsPage{
		title:        "Controls",
		content:      c,
		tickUpdaters: []tickUpdater{pressAnyKeyScanner},
	}

	modifyControlsButton := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("Modify Control Binds", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			openEditKeysWindow(m, page)
		}),
	)
	c.AddChild(modifyControlsButton)
	return page
}

type keyScanHandler struct {
	key              input.Key
	keyScanner       *input.KeyScanner
	scanning         bool
	scanCompleteFunc func()
}

func (s *keyScanHandler) update() {
	if !s.scanning {
		return
	}

	key, status := s.keyScanner.Scan()
	if status == input.KeyScanCompleted {
		s.key = key
		s.scanning = false
		s.scanCompleteFunc()
	}
}

func (s *keyScanHandler) startScan(scanCompleteFunc func()) {
	s.scanCompleteFunc = scanCompleteFunc
	s.scanning = true
}

func openEditKeysWindow(m Menu, page *settingsPage) {
	var rmWindow widget.RemoveWindowFunc
	var window *widget.Window

	g := m.Game()
	uiRect := g.uiRect()
	res := m.Resources()
	padding := m.Padding()
	spacing := m.Spacing()

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

			// TODO: set all controls to their defaults and update the current view

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

			// TODO: revert any recent keybind changes

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

			// TODO: save current keybinds and apply to the runtime

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
	var bindButton *widget.Button
	keyScanCompleteFunc := func() {
		log.Debugf("[%s] add keybind to '%s'", actionString(action), pressAnyKeyScanner.key.String())
		g := m.Game()
		err := g.input.AddKeyBind(action, pressAnyKeyScanner.key)
		if err != nil {
			log.Errorf("%v - TODO: display error on screen!", err)
			return
		}

		g.input.SetControls(g.input.keymap)

		keyNames := g.input.ActionKeyNames(action, input.AnyDevice)
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

	label := widget.NewLabel(widget.LabelOpts.Text(actionString(action), res.fonts.face, res.label.text))
	c.AddChild(label)

	c.AddChild(newBlankSeparator(res, m.Padding(), widget.RowLayoutData{
		Stretch: true,
	}))

	keyNames := g.input.ActionKeyNames(action, input.AnyDevice)
	bindButton = widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text(strings.Join(keyNames, ", "), res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// open a modal window to wait for key press update
			openRebindWindow(m, action, keyScanCompleteFunc)
		}),
	)
	c.AddChild(bindButton)

	clearButton := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("clear", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			g.input.ClearAction(action)
			g.input.SetControls(g.input.keymap)
			bindButton.SetText("")
			page.content.RequestRelayout()
		}),
	)
	c.AddChild(clearButton)

	return c
}

func openRebindWindow(m Menu, action input.Action, rebindCompleteFunc func()) {
	var rmWindow widget.RemoveWindowFunc
	var window *widget.Window

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

	label := widget.NewLabel(widget.LabelOpts.Text("Press a Key or Button...", res.fonts.face, res.label.text))
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
	pressAnyKeyScanner.startScan(func() {
		rebindCompleteFunc()
		rmWindow()
	})
	m.AddWindow(window)
}
