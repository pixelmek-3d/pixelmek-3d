package render

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
	"golang.org/x/image/math/fixed"
)

var (
	// define default colors
	_colorWeaponGroup1 = _colorDefaultGreen
	_colorWeaponGroup2 = color.NRGBA{R: 240, G: 240, B: 240, A: 255}
	_colorWeaponGroup3 = color.NRGBA{R: 255, G: 206, B: 0, A: 255}
	_colorWeaponGroup4 = color.NRGBA{R: 145, G: 60, B: 200, A: 255}
	_colorWeaponGroup5 = color.NRGBA{R: 0, G: 200, B: 200, A: 255}
)

type Armament struct {
	HUDSprite
	fontRenderer    *etxt.Renderer
	fontSizeWeapons fixed.Int26_6
	fontSizeGroups  fixed.Int26_6
	weapons         []*Weapon
	weaponGroups    [][]model.Weapon
	selectedWeapon  uint
	selectedGroup   uint
	fireMode        model.WeaponFireMode
}

type Weapon struct {
	HUDSprite
	weapon      model.Weapon
	weaponColor color.NRGBA
}

// NewArmament creates a weapon list image to be rendered on demand
func NewArmament(font *Font) *Armament {
	// create and configure renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)

	a := &Armament{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
	}

	return a
}

func (a *Armament) SetWeapons(weapons []model.Weapon) {
	a.weapons = make([]*Weapon, len(weapons))

	for i, weapon := range weapons {
		a.weapons[i] = &Weapon{
			HUDSprite: NewHUDSprite(nil, 1.0),
			weapon:    weapon,
		}
	}
}

func (a *Armament) SetWeaponGroups(weaponGroups [][]model.Weapon) {
	a.weaponGroups = weaponGroups

	// set default group colors on weapon displays
	for _, w := range a.weapons {
		groups := model.GetGroupsForWeapon(w.weapon, weaponGroups)
		if len(groups) == 0 {
			w.weaponColor = _colorWeaponGroup1
			continue
		}

		switch groups[0] {
		case 0:
			w.weaponColor = _colorWeaponGroup1
		case 1:
			w.weaponColor = _colorWeaponGroup2
		case 2:
			w.weaponColor = _colorWeaponGroup3
		case 3:
			w.weaponColor = _colorWeaponGroup4
		case 4:
			w.weaponColor = _colorWeaponGroup5
		}
	}
}

func (a *Armament) SetSelectedWeapon(weaponOrGroupIndex uint, weaponFireMode model.WeaponFireMode) {
	a.fireMode = weaponFireMode
	switch weaponFireMode {
	case model.CHAIN_FIRE:
		a.selectedWeapon = weaponOrGroupIndex
	case model.GROUP_FIRE:
		a.selectedGroup = weaponOrGroupIndex
	}
}

func (a *Armament) updateFontSize(width, height int) {
	// set font size based on individual weapon element size
	pxSize := float64(height) / 2
	if pxSize < 1 {
		pxSize = 1
	}

	fractSize, _ := efixed.FromFloat64(pxSize)
	a.fontSizeWeapons = fractSize
	a.fontSizeGroups = fractSize / 2
}

func (a *Armament) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen
	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()

	// individual weapon size based on number of weapons and size of armament area
	numWeapons := len(a.weapons)
	numForSizing := 10
	if numWeapons > numForSizing {
		// reduce sizing if weapon count gets overly high
		numForSizing = numWeapons
		if numForSizing%2 != 0 {
			numForSizing++
		}
	}

	wWidth, wHeight := bW/2, int(float64(bH)/float64(numForSizing/2))
	a.updateFontSize(wWidth, wHeight)

	// render weapons as individual sub-images within the display
	for i, w := range a.weapons {
		var wX, wY float64 = float64(bX), float64(bY) + float64((i/2)*wHeight)
		if i%2 != 0 {
			wX, wY = float64(bX)+float64(bW)/2, float64(bY)+float64((i/2)*wHeight)
		}

		wBounds := image.Rect(
			int(wX), int(wY), int(wX)+wWidth, int(wY)+wHeight,
		)

		a.drawWeapon(w, wBounds, hudOpts)

		// render weapon select box
		isWeaponSelected := (a.fireMode == model.CHAIN_FIRE && i == int(a.selectedWeapon)) ||
			(a.fireMode == model.GROUP_FIRE && model.IsWeaponInGroup(w.weapon, a.selectedGroup, a.weaponGroups))

		if isWeaponSelected {
			wColor := hudOpts.HudColor(w.weaponColor)

			if w.weapon.Cooldown() > 0 {
				wAlpha := uint8(2 * (int(wColor.A) / 5))
				wColor = color.NRGBA{wColor.R, wColor.G, wColor.B, wAlpha}
			}

			// TODO: move to Weapon update and add margins
			var wT float32 = 2 // TODO: calculate line thickness based on image height
			wW, wH := float32(wWidth), float32(wHeight)
			vector.StrokeRect(screen, float32(wX), float32(wY), wW, wH, wT, wColor, false)
		}
	}
}

func (a *Armament) drawWeapon(w *Weapon, bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen
	a.fontRenderer.SetTarget(screen)
	a.fontRenderer.SetAlign(etxt.YCenter, etxt.Left)
	a.fontRenderer.SetSizePxFract(a.fontSizeWeapons)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()

	wColor := hudOpts.HudColor(w.weaponColor)

	// render weapon name and status indicator
	weapon := w.weapon
	if weapon.Cooldown() > 0 {
		wColor.A = uint8(2 * (int(wColor.A) / 5))
	}
	a.fontRenderer.SetColor(color.RGBA(wColor))

	wX, wY := bX+3, bY+bH/2 // TODO: calculate better margin spacing

	a.fontRenderer.Draw(weapon.ShortName(), wX, wY)

	// render weapon group indicator
	if len(a.weaponGroups) > 0 {
		a.fontRenderer.SetAlign(etxt.Top, etxt.Right)
		a.fontRenderer.SetSizePxFract(a.fontSizeGroups)

		var groupsTxt string
		for _, g := range model.GetGroupsForWeapon(w.weapon, a.weaponGroups) {
			groupsTxt += fmt.Sprintf("%d ", g+1)
		}

		if len(groupsTxt) > 0 {
			a.fontRenderer.Draw(groupsTxt, bX+bW, bY+2) // TODO: calculate better margin spacing
		}
	}
}
