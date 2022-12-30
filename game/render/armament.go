package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
)

type Armament struct {
	HUDSprite
	fontRenderer *etxt.Renderer
	weapons      []*Weapon
}

type Weapon struct {
	HUDSprite
	weapon model.Weapon
}

//NewArmament creates a weapon list image to be rendered on demand
func NewArmament(font *Font) *Armament {
	// create and configure renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.YCenter, etxt.Left)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

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

func (a *Armament) updateFontSize(width, height int) {
	// set font size based on individual weapon element size
	pxSize := float64(height) / 2
	if pxSize < 1 {
		pxSize = 1
	}

	fractSize, _ := efixed.FromFloat64(pxSize)
	a.fontRenderer.SetSizePxFract(fractSize)
}

func (a *Armament) Draw(screen *ebiten.Image, bounds image.Rectangle, clr *color.RGBA) {
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
		var wX, wY float64 = float64(bX), float64(bY) + float64(i*wHeight)
		if i >= numWeapons/2 {
			wX, wY = float64(bX)+float64(bW)/2, float64(bY)+float64((i-numWeapons/2)*(wHeight))
		}

		wBounds := image.Rect(
			int(wX), int(wY), int(wX)+wWidth, int(wY)+wHeight,
		)
		a.drawWeapon(screen, wBounds, clr, w)

		// --- TESTING WEAPON SELECT BOX ---
		if w.weapon.Cooldown() == 0 {
			// TODO: move to Weapon update and add margins
			// FIXME: when ebitengine v2.5 releases can draw rect outline using StrokeRect
			//        - import "github.com/hajimehoshi/ebiten/v2/vector"
			//        - StrokeRect(dst *ebiten.Image, x, y, width, height float32, strokeWidth float32, clr color.Color)
			var wT float64 = 2 // TODO: calculate line thickness based on image height
			wW, wH := float64(wWidth), float64(wHeight)
			ebitenutil.DrawRect(screen, wX, wY, wW, wT, clr)
			ebitenutil.DrawRect(screen, wX+wW-wT, wY, wT, wH, clr)
			ebitenutil.DrawRect(screen, wX, wY+wH-wT, wW, wT, clr)
			ebitenutil.DrawRect(screen, wX, wY, wT, wH, clr)
		}
	}
}

func (a *Armament) drawWeapon(screen *ebiten.Image, bounds image.Rectangle, clr *color.RGBA, w *Weapon) {
	a.fontRenderer.SetTarget(screen)
	a.fontRenderer.SetColor(clr)

	bX, bY, bH := bounds.Min.X, bounds.Min.Y, bounds.Dy()

	weapon := w.weapon
	if weapon.Cooldown() == 0 {
		a.fontRenderer.SetColor(clr)
	} else {
		wAlpha := uint8(2 * (int(clr.A) / 5))
		a.fontRenderer.SetColor(color.RGBA{clr.R, clr.G, clr.B, wAlpha})
	}

	wX, wY := bX+3, bY+bH/2 // TODO: calculate better margin spacing

	a.fontRenderer.Draw(weapon.ShortName(), wX, wY)
}
