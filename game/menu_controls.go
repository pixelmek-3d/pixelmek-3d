package game

import (
	"strings"

	"github.com/ebitenui/ebitenui/widget"
	input "github.com/quasilyte/ebitengine-input"
)

func controlsPage(m Menu) *settingsPage {
	c := newPageContentContainer()
	res := m.Resources()
	g := m.Game()

	// Create the container to layout the control rebind rows
	rebinds := widget.NewContainer(
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(4),
				widget.GridLayoutOpts.Stretch([]bool{true, false, false, false}, []bool{false}),
				widget.GridLayoutOpts.Spacing(4, 2),
			),
		),
	)

	// TODO: add control binds for all actions
	addControlBind(g, res, rebinds, ActionUp)
	addControlBind(g, res, rebinds, ActionDown)
	addControlBind(g, res, rebinds, ActionLeft)
	addControlBind(g, res, rebinds, ActionRight)

	c.AddChild(rebinds)
	return &settingsPage{
		title:   "Controls",
		content: c,
	}
}

func addControlBind(g *Game, res *uiResources, parent *widget.Container, action input.Action) {
	keyNames := g.input.ActionKeyNames(action, input.AnyDevice)

	label := widget.NewLabel(widget.LabelOpts.Text(actionString(action), res.fonts.face, res.label.text))
	parent.AddChild(label)

	bindButton := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text(strings.Join(keyNames, ", "), res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// TODO: ...
		}),
	)
	parent.AddChild(bindButton)

	resetButton := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("reset", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// TODO: ...
		}),
	)
	parent.AddChild(resetButton)

	clearButton := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("clear", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// TODO: ...
		}),
	)
	parent.AddChild(clearButton)
}
