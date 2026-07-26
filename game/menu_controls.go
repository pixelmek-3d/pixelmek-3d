package game

import (
	"strings"

	"github.com/ebitenui/ebitenui/widget"
	input "github.com/quasilyte/ebitengine-input"
	log "github.com/sirupsen/logrus"
)

var pressAnyKey *keyScanHandler

func controlsPage(m Menu) *settingsPage {
	c := newPageContentContainer()
	g := m.Game()

	// Create the container to layout the control rebind rows
	rebinds := widget.NewContainer(
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(4),
				widget.GridLayoutOpts.Stretch([]bool{false, true, false, false}, []bool{false}),
				widget.GridLayoutOpts.Spacing(4, 2),
			),
		),
	)

	// TODO: add control binds for all actions
	addControlBind(g, m, rebinds, ActionUp)
	addControlBind(g, m, rebinds, ActionDown)
	addControlBind(g, m, rebinds, ActionLeft)
	addControlBind(g, m, rebinds, ActionRight)

	// add key scan handler
	pressAnyKey = &keyScanHandler{}

	c.AddChild(rebinds)
	return &settingsPage{
		title:        "Controls",
		content:      c,
		tickUpdaters: []tickUpdater{pressAnyKey},
	}
}

type keyScanHandler struct {
	key              input.Key
	keyScanner       input.KeyScanner
	scanning         bool
	scanCompleteFunc func()
}

func (s *keyScanHandler) update() {
	if !s.scanning {
		return
	}

	key, status := s.keyScanner.Scan()
	if status != input.KeyScanChanged {
		s.key = key
	}
	if status == input.KeyScanCompleted {
		s.scanning = false

		// TODO: check for the new key to be available, resolve the conflicts here?

		s.scanCompleteFunc()
	}
}

func (s *keyScanHandler) startScan(scanCompleteFunc func()) {
	s.scanCompleteFunc = scanCompleteFunc
	s.scanning = true
}

func addControlBind(g *Game, m Menu, parent *widget.Container, action input.Action) {
	keyNames := g.input.ActionKeyNames(action, input.AnyDevice)

	res := m.Resources()
	label := widget.NewLabel(widget.LabelOpts.Text(actionString(action), res.fonts.face, res.label.text))
	parent.AddChild(label)

	parent.AddChild(newBlankSeparator(res, 20, widget.RowLayoutData{
		Stretch: true,
	}))

	var bindButton *widget.Button
	bindButton = widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text(strings.Join(keyNames, ", "), res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// open a modal window to wait for key press update
			openRebindWindow(m, action, func() {
				log.Debugf("[%s] add keybind to '%s'", actionString(action), pressAnyKey.key.String())
				g := m.Game()
				g.input.AddKeyBind(action, pressAnyKey.key)
				g.input.SetControls(g.input.keymap)

				keyNames := g.input.ActionKeyNames(action, input.AnyDevice)
				bindButton.SetText(strings.Join(keyNames, ", "))

				// TODO: save to file
			})
		}),
	)
	parent.AddChild(bindButton)

	clearButton := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("clear", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			g.input.ClearAction(action)
			g.input.SetControls(g.input.keymap)
			bindButton.SetText("")

			// TODO: save to file
		}),
	)
	parent.AddChild(clearButton)
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
	pressAnyKey.startScan(func() {
		rebindCompleteFunc()
		rmWindow()
	})
	m.SetWindow(window)
}
