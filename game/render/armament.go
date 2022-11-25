package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/tinne26/etxt"
)

type Armament struct {
	HUDSprite
	image        *ebiten.Image
	fontRenderer *etxt.Renderer
	weapons      []*Weapon
}

type Weapon struct {
	HUDSprite
	image  *ebiten.Image
	weapon model.Weapon
}

//NewArmament creates a weapon list image to be rendered on demand
func NewArmament(width, height int, font *Font) *Armament {
	img := ebiten.NewImage(width, height)

	// create and configure renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetSizePx(20)
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.YCenter, etxt.Left)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	a := &Armament{
		HUDSprite:    NewHUDSprite(img, 1.0),
		image:        img,
		fontRenderer: renderer,
	}

	return a
}

func (a *Armament) SetWeapons(weapons []model.Weapon) {
	a.weapons = make([]*Weapon, len(weapons))

	aWidth, _ := a.image.Size()
	fontPx := a.fontRenderer.GetSizePxFract().Ceil()
	width, height := aWidth/2, fontPx*2
	img := ebiten.NewImage(width, height)

	for i, weapon := range weapons {
		a.weapons[i] = &Weapon{
			HUDSprite: NewHUDSprite(img, 1.0),
			image:     img,
			weapon:    weapon,
		}
	}
}

func (a *Armament) Update() {
	a.image.Clear()

	// render weapons as individual sub-images within the display
	numWeapons := len(a.weapons)
	for i, w := range a.weapons {
		a.updateWeapon(w)

		wWidth, wHeight := w.image.Size()
		var wX, wY float64 = 0, float64(i * wHeight)
		if i >= numWeapons/2 {
			wX, wY = float64(a.Width())/2, float64((i-numWeapons/2)*(wHeight))
		}

		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Translate(wX, wY)

		a.image.DrawImage(w.image, op)

		// --- TESTING WEAPON SELECT BOX ---
		if w.weapon.Cooldown() == 0 {
			// TODO: move to Weapon update and add margins
			// FIXME: when ebitengine v2.5 releases can draw rect outline using StrokeRect
			//        - import "github.com/hajimehoshi/ebiten/v2/vector"
			//        - StrokeRect(dst *ebiten.Image, x, y, width, height float32, strokeWidth float32, clr color.Color)
			var wT float64 = 2 // TODO: calculate line thickness based on image height
			wW, wH := float64(wWidth), float64(wHeight)
			ebitenutil.DrawRect(a.image, wX, wY, wW, wT, color.RGBA{255, 255, 255, 255})
			ebitenutil.DrawRect(a.image, wX+wW-wT, wY, wT, wH, color.RGBA{255, 255, 255, 255})
			ebitenutil.DrawRect(a.image, wX, wY+wH-wT, wW, wT, color.RGBA{255, 255, 255, 255})
			ebitenutil.DrawRect(a.image, wX, wY, wT, wH, color.RGBA{255, 255, 255, 255})
		}
	}
}

func (a *Armament) updateWeapon(w *Weapon) {
	w.image.Clear()

	a.fontRenderer.SetTarget(w.image)

	weapon := w.weapon
	if weapon.Cooldown() == 0 {
		a.fontRenderer.SetColor(color.RGBA{255, 255, 255, 255})
	} else {
		a.fontRenderer.SetColor(color.RGBA{255, 255, 255, 96})
	}

	_, wHeight := w.image.Size()
	wX, wY := 3, wHeight/2 // TODO: calculate better margin spacing

	a.fontRenderer.Draw(weapon.ShortName(), wX, wY)
}

func (a *Armament) Texture() *ebiten.Image {
	return a.image
}
