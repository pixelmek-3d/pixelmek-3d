package game

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"
)

type UnitMenu struct {
	*MenuModel
}

func createUnitMenu(g *Game) *UnitMenu {
	var ui *ebitenui.UI = &ebitenui.UI{}

	menu := &UnitMenu{
		MenuModel: &MenuModel{
			game:   g,
			ui:     ui,
			active: true,
		},
	}

	menu.initResources()
	menu.initMenu()

	return menu
}

func (m *UnitMenu) initMenu() {
	m.MenuModel.initMenu()
	m.root.BackgroundImage = m.Resources().background

	// menu title
	titleBar := unitTitleContainer(m)
	m.root.AddChild(titleBar)

	// unit selection
	selection := unitSelectionContainer(m)
	m.root.AddChild(selection)

	// footer
	footer := unitMenuFooterContainer(m)
	m.root.AddChild(footer)
}

func (m *UnitMenu) Update() {
	m.ui.Update()
}

func (m *UnitMenu) Draw(screen *ebiten.Image) {
	m.ui.Draw(screen)
}

func unitTitleContainer(m *UnitMenu) *widget.Container {
	res := m.Resources()

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:   m.Padding(),
				Right:  m.Padding(),
				Top:    m.Padding(),
				Bottom: m.Padding(),
			}))))

	c.AddChild(widget.NewText(
		widget.TextOpts.Text("Unit Selection", res.text.bigTitleFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	return c
}

func unitMenuFooterContainer(m *UnitMenu) *widget.Container {
	game := m.Game()
	res := m.Resources()

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(3),
			widget.GridLayoutOpts.Stretch([]bool{false, true, false}, []bool{false}),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:   m.Padding(),
				Right:  m.Padding(),
				Top:    m.Padding(),
				Bottom: m.Padding(),
			}))))

	back := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Back", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			iScene, _ := game.scene.(*InstantActionScene)
			iScene.back()
		}),
	)
	c.AddChild(back)

	c.AddChild(newBlankSeparator(m, widget.RowLayoutData{
		Stretch: true,
	}))

	next := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Next", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			iScene, _ := game.scene.(*InstantActionScene)
			iScene.next()
		}),
	)
	c.AddChild(next)

	return c
}

type unitPageContainer struct {
	widget    widget.PreferredSizeLocateableWidget
	titleText *widget.Text
	flipBook  *widget.FlipBook
}

type unitPage struct {
	title    string
	content  widget.PreferredSizeLocateableWidget
	unit     model.Unit
	variants []model.Unit
}

func newUnitPageContainer(m *UnitMenu) *unitPageContainer {
	res := m.Resources()

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.image),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(res.panel.padding),
			widget.RowLayoutOpts.Spacing(m.Spacing()))),
	)

	titleText := widget.NewText(
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.TextOpts.Text("", res.text.face, res.text.idleColor))
	c.AddChild(titleText)

	flipBook := widget.NewFlipBook(
		widget.FlipBookOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		}))),
	)
	c.AddChild(flipBook)

	return &unitPageContainer{
		widget:    c,
		titleText: titleText,
		flipBook:  flipBook,
	}
}

func (p *unitPageContainer) setPage(page *unitPage) {
	if page.unit == nil {
		p.titleText.Label = "Random"
	} else {
		p.titleText.Label = page.unit.Name()
	}
	p.flipBook.SetPage(page.content)
	p.flipBook.RequestRelayout()
}

func unitSelectionContainer(m *UnitMenu) widget.PreferredSizeLocateableWidget {
	res := m.Resources()
	game := m.Game()

	chassisList := []string{}
	chassisMap := make(map[string][]model.Unit, 32)
	for _, unitResource := range game.resources.GetMechResourceList() {
		chassis := unitResource.Name
		_, found := chassisMap[chassis]
		if !found {
			chassisList = append(chassisList, chassis)
			chassisMap[chassis] = make([]model.Unit, 0, 4)
		}
		unit := game.createModelMechFromResource(unitResource)
		chassisMap[chassis] = append(chassisMap[chassis], unit)
	}

	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:  m.Spacing(),
				Right: m.Spacing(),
			}),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(m.Spacing(), 0),
		)))

	// sort by weight and then chassis name
	sort.Slice(chassisList, func(i, j int) bool {
		unitA, unitB := chassisMap[chassisList[i]][0], chassisMap[chassisList[j]][0]
		if unitA.Tonnage() == unitB.Tonnage() {
			return unitA.Name() < unitB.Name()
		}
		return unitA.Tonnage() < unitB.Tonnage()
	})

	// sort within chassis by variant designation (except Prime comes first)
	for _, variantList := range chassisMap {
		sort.Slice(variantList, func(i, j int) bool {
			unitA, unitB := variantList[i], variantList[j]

			if strings.HasSuffix(strings.ToLower(unitA.Variant()), "prime") {
				return true
			} else if strings.HasSuffix(strings.ToLower(unitB.Variant()), "prime") {
				return false
			}
			return unitA.Variant() < unitB.Variant()
		})
	}

	pages := make([]interface{}, 0, 1+len(chassisMap))

	// add entry for random unit
	randomUnitPage := unitSelectionPage(m, nil, []model.Unit{})
	pages = append(pages, randomUnitPage)

	for _, chassis := range chassisList {
		unitList := chassisMap[chassis]
		unitPage := unitSelectionPage(m, unitList[0], unitList)
		pages = append(pages, unitPage)
	}

	pageContainer := newUnitPageContainer(m)

	pageList := widget.NewList(
		widget.ListOpts.Entries(pages),
		widget.ListOpts.EntryLabelFunc(func(e interface{}) string {
			return e.(*unitPage).title
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
			nextPage := args.Entry.(*unitPage)
			pageContainer.setPage(nextPage)
			m.Root().RequestRelayout()
		}))

	c.AddChild(pageList)

	c.AddChild(pageContainer.widget)

	pageList.SetSelectedEntry(pages[0])

	return c
}

func unitSelectionPage(m *UnitMenu, unit model.Unit, variants []model.Unit) *unitPage {
	c := newPageContentContainer()
	res := m.Resources()

	unitTable := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(10),
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
	)
	c.AddChild(unitTable)

	// TODO: do not load sprite/graphic until the page is selected

	// show unit image graphic
	var sprite *render.Sprite
	switch interfaceType := unit.(type) {
	case *model.Mech:
		sprite = m.game.createUnitSprite(unit).(*render.MechSprite).Sprite
	case nil:
		// nil represents random unit selection
		sprite = nil
	default: // TODO: handle any unit type
		panic(fmt.Errorf("currently unable to handle selection of model.Unit for type %v", interfaceType))
	}

	var imageLabel widget.PreferredSizeLocateableWidget
	if sprite == nil {
		imageLabel = widget.NewLabel(widget.LabelOpts.Text("?", res.fonts.bigTitleFace, res.label.text))
	} else {
		imageLabel = widget.NewGraphic(
			widget.GraphicOpts.Image(sprite.Texture()),
		)
	}
	unitTable.AddChild(imageLabel)

	// show unit variant selection
	if unit != nil {
		comboVariants := []interface{}{}
		for _, v := range variants {
			comboVariants = append(comboVariants, v)
		}

		variantCombo := newListComboButton(
			comboVariants,
			unit,
			func(e interface{}) string {
				u := e.(model.Unit)
				if u != nil {
					return u.Variant()
				}
				return "?"
			},
			func(e interface{}) string {
				u := e.(model.Unit)
				if u != nil {
					return u.Variant()
				}
				return "?"
			},
			func(args *widget.ListComboButtonEntrySelectedEventArgs) {
				u := args.Entry.(model.Unit)
				if u != nil {
					// TODO: set selected as unit
					fmt.Printf("%s\n", u.Variant())
				}
			},
			res)
		unitTable.AddChild(variantCombo)

		if len(comboVariants) <= 1 {
			// only allow variant selection if more than one to choose from
			variantCombo.GetWidget().Disabled = true
		}
	}

	// TODO: more content

	var unitName, unitTonnage string
	if unit == nil {
		unitName = "Random"
		unitTonnage = "??"
	} else {
		unitName = unit.Name()
		unitTonnage = fmt.Sprintf("%0.0f", unit.Tonnage())
	}

	return &unitPage{
		title:    fmt.Sprintf("%s - %s", unitTonnage, unitName),
		content:  c,
		unit:     unit,
		variants: variants,
	}
}
