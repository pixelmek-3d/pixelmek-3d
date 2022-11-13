package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/tinne26/etxt"
)

type Armament struct {
	HUDSprite
	image        *ebiten.Image
	fontRenderer *etxt.Renderer
}

//NewArmament creates a weapon list image to be rendered on demand
func NewArmament(width, height int, font *Font) *Armament {
	img := ebiten.NewImage(width, height)

	// create and configure renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetSizePx(20)
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.Top, etxt.Left)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	a := &Armament{
		HUDSprite:    NewHUDSprite(img, 1.0),
		image:        img,
		fontRenderer: renderer,
	}

	return a
}

func (a *Armament) Update(armament []model.Weapon) {
	a.image.Clear()

	a.fontRenderer.SetTarget(a.image)

	fontPx := a.fontRenderer.GetSizePxFract().Ceil()

	// TODO: render weapons as individual sub-images within the display
	numWeapons := len(armament)
	for i, weapon := range armament {
		wX, wY := 0, i*(fontPx+fontPx/2)
		if i >= numWeapons/2 {
			wX, wY = a.Width()/2, (i-numWeapons/2)*(fontPx+fontPx/2)
		}

		if weapon.Cooldown() == 0 {
			a.fontRenderer.SetColor(color.RGBA{255, 255, 255, 255})
		} else {
			a.fontRenderer.SetColor(color.RGBA{255, 255, 255, 96})
		}

		a.fontRenderer.Draw(weapon.ShortName(), wX, wY)
	}
}

func (a *Armament) Texture() *ebiten.Image {
	return a.image
}
