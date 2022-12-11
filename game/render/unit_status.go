package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tinne26/etxt"
)

type UnitStatus struct {
	HUDSprite
	image        *ebiten.Image
	fontRenderer *etxt.Renderer
	unit         *Sprite
	unitImage    *ebiten.Image
}

//NewUnitStatus creates a unit status element image to be rendered on demand
func NewUnitStatus(width, height int, font *Font) *UnitStatus {
	img := ebiten.NewImage(width, height)
	unitImg := ebiten.NewImage(width, height)

	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetSizePx(16)
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.YCenter, etxt.Left)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	u := &UnitStatus{
		HUDSprite:    NewHUDSprite(img, 1.0),
		image:        img,
		unitImage:    unitImg,
		fontRenderer: renderer,
	}

	return u
}

func (u *UnitStatus) SetUnit(unit *Sprite) {
	u.unit = unit
	u.unitImage.Clear()

	// create static outline image of unit and store it
	uTexture := unit.Texture()

	op := &ebiten.DrawImageOptions{}
	// Reset RGB (not Alpha) 0 forcibly
	op.ColorM.Scale(0, 0, 0, 1)

	// Set color
	r, g, b := 1.0, 1.0, 1.0
	op.ColorM.Translate(r, g, b, 0)

	_, iH := u.image.Size()
	_, uH := uTexture.Size()
	uScale := (0.9 * float64(iH)) / float64(uH)
	op.GeoM.Scale(uScale, uScale)

	u.unitImage.DrawImage(uTexture, op)
}

func (u *UnitStatus) Update() {
	u.image.Clear()

	u.fontRenderer.SetTarget(u.image)

	// background box
	bW, bH := float64(u.Width()), float64(u.Height())
	ebitenutil.DrawRect(u.image, 0, 0, bW, bH, color.RGBA{255, 255, 255, 48})

	// unit image
	op := &ebiten.DrawImageOptions{}
	u.image.DrawImage(u.unitImage, op)

	// armor readout
	armorPercent := 100 * u.unit.ArmorPoints() / u.unit.MaxArmorPoints()
	armorStr := fmt.Sprintf("ARMOR\n%0.0f%%", armorPercent)
	u.fontRenderer.Draw(armorStr, int(2*bW/3), int(bH/3))

	// internal structure readout
	internalPercent := 100 * u.unit.StructurePoints() / u.unit.MaxStructurePoints()
	internalStr := fmt.Sprintf("STRUCT\n%0.0f%%", internalPercent)
	u.fontRenderer.Draw(internalStr, int(2*bW/3), int(2*bH/3))
}

func (u *UnitStatus) Texture() *ebiten.Image {
	return u.image
}
