package game

import (
	"fmt"
	"strings"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/common"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/mapimage"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/missionimage"
	"github.com/pixelmek-3d/pixelmek-3d/game/texture"

	log "github.com/sirupsen/logrus"
)

type MissionMenu struct {
	*MenuModel
	selectedMission *model.Mission
}

type missionMenuPageContainer struct {
	missionMenu *MissionMenu
	widget      *widget.Container
	titleText   *widget.Text
	flipBook    *widget.FlipBook
}

type missionMenuPage struct {
	title       string
	missionFile string
	content     *widget.Container
	mission     *model.Mission
}

type MissionCardStyle int

const (
	MissionCardSelect MissionCardStyle = iota
	MissionCardLaunch
	MissionCardGame
	MissionCardDebrief
)

type MissionCard struct {
	*widget.Container
	style          MissionCardStyle
	objectivesText *widget.TextArea
}

func createMissionMenu(g *Game) *MissionMenu {
	var ui *ebitenui.UI = &ebitenui.UI{}

	menu := &MissionMenu{
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

func (m *MissionMenu) initMenu() {
	m.MenuModel.initMenu()
	m.root.SetBackgroundImage(m.Resources().background)

	// menu title
	titleBar := missionTitleContainer(m)
	m.root.AddChild(titleBar)

	// mission selection
	selection := missionMenuSelectionContainer(m)
	m.root.AddChild(selection)

	// footer
	footer := missionMenuFooterContainer(m)
	m.root.AddChild(footer)
}

func (m *MissionMenu) Update() {
	m.ui.Update()
}

func (m *MissionMenu) Draw(screen *ebiten.Image) {
	m.ui.Draw(screen)
}

func missionTitleContainer(m *MissionMenu) *widget.Container {
	res := m.Resources()

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Left:   m.Padding(),
				Right:  m.Padding(),
				Top:    m.Padding(),
				Bottom: m.Padding(),
			}))))

	c.AddChild(widget.NewText(
		widget.TextOpts.Text("Mission Selection", res.text.bigTitleFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	return c
}

func missionMenuFooterContainer(m *MissionMenu) *widget.Container {
	game := m.Game()
	res := m.Resources()

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(3),
			widget.GridLayoutOpts.Stretch([]bool{false, true, false}, []bool{false}),
			widget.GridLayoutOpts.Padding(&widget.Insets{
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

func missionMenuSelectionContainer(m *MissionMenu) widget.PreferredSizeLocateableWidget {
	res := m.Resources()
	g := m.Game()

	missionList, err := model.ListMissionFilenames()
	if err != nil {
		log.Error(err)
	}

	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Left:  m.Spacing(),
				Right: m.Spacing(),
			}),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(m.Spacing(), 0),
		)))

	pages := make([]any, 0, len(missionList))

	// TODO: add entry for random mission

	for _, missionFile := range missionList {
		if !g.debug && strings.HasPrefix(strings.ToLower(missionFile), "debug") {
			// only show debug prefixed missions in debug mode
			continue
		}
		missionPage := missionSelectionPage(m, missionFile)
		pages = append(pages, missionPage)
	}

	pageContainer := newMissionMenuPageContainer(m)

	pageList := widget.NewList(
		widget.ListOpts.Entries(pages),
		widget.ListOpts.EntryLabelFunc(func(e any) string {
			return e.(*missionMenuPage).title
		}),
		widget.ListOpts.ScrollContainerImage(res.list.image),
		widget.ListOpts.SliderParams(&widget.SliderParams{
			TrackImage:    res.list.track,
			HandleImage:   res.list.handle,
			MinHandleSize: res.list.handleSize,
			TrackPadding:  res.list.trackPadding,
		},
		),
		widget.ListOpts.EntryColor(res.list.entry),
		widget.ListOpts.EntryFontFace(res.list.face),
		widget.ListOpts.EntryTextPadding(res.list.entryPadding),
		widget.ListOpts.HideHorizontalSlider(),
		widget.ListOpts.EntryTextPosition(widget.TextPositionStart, widget.TextPositionCenter),
		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			nextPage := args.Entry.(*missionMenuPage)
			pageContainer.setPage(nextPage)
			m.Root().RequestRelayout()

			m.selectedMission = nextPage.mission
		}))

	c.AddChild(pageList)

	c.AddChild(pageContainer.widget)

	pageList.SetSelectedEntry(pages[0])

	return c
}

func newMissionMenuPageContainer(m *MissionMenu) *missionMenuPageContainer {
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
		widget.TextOpts.Text("", res.text.titleFace, res.text.idleColor))
	c.AddChild(titleText)

	flipBook := widget.NewFlipBook(
		widget.FlipBookOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		}))),
	)
	c.AddChild(flipBook)

	return &missionMenuPageContainer{
		missionMenu: m,
		widget:      c,
		titleText:   titleText,
		flipBook:    flipBook,
	}
}

func (p *missionMenuPageContainer) setPage(page *missionMenuPage) {
	m := p.missionMenu

	// update page mission content to current mission
	page.setMission(m)

	// show mission title
	p.titleText.Label = page.mission.Title

	p.flipBook.SetPage(page.content)
	p.flipBook.RequestRelayout()
}

func missionSelectionPage(_ *MissionMenu, missionFile string) *missionMenuPage {
	// create page stub container, not loading mission data until it is selected
	titleStr := strings.ToTitle(strings.ReplaceAll(strings.TrimSuffix(missionFile, ".yaml"), "_", " "))
	page := &missionMenuPage{
		title:       titleStr,
		missionFile: missionFile,
		content:     newPageContentContainer(),
	}
	return page
}

func (p *missionMenuPage) setMission(m *MissionMenu) {
	p.content.RemoveChildren()
	if p.mission == nil {
		// load mission data
		var err error
		p.mission, err = model.LoadMission(p.missionFile)
		if err != nil {
			log.Error("Error loading mission: ", p.missionFile)
			log.Error(err)
			exit(1)
		}
	}

	missionCard := createMissionCard(m.game, m.Resources(), p.mission, MissionCardSelect)
	p.content.AddChild(missionCard)
}

func createMissionCard(g *Game, res *uiResources, mission *model.Mission, style MissionCardStyle) *MissionCard {

	cardContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, false, false, false, true}),
			widget.GridLayoutOpts.Spacing(0, 0)),
		),
	)

	switch style {
	case MissionCardLaunch, MissionCardGame, MissionCardDebrief:
		missionText := widget.NewText(widget.TextOpts.Text(mission.Title, res.text.titleFace, res.text.idleColor),
			widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			})),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		)
		cardContainer.AddChild(missionText)
	}

	var objectivesText *widget.TextArea

	switch style {
	case MissionCardSelect, MissionCardLaunch:
		// mission map area text
		worldMap := mission.Map().Level(0)
		mapWidthKm := float64(len(worldMap)) * model.METERS_PER_UNIT / 1000
		mapHeightKm := float64(len(worldMap[0])) * model.METERS_PER_UNIT / 1000
		mapString := fmt.Sprintf("Area: %0.0fkm x %0.0fkm", mapWidthKm, mapHeightKm)
		mapText := widget.NewText(widget.TextOpts.Text(mapString, res.text.face, res.text.idleColor))
		cardContainer.AddChild(mapText)

		// mission map thumbnail
		missionThumb := createMissionThumbnail(g, mission)
		cardContainer.AddChild(missionThumb)

		// mission briefing text
		briefingLabel := widget.NewText(widget.TextOpts.Text("Mission Briefing", res.text.face, res.text.idleColor))
		cardContainer.AddChild(briefingLabel)
		briefingText := newTextArea(mission.Briefing, res, widget.WidgetOpts.LayoutData(widget.GridLayoutData{
			MaxHeight: g.uiRect().Dy() / 5,
		}))
		cardContainer.AddChild(briefingText)

		// mission objectives text
		objectivesLabel := widget.NewText(widget.TextOpts.Text("Objectives", res.text.face, res.text.idleColor))
		cardContainer.AddChild(objectivesLabel)

		objectivesText = newTextArea(mission.Objectives.Text(), res, widget.WidgetOpts.LayoutData(widget.GridLayoutData{
			MaxHeight: g.uiRect().Dy() / 5,
		}))
		cardContainer.AddChild(objectivesText)

	case MissionCardGame:
		// mission map thumbnail
		missionThumb := createMissionThumbnail(g, mission)
		cardContainer.AddChild(missionThumb)

		// in-game mission objectives text
		objectivesLabel := widget.NewText(widget.TextOpts.Text("Objectives", res.text.face, res.text.idleColor))
		cardContainer.AddChild(objectivesLabel)

		objectivesText = newTextArea(g.objectives.Text(), res, widget.WidgetOpts.LayoutData(widget.GridLayoutData{
			MaxHeight: g.uiRect().Dy() / 5,
		}))
		cardContainer.AddChild(objectivesText)
	case MissionCardDebrief:
		// post-mission objectives text
		objectivesLabel := widget.NewText(widget.TextOpts.Text("Objectives", res.text.face, res.text.idleColor))
		cardContainer.AddChild(objectivesLabel)

		objectivesText = newTextArea(g.objectives.Text(), res, widget.WidgetOpts.LayoutData(widget.GridLayoutData{
			MaxHeight: g.uiRect().Dy() / 5,
		}))
		cardContainer.AddChild(objectivesText)
	}

	missionCard := &MissionCard{
		Container:      cardContainer,
		style:          style,
		objectivesText: objectivesText,
	}

	return missionCard
}

func createMissionThumbnail(g *Game, mission *model.Mission) *widget.Container {
	mapOpts := mapimage.MapImageOptions{PxPerCell: 2, RenderDefaultFloorTexture: false}
	missionOpts := missionimage.MissionImageOptions{RenderDropZone: true, RenderNavPoints: true}

	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(g.menu.Spacing()),
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
	)

	var mapTex *texture.TextureHandler
	if g.mission == mission {
		mapTex = g.tex
	} else {
		mapTex = texture.NewTextureHandler(mission.Map())
	}
	missionImage, err := missionimage.NewMissionImage(mission, g.resources, mapTex, mapOpts, missionOpts)
	if err != nil {
		log.Error("Error loading mission image: ", err)
	} else if missionImage != nil {
		// scale image down to fit thumbnail space
		missionImage = common.ScaleImageToHeight(missionImage, g.uiRect().Dy()/5, ebiten.FilterNearest)
		imageLabel := widget.NewGraphic(
			widget.GraphicOpts.Image(missionImage),
		)
		c.AddChild(imageLabel)
	}
	return c
}

func (c *MissionCard) update(g *Game) {
	switch c.style {
	case MissionCardGame:
		c.objectivesText.SetText(g.objectives.Text())
	}
}
