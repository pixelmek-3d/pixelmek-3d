package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"
	log "github.com/sirupsen/logrus"
)

var _weaponGroupsMenuConfig *weaponGroupsMenuConfig

type weaponGroupsMenuConfig struct {
	wg    [][]model.Weapon
	audio *AudioHandler
}

func openWeaponGroupsWindow(g *Game, res *uiResources) {
	var window *widget.Window
	var rmWindow widget.RemoveWindowFunc
	_weaponGroupsMenuConfig = &weaponGroupsMenuConfig{
		wg:    make([][]model.Weapon, len(g.player.weaponGroups)),
		audio: g.audio,
	}
	copy(_weaponGroupsMenuConfig.wg, g.player.weaponGroups)

	m := g.menu
	uiRect := g.uiRect()
	padding := m.Padding()
	spacing := m.Spacing()

	titleBar := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1), widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}), widget.GridLayoutOpts.Padding(&widget.Insets{
			Left:   padding,
			Right:  padding,
			Top:    padding,
			Bottom: padding,
		}))))

	titleBar.AddChild(widget.NewText(
		widget.TextOpts.Text("Weapon Groups", res.text.titleFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.image),
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(1),
				widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true, false}),
				widget.GridLayoutOpts.Padding(res.panel.padding),
				widget.GridLayoutOpts.Spacing(1, spacing),
			),
		),
	)

	groupsContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(2),
				widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{true}),
				widget.GridLayoutOpts.Spacing(4, 0),
			),
		),
	)
	c.AddChild(groupsContainer)

	leftContainer := newPageContentContainer()
	rightContainer := newPageContentContainer()
	groupsContainer.AddChild(leftContainer)
	groupsContainer.AddChild(rightContainer)

	// add weapons to the left or right container based on index like the HUD
	weapons := g.player.Armament()
	for i, w := range weapons {
		weaponSelector := createWeaponGroupsSelector(res, w)
		if i%2 == 0 {
			leftContainer.AddChild(weaponSelector)
		} else {
			rightContainer.AddChild(weaponSelector)
		}
	}

	footerContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.image),
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(3),
				widget.GridLayoutOpts.Stretch([]bool{false, true, false}, []bool{true}),
				widget.GridLayoutOpts.Padding(res.panel.padding),
				widget.GridLayoutOpts.Spacing(1, spacing),
			),
		),
	)
	c.AddChild(footerContainer)

	cancelButton := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("Cancel", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			rmWindow()
		}),
	)
	footerContainer.AddChild(cancelButton)

	footerContainer.AddChild(newBlankSeparator(res, padding, widget.RowLayoutData{
		Stretch: true,
	}))

	acceptButton := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("Accept", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			// save weapon groups
			g.player.weaponGroups = _weaponGroupsMenuConfig.wg
			setUnitWeaponGroups(g.player, g.player.weaponGroups)
			if err := saveUserWeaponGroups(); err != nil {
				log.Error("failed to save user weapon groups: " + err.Error())
			}
			rmWindow()
		}),
	)
	footerContainer.AddChild(acceptButton)

	window = widget.NewWindow(
		widget.WindowOpts.Modal(),
		widget.WindowOpts.Contents(c),
		widget.WindowOpts.TitleBar(titleBar, uiRect.Dy()/12),
	)

	wRect := uiRect.Inset(padding)
	window.SetLocation(wRect)

	rmWindow = m.UI().AddWindow(window)
	m.SetWindow(window)
}

func createWeaponGroupsSelector(res *uiResources, w model.Weapon) *widget.Container {
	border, _ := loadImageNineSlice("menu/titlebar-idle.png", 10, 10, 0.5)

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(border),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(4),
				widget.GridLayoutOpts.Stretch([]bool{false, false, true, false}, []bool{false}),
				widget.GridLayoutOpts.Spacing(1, 1),
				widget.GridLayoutOpts.Padding(&widget.Insets{
					Top:    2,
					Bottom: 2,
					Left:   4,
					Right:  4,
				}),
			),
		),
	)

	wLocation := widget.NewText(
		widget.TextOpts.Text(strings.ToUpper(w.Location().ShortName()+` `), res.text.smallFace, res.text.disabledColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	)
	c.AddChild(wLocation)

	toolTip := createWeaponToolTip(res, w)
	wLabel := widget.NewText(
		widget.TextOpts.Text(w.ShortName(), res.text.face, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.ToolTip(widget.NewToolTip(
			widget.ToolTipOpts.Content(toolTip),
		))),
	)
	c.AddChild(wLabel)

	c.AddChild(newBlankSeparator(res, 1, widget.RowLayoutData{
		Stretch: true,
	}))

	groupsContainer := createWeaponGroupButtons(res, w)
	c.AddChild(groupsContainer)

	return c
}

func createWeaponToolTip(res *uiResources, w model.Weapon) *widget.Container {
	weaponFull := w.Name()
	weaponStats := w.Summary()

	toolTipString := fmt.Sprintf("%s\n\n%s", weaponFull, weaponStats)
	toolTip := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.toolTip.background),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(res.toolTip.padding),
			widget.RowLayoutOpts.Spacing(2),
		)))
	toolTipText := widget.NewText(
		widget.TextOpts.Text(toolTipString, res.toolTip.monoFace, res.toolTip.color),
	)
	toolTip.AddChild(toolTipText)
	return toolTip
}

func createWeaponGroupButtons(res *uiResources, w model.Weapon) *widget.Container {
	c := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(5),
				widget.GridLayoutOpts.Stretch([]bool{false, false, false, false, false}, []bool{false}),
				widget.GridLayoutOpts.Spacing(4, 4),
			),
		),
	)

	for g := model.WEAPON_GROUP_1; g <= model.WEAPON_GROUP_5; g++ {
		groupColor := render.ColorWeaponGroupAll[g]
		textColor := &widget.ButtonTextColor{
			Idle:     alphaColor(groupColor, 125),
			Disabled: alphaColor(groupColor, 75),
			Hover:    alphaColor(groupColor, 200),
			Pressed:  groupColor,
		}
		wgButton := widget.NewButton(
			widget.ButtonOpts.Image(res.miniButton.image),
			widget.ButtonOpts.TextPadding(res.miniButton.padding),
			widget.ButtonOpts.Text(strconv.Itoa(int(g)), res.miniButton.face, textColor),
			widget.ButtonOpts.ToggleMode(),
			widget.ButtonOpts.KeepPressedOnExit(),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				if args.Button.State() == widget.WidgetChecked {
					_weaponGroupsMenuConfig.wg = model.AddWeaponToGroup(w, g, _weaponGroupsMenuConfig.wg)
				} else {
					_weaponGroupsMenuConfig.wg = model.RemoveWeaponFromGroup(w, g, _weaponGroupsMenuConfig.wg)
				}
				if _weaponGroupsMenuConfig.audio != nil {
					go _weaponGroupsMenuConfig.audio.PlayButtonAudio(AUDIO_BUTTON_OVER)
				}
			}),
		)
		if model.IsWeaponInGroup(w, g, _weaponGroupsMenuConfig.wg) {
			wgButton.SetState(widget.WidgetChecked)
		}
		c.AddChild(wgButton)
	}
	return c
}
