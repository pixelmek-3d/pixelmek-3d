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
	selectedUnit model.Unit
}

type unitPageContainer struct {
	unitMenu         *UnitMenu
	widget           *widget.Container
	variantContainer *widget.Container
	flipBook         *widget.FlipBook
}

type unitPage struct {
	title    string
	content  *widget.Container
	unit     model.Unit
	variants []model.Unit
}

type UnitCardStyle int

const (
	UnitCardSelect UnitCardStyle = iota
	UnitCardLaunch
	UnitCardMission
)

type UnitCard struct {
	*widget.Container
	style UnitCardStyle
}

type unitCardWeapon struct {
	weapon   model.Weapon
	quantity int
}

func createUnitMenu(g *Game) *UnitMenu {
	var ui *ebitenui.UI = &ebitenui.UI{}

	menu := &UnitMenu{
		MenuModel: &MenuModel{
			game:   g,
			ui:     ui,
			active: true,
		},
		selectedUnit: nil,
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
	selection := unitMenuSelectionContainer(m)
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

func unitMenuSelectionContainer(m *UnitMenu) widget.PreferredSizeLocateableWidget {
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

			m.selectedUnit = nextPage.unit
		}))

	c.AddChild(pageList)

	c.AddChild(pageContainer.widget)

	pageList.SetSelectedEntry(pages[0])

	return c
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

	variantContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:  0,
				Right: 0,
			}),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{false}),
			widget.GridLayoutOpts.Spacing(m.Spacing(), 0),
		)))
	c.AddChild(variantContainer)

	flipBook := widget.NewFlipBook(
		widget.FlipBookOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		}))),
	)
	c.AddChild(flipBook)

	return &unitPageContainer{
		unitMenu:         m,
		widget:           c,
		variantContainer: variantContainer,
		flipBook:         flipBook,
	}
}

func (p *unitPageContainer) setPage(page *unitPage) {
	p.variantContainer.RemoveChildren()
	m := p.unitMenu
	res := m.Resources()

	// update page unit content to current unit
	page.setUnit(m, page.unit)

	// show unit chassis name
	chassisName := "Random"
	if page.unit != nil {
		chassisName = page.unit.Name()
	}
	chassisText := widget.NewText(widget.TextOpts.Text(chassisName, res.text.face, res.text.idleColor))
	p.variantContainer.AddChild(chassisText)

	// show unit variant selection
	if page.unit != nil {
		comboVariants := []interface{}{}
		for _, v := range page.variants {
			comboVariants = append(comboVariants, v)
		}

		variantCombo := newListComboButton(
			comboVariants,
			page.unit,
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
				m.selectedUnit = u

				// update page content info
				page.setUnit(m, u)
			},
			res)

		p.variantContainer.AddChild(variantCombo)

		if len(comboVariants) <= 1 {
			// only allow variant selection if more than one to choose from
			variantCombo.GetWidget().Disabled = true
		}
	}
	p.variantContainer.RequestRelayout()

	p.flipBook.SetPage(page.content)
	p.flipBook.RequestRelayout()
}

func unitSelectionPage(m *UnitMenu, unit model.Unit, variants []model.Unit) *unitPage {
	// create page stub container, not loading unit data until the page/variant is selected

	var unitName, unitTonnage string
	if unit == nil {
		unitName = "Random"
		unitTonnage = "??"
	} else {
		unitName = unit.Name()
		unitTonnage = fmt.Sprintf("%0.0f", unit.Tonnage())
	}

	page := &unitPage{
		title:    fmt.Sprintf("%s - %s", unitTonnage, unitName),
		content:  newPageContentContainer(),
		unit:     unit,
		variants: variants,
	}
	return page
}

func (p *unitPage) setUnit(m *UnitMenu, unit model.Unit) {
	p.content.RemoveChildren()
	p.unit = unit

	unitCard := createUnitCard(m.game, m.Resources(), unit, UnitCardSelect)
	p.content.AddChild(unitCard)
}

func createUnitCard(g *Game, res *uiResources, unit model.Unit, style UnitCardStyle) *UnitCard {
	cardContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, false, true}),
			widget.GridLayoutOpts.Spacing(0, 0)),
		),
	)
	unitCard := &UnitCard{
		Container: cardContainer,
		style:     style,
	}

	switch style {
	case UnitCardLaunch:
		// also show chassis name and variant header
		chassisVariant := "Random Mech"
		if unit != nil {
			chassisVariant = fmt.Sprintf("%s\n%s", unit.Name(), unit.Variant())
		}
		chassisText := widget.NewText(widget.TextOpts.Text(chassisVariant, res.text.titleFace, res.text.idleColor))
		cardContainer.AddChild(chassisText)
	case UnitCardMission:
		// also show chassis name and variant header
		chassisVariant := fmt.Sprintf("%s / %s", unit.Name(), unit.Variant())
		chassisText := widget.NewText(widget.TextOpts.Text(chassisVariant, res.text.face, res.text.idleColor))
		cardContainer.AddChild(chassisText)
	}

	unitTable := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(g.menu.Spacing()),
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
	)
	cardContainer.AddChild(unitTable)

	// show unit image graphic
	var sprite *render.Sprite
	switch interfaceType := unit.(type) {
	case *model.Mech:
		sprite = g.createUnitSprite(unit).(*render.MechSprite).Sprite
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
		imageH := float64(g.screenHeight) / 5
		spriteW, spriteH := float64(sprite.Texture().Bounds().Dx()), float64(sprite.Texture().Bounds().Dy())
		imageScale := imageH / spriteH

		unitImage := ebiten.NewImage(int(spriteW*imageScale), int(spriteH*imageScale))
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(imageScale, imageScale)
		unitImage.DrawImage(sprite.Texture(), op)

		imageLabel = widget.NewGraphic(
			widget.GraphicOpts.Image(unitImage),
		)
	}
	unitTable.AddChild(imageLabel)

	if unit == nil {
		// no more content to add for random unit select
		return unitCard
	}

	// unit content container
	unitContent := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:  g.menu.Spacing(),
				Right: g.menu.Spacing(),
			}),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{false, false}),
			widget.GridLayoutOpts.Spacing(g.menu.Spacing(), g.menu.Spacing()),
		)))
	unitTable.AddChild(unitContent)

	// unit specifications (tonnage, speed, jumpjets, armor, structure)
	massString := fmt.Sprintf("%0.0f Tons", unit.Tonnage())
	massText := newUnitContentText(res, massString)
	unitContent.AddChild(newUnitContentText(res, "Mass:"))
	unitContent.AddChild(massText)

	topSpeedKph := unit.MaxVelocity() * model.VELOCITY_TO_KPH
	speedString := fmt.Sprintf("%0.1f kph", topSpeedKph)
	speedText := newUnitContentText(res, speedString)
	unitContent.AddChild(newUnitContentText(res, "Top Speed:"))
	unitContent.AddChild(speedText)

	// armament summary container
	armsContent := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(0),
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)),
	)
	cardContainer.AddChild(armsContent)

	armamentText := widget.NewText(
		widget.TextOpts.Text("Armament:", res.text.face, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	)
	armsContent.AddChild(armamentText)

	for _, weapon := range armamentSummary(unit) {
		weaponFull := weapon.weapon.Name()
		weaponShort := weapon.weapon.ShortName()

		// TODO: add more weapon data in tooltip
		toolTipString := fmt.Sprintf("%dx %s", weapon.quantity, weaponFull)
		toolTip := widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(res.toolTip.background),
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
				widget.RowLayoutOpts.Padding(res.toolTip.padding),
				widget.RowLayoutOpts.Spacing(2),
			)))
		toolTipText := widget.NewText(
			widget.TextOpts.Text(toolTipString, res.toolTip.face, res.toolTip.color),
		)
		toolTip.AddChild(toolTipText)

		weaponString := fmt.Sprintf("%dx %s", weapon.quantity, weaponShort)
		weaponText := widget.NewText(
			widget.TextOpts.Text(weaponString, res.text.smallFace, res.text.idleColor),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(widget.WidgetOpts.ToolTip(widget.NewToolTip(
				widget.ToolTipOpts.Content(toolTip),
			))),
		)
		armsContent.AddChild(weaponText)
	}

	// TODO: add more content

	return unitCard
}

func newUnitContentText(res *uiResources, str string) *widget.Text {
	return widget.NewText(
		widget.TextOpts.Text(str, res.text.smallFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	)
}

func armamentSummary(unit model.Unit) []*unitCardWeapon {
	if unit == nil {
		return []*unitCardWeapon{}
	}

	weaponSummaryList := make([]*unitCardWeapon, 0, len(unit.Armament()))
	for _, weapon := range unit.Armament() {
		name := weapon.Name()

		var foundSummary *unitCardWeapon
		for _, summary := range weaponSummaryList {
			if summary.weapon.Name() == name {
				foundSummary = summary
				break
			}
		}

		if foundSummary == nil {
			newSummary := &unitCardWeapon{
				weapon:   weapon,
				quantity: 1,
			}
			weaponSummaryList = append(weaponSummaryList, newSummary)
		} else {
			foundSummary.quantity += 1
		}
	}

	return weaponSummaryList
}
