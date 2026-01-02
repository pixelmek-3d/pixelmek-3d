package render

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/fonts"
	"github.com/tinne26/etxt"
)

var (
	// define default colors
	_colorWeaponGroup1   = _colorDefaultGreen
	_colorWeaponGroup2   = color.NRGBA{R: 240, G: 240, B: 240, A: 255}
	_colorWeaponGroup3   = color.NRGBA{R: 255, G: 206, B: 0, A: 255}
	_colorWeaponGroup4   = color.NRGBA{R: 145, G: 60, B: 200, A: 255}
	_colorWeaponGroup5   = color.NRGBA{R: 0, G: 200, B: 200, A: 255}
	_colorWeaponGroupAll = []color.NRGBA{
		_colorWeaponGroup1,
		_colorWeaponGroup2,
		_colorWeaponGroup3,
		_colorWeaponGroup4,
		_colorWeaponGroup5,
	}
)

type Armament struct {
	HUDSprite
	fontRenderer    *etxt.Renderer
	fontSizeWeapons float64
	fontSizeAmmo    float64
	fontSizeGroups  float64
	weapons         []*Weapon
	weaponGroups    [][]model.Weapon
	selectedWeapon  uint
	selectedGroup   uint
	fireMode        model.WeaponFireMode
	debug           bool
}

type Weapon struct {
	HUDSprite
	weapon      model.Weapon
	weaponColor color.NRGBA
}

// NewArmament creates a weapon list image to be rendered on demand
func NewArmament(font *fonts.Font) *Armament {
	// create and configure renderer
	renderer := etxt.NewRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)

	a := &Armament{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
	}

	return a
}

func (a *Armament) SetWeapons(weapons []model.Weapon) {
	a.debug = false
	a.weapons = make([]*Weapon, len(weapons))

	for i, weapon := range weapons {
		a.weapons[i] = &Weapon{
			HUDSprite: NewHUDSprite(nil, 1.0),
			weapon:    weapon,
		}
	}
}

func (a *Armament) SetDebugWeapons(weapons []model.Weapon) {
	a.debug = true
	a.weapons = make([]*Weapon, len(weapons))

	for i, weapon := range weapons {
		a.weapons[i] = &Weapon{
			HUDSprite: NewHUDSprite(nil, 1.0),
			weapon:    weapon,
		}
	}
}

func (a *Armament) IsDebugWeapons() bool {
	return a.debug
}

func (a *Armament) SetWeaponGroups(weaponGroups [][]model.Weapon) {
	a.weaponGroups = weaponGroups
	a.updateWeaponGroupColors()
}

func (a *Armament) SetWeaponFireMode(weaponFireMode model.WeaponFireMode) {
	a.fireMode = weaponFireMode
}

func (a *Armament) SetSelectedWeapon(weaponIndex uint) {
	a.selectedWeapon = weaponIndex
}

func (a *Armament) SetSelectedWeaponGroup(weaponGroup uint) {
	a.selectedGroup = weaponGroup
	a.updateWeaponGroupColors()
}

func (a *Armament) updateWeaponGroupColors() {
	// set appropriate weapon group colors on weapon displays
	for _, w := range a.weapons {
		groups := model.GetGroupsForWeapon(w.weapon, a.weaponGroups)

		switch len(groups) {
		case 0:
			w.weaponColor = _colorWeaponGroup1
		case 1:
			w.weaponColor = _colorWeaponGroupAll[groups[0]]
		default:
			// more than one group found on weapon, dynamically determine weapon color based on group
			if model.IsWeaponInGroup(w.weapon, a.selectedGroup, a.weaponGroups) {
				// set color to the selected weapon group color
				w.weaponColor = _colorWeaponGroupAll[a.selectedGroup]
			} else {
				// set color to first group the weapon is in
				w.weaponColor = _colorWeaponGroupAll[groups[0]]
			}
		}
	}
}

func (a *Armament) updateFontSize(_, height int) {
	// set font size based on individual weapon element size
	pxSize := float64(height) / 2
	if pxSize < 1 {
		pxSize = 1
	}

	a.fontSizeWeapons = geom.Clamp(pxSize, 1, math.MaxInt)
	a.fontSizeAmmo = geom.Clamp(3*pxSize/5, 1, math.MaxInt)
	a.fontSizeGroups = geom.Clamp(pxSize/2, 1, math.MaxInt)
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
		isWeaponInSameGroup := model.IsWeaponInGroup(w.weapon, a.selectedGroup, a.weaponGroups)
		isWeaponSelected := (a.fireMode == model.CHAIN_FIRE && i == int(a.selectedWeapon)) ||
			(a.fireMode == model.GROUP_FIRE && isWeaponInSameGroup)

		if isWeaponSelected || isWeaponInSameGroup {
			wColor := hudOpts.HudColor(w.weaponColor)
			if w.weapon.Cooldown() > 0 {
				wAlpha := uint8(2 * int(wColor.A) / 5)
				wColor = color.NRGBA{wColor.R, wColor.G, wColor.B, wAlpha}
			}

			var wT float32 = 3 // TODO: calculate line thickness based on image height

			if !isWeaponSelected {
				// reduce selection box thickness and opacity if weapon not selected but in same group
				wAlpha := uint8(2 * int(wColor.A) / 5)
				wColor = color.NRGBA{wColor.R, wColor.G, wColor.B, wAlpha}
				wT = float32(geom.Clamp(float64(wT)/3, 1, float64(wT)))
			}

			wW, wH := float32(wWidth), float32(wHeight)
			vector.StrokeRect(screen, float32(wX), float32(wY), wW, wH, wT, wColor, false)
		}
	}
}

func (a *Armament) drawWeapon(w *Weapon, bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen
	a.fontRenderer.SetAlign(etxt.VertCenter | etxt.Left)
	a.fontRenderer.SetSize(a.fontSizeWeapons)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()

	wColor := hudOpts.HudColor(w.weaponColor)

	// render weapon name and status indicator
	weapon := w.weapon
	wAmmoBin := weapon.AmmoBin()
	isAmmoEmpty := wAmmoBin != nil && wAmmoBin.AmmoCount() == 0

	if weapon.Cooldown() > 0 || isAmmoEmpty {
		wColor.A = uint8(2 * (int(wColor.A) / 5))
	}
	a.fontRenderer.SetColor(color.NRGBA(wColor))

	wX, wY := bX+3, bY+bH/2 // TODO: calculate better margin spacing

	weaponDisplayTxt := weapon.ShortName()
	a.fontRenderer.Draw(screen, weaponDisplayTxt, wX, wY)

	// render ammo indicator
	if wAmmoBin != nil {
		a.fontRenderer.SetAlign(etxt.Bottom | etxt.Right)
		a.fontRenderer.SetSize(a.fontSizeAmmo)

		// just picked a character to indicate as empty
		ammoDisplayTxt := "/ "
		if !isAmmoEmpty {
			ammoDisplayTxt = fmt.Sprintf("_%d ", wAmmoBin.AmmoCount())
		}
		a.fontRenderer.Draw(screen, ammoDisplayTxt, bX+bW, bY+bH-2) // TODO: calculate better margin spacing
	}

	// render weapon group indicator
	if len(a.weaponGroups) > 0 {
		a.fontRenderer.SetAlign(etxt.Top | etxt.Right)
		a.fontRenderer.SetSize(a.fontSizeGroups)
		fontSpacing := int(a.fontSizeGroups)

		weaponGroups := model.GetGroupsForWeapon(w.weapon, a.weaponGroups)
		numWeaponGroups := len(weaponGroups)
		for i, g := range weaponGroups {
			groupTxt := fmt.Sprintf("%d", g+1)

			// set each group number color corresponding to that weapon group color
			gColor := hudOpts.HudColor(_colorWeaponGroupAll[g])
			a.fontRenderer.SetColor(color.NRGBA(gColor))

			a.fontRenderer.Draw(screen, groupTxt, bX+bW-((numWeaponGroups-i)*fontSpacing), bY+2) // TODO: calculate better margin spacing
		}
	}
}
