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
	"github.com/pixelmek-3d/pixelmek-3d/game/texture"

	log "github.com/sirupsen/logrus"
)

type MapMenu struct {
	*MenuModel
	selectedMap *model.Map
}

type mapMenuPageContainer struct {
	mapMenu   *MapMenu
	widget    *widget.Container
	titleText *widget.Text
	flipBook  *widget.FlipBook
}

type mapMenuPage struct {
	title    string
	mapFile  string
	content  *widget.Container
	modelMap *model.Map
}

type MapCardStyle int

const (
	MapCardSelect MapCardStyle = iota
)

type MapCard struct {
	*widget.Container
	style          MapCardStyle
	objectivesText *widget.TextArea
}

var mapImage *mapMapImage

type mapMapImage struct {
	modelMap *model.Map
	image    *ebiten.Image
}

func createMapMenu(g *Game) *MapMenu {
	var ui *ebitenui.UI = &ebitenui.UI{}

	menu := &MapMenu{
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

func (m *MapMenu) initMenu() {
	m.MenuModel.initMenu()
	m.root.SetBackgroundImage(m.Resources().background)

	// menu title
	titleBar := mapTitleContainer(m)
	m.root.AddChild(titleBar)

	// map selection
	selection := mapMenuSelectionContainer(m)
	m.root.AddChild(selection)

	// footer
	footer := mapMenuFooterContainer(m)
	m.root.AddChild(footer)
}

func (m *MapMenu) Update() {
	m.ui.Update()
}

func (m *MapMenu) Draw(screen *ebiten.Image) {
	m.ui.Draw(screen)
}

func mapTitleContainer(m *MapMenu) *widget.Container {
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
		widget.TextOpts.Text("Map Selection", res.text.bigTitleFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	return c
}

func mapMenuFooterContainer(m *MapMenu) *widget.Container {
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
			iScene, _ := game.scene.(MenuScene)
			iScene.back()
		}),
	)
	c.AddChild(back)

	c.AddChild(newBlankSeparator(m.Resources(), m.Padding(), widget.RowLayoutData{
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
			iScene, _ := game.scene.(MenuScene)
			iScene.next()
		}),
	)
	c.AddChild(next)

	return c
}

func mapMenuSelectionContainer(m *MapMenu) widget.PreferredSizeLocateableWidget {
	res := m.Resources()
	g := m.Game()

	mapList, err := model.ListMapFilenames()
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

	pages := make([]any, 0, len(mapList))

	// TODO: add entry for random map

	for _, mapFile := range mapList {
		if !g.debug && strings.HasPrefix(strings.ToLower(mapFile), "debug") {
			// only show debug prefixed maps in debug mode
			continue
		}
		mapPage := mapSelectionPage(m, mapFile)
		pages = append(pages, mapPage)
	}

	pageContainer := newMapMenuPageContainer(m)

	pageList := widget.NewList(
		widget.ListOpts.Entries(pages),
		widget.ListOpts.EntryLabelFunc(func(e any) string {
			return e.(*mapMenuPage).title
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
			nextPage := args.Entry.(*mapMenuPage)
			pageContainer.setPage(nextPage)
			m.Root().RequestRelayout()

			m.selectedMap = nextPage.modelMap
		}))

	c.AddChild(pageList)

	c.AddChild(pageContainer.widget)

	pageList.SetSelectedEntry(pages[0])

	return c
}

func newMapMenuPageContainer(m *MapMenu) *mapMenuPageContainer {
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

	return &mapMenuPageContainer{
		mapMenu:   m,
		widget:    c,
		titleText: titleText,
		flipBook:  flipBook,
	}
}

func (p *mapMenuPageContainer) setPage(page *mapMenuPage) {
	m := p.mapMenu

	// update page map content to current map
	page.setMap(m)

	// show map title
	p.titleText.Label = page.modelMap.Name

	p.flipBook.SetPage(page.content)
	p.flipBook.RequestRelayout()
}

func mapSelectionPage(_ *MapMenu, mapFile string) *mapMenuPage {
	// create page stub container, not loading map data until it is selected
	titleStr := strings.ToTitle(strings.ReplaceAll(strings.TrimSuffix(mapFile, ".yaml"), "_", " "))
	page := &mapMenuPage{
		title:   titleStr,
		mapFile: mapFile,
		content: newPageContentContainer(),
	}
	return page
}

func (p *mapMenuPage) setMap(m *MapMenu) {
	p.content.RemoveChildren()
	if p.modelMap == nil {
		// load map data
		var err error
		p.modelMap, err = model.LoadMap(p.mapFile)
		if err != nil {
			log.Error("Error loading map: ", p.mapFile)
			log.Error(err)
			exit(1)
		}
	}

	mapCard := createMapCard(m.game, m.Resources(), p.modelMap, MapCardSelect)
	p.content.AddChild(mapCard)
}

func createMapCard(g *Game, res *uiResources, modelMap *model.Map, style MapCardStyle) *MapCard {
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

	// switch style {
	// case MapCardLaunch, MapCardGame, MapCardDebrief:
	// 	mapText := widget.NewText(widget.TextOpts.Text(modelMap.Name, res.text.titleFace, res.text.idleColor),
	// 		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
	// 			Stretch: true,
	// 		})),
	// 		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	// 	)
	// 	cardContainer.AddChild(mapText)
	// }

	var objectivesText *widget.TextArea

	switch style {
	case MapCardSelect: //, MapCardLaunch:
		// map map area text
		worldMap := modelMap.Level(0)
		mapWidthKm := float64(len(worldMap)) * model.METERS_PER_UNIT / 1000
		mapHeightKm := float64(len(worldMap[0])) * model.METERS_PER_UNIT / 1000
		mapString := fmt.Sprintf("Area: %0.0fkm x %0.0fkm", mapWidthKm, mapHeightKm)
		mapText := widget.NewText(widget.TextOpts.Text(mapString, res.text.face, res.text.idleColor))
		cardContainer.AddChild(mapText)

		// map map thumbnail
		mapThumb := createMapThumbnail(g, res, modelMap)
		cardContainer.AddChild(mapThumb)

		// case MapCardGame:
		// 	// map map thumbnail
		// 	mapThumb := createMapThumbnail(g, res, modelMap)
		// 	cardContainer.AddChild(mapThumb)

		// 	// in-game map objectives text
		// 	objectivesLabel := widget.NewText(widget.TextOpts.Text("Objectives", res.text.face, res.text.idleColor))
		// 	cardContainer.AddChild(objectivesLabel)

		// 	objectivesText = newTextArea(g.objectives.Text(), res, widget.WidgetOpts.LayoutData(widget.GridLayoutData{
		// 		MaxHeight: g.uiRect().Dy() / 5,
		// 	}))
		// 	cardContainer.AddChild(objectivesText)

		// case MapCardDebrief:
		// 	// post-map objectives text
		// 	objectivesLabel := widget.NewText(widget.TextOpts.Text("Objectives", res.text.face, res.text.idleColor))
		// 	cardContainer.AddChild(objectivesLabel)

		// 	objectivesText = newTextArea(g.objectives.Text(), res, widget.WidgetOpts.LayoutData(widget.GridLayoutData{
		// 		MaxHeight: g.uiRect().Dy() / 5,
		// 	}))
		// 	cardContainer.AddChild(objectivesText)
	}

	mapCard := &MapCard{
		Container:      cardContainer,
		style:          style,
		objectivesText: objectivesText,
	}

	return mapCard
}

func (c *MapCard) update(g *Game) {
	// switch c.style {
	// case MapCardGame:
	// 	c.objectivesText.SetText(g.objectives.Text())
	// }
}

func createMapThumbnail(g *Game, res *uiResources, modelMap *model.Map) *widget.Container {
	mapOpts := mapimage.MapImageOptions{PxPerCell: 2, RenderDefaultFloorTexture: false}
	//mapOpts := mapimage.MapImageOptions{RenderDropZone: true, RenderNavPoints: true}

	// container for map map image and button to show map larger in window
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(g.menu.Spacing()),
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
	)

	// var mapTex *texture.TextureHandler
	// if g.modelMap == modelMap {
	// 	mapTex = g.tex
	// } else {
	mapTex := texture.NewTextureHandler(modelMap)
	// }

	mapImage, err := mapimage.NewMapImage(modelMap, mapTex, mapOpts)
	if err != nil {
		log.Error("Error loading map image: ", err)
	} else if mapImage != nil {
		// scale image down to fit thumbnail space
		mapImage = common.ScaleImageToHeight(mapImage, g.uiRect().Dy()/5, ebiten.FilterNearest)
		imageButton := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			})),
			widget.ButtonOpts.Image(res.button.image),
			widget.ButtonOpts.Graphic(&widget.GraphicImage{
				Idle: mapImage,
			}),
			widget.ButtonOpts.GraphicPadding(widget.Insets{Top: 4, Bottom: 4, Left: 25, Right: 25}),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				// show pop-up with large map
				openMapWindow(g, res, modelMap)
			}),
		)
		c.AddChild(imageButton)
	}

	return c
}

func openMapWindow(g *Game, res *uiResources, modelMap *model.Map) {
	// TODO: refactor shared functionality with menumission.openMissionMapWindow

	mapOpts := mapimage.MapImageOptions{PxPerCell: 8, RenderDefaultFloorTexture: true}
	//mapOpts := mapimage.MapImageOptions{RenderDropZone: true, RenderNavPoints: true}

	var rmWindow widget.RemoveWindowFunc
	var window *widget.Window

	m := g.menu
	uiRect := g.uiRect()
	padding := m.Padding()
	spacing := m.Spacing()

	titleBar := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true}), widget.GridLayoutOpts.Padding(&widget.Insets{
			Left:   padding,
			Right:  padding,
			Top:    padding,
			Bottom: padding,
		}))))

	titleBar.AddChild(widget.NewText(
		widget.TextOpts.Text("Map", res.text.titleFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	titleBar.AddChild(widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("X", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			rmWindow()
		}),
		widget.ButtonOpts.TabOrder(99),
	))

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.image),
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(1),
				widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, false, true}),
				widget.GridLayoutOpts.Padding(res.panel.padding),
				widget.GridLayoutOpts.Spacing(1, spacing),
			),
		),
	)

	if mapImage == nil || mapImage.modelMap != modelMap || mapImage.image == nil {
		var mapTex *texture.TextureHandler
		if g.tex != nil && g.tex.IsHandlerForMap(modelMap) {
			mapTex = g.tex
		} else {
			mapTex = texture.NewTextureHandler(modelMap)
		}
		img, err := mapimage.NewMapImage(modelMap, mapTex, mapOpts)
		if err != nil {
			log.Error("Error loading map image: ", err)
		}
		mapImage = &mapMapImage{
			modelMap: modelMap,
			image:    img,
		}
	}

	if mapImage != nil && mapImage.image != nil {
		// resize map image to fit window
		iWidth, iHeight := mapImage.image.Bounds().Dx(), mapImage.image.Bounds().Dy()

		iScale := (float64(uiRect.Dy()) / 2) / float64(iHeight)
		if int(float64(iWidth)*iScale) > uiRect.Dx()/2 {
			// handle ultrawide maps
			iScale = (float64(uiRect.Dx()) / 2) / float64(iWidth)
		}

		scaledImage := ebiten.NewImage(int(float64(iWidth)*iScale), int(float64(iHeight)*iScale))
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(iScale, iScale)
		scaledImage.DrawImage(mapImage.image, op)

		imageLabel := widget.NewGraphic(
			widget.GraphicOpts.Image(scaledImage),
		)
		c.AddChild(imageLabel)
	}

	// TODO: map navigation/zoom controls

	window = widget.NewWindow(
		widget.WindowOpts.Modal(),
		widget.WindowOpts.Contents(c),
		widget.WindowOpts.TitleBar(titleBar, uiRect.Dy()/12),
	)

	wRect := uiRect.Inset(uiRect.Dy() / 6)
	window.SetLocation(wRect)

	rmWindow = m.UI().AddWindow(window)
	m.SetWindow(window)
}
